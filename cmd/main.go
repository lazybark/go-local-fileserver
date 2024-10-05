package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/skip2/go-qrcode"
)

var (
	dir         string
	port        int
	user        string
	pass        string
	logRequests bool
	tmpl        *template.Template
	tmpl404     *template.Template
	tmpl500     *template.Template
)

const (
	Red   = "\033[31m"
	Green = "\033[32m"
	Reset = "\033[0m"
)

func init() {
	flag.StringVar(&dir, "dir", "./public", "Directory to serve files from")
	flag.IntVar(&port, "port", 8080, "Port to run the server on")
	flag.StringVar(&user, "user", "", "Username for basic authentication")
	flag.StringVar(&pass, "pass", "", "Password for basic authentication")
	flag.BoolVar(&logRequests, "log-requests", false, "Log HTTP requests")
}

type FileInfo struct {
	Name         string
	Link         string
	IsImage      bool
	IsDir        bool
	CreatedTime  string
	ModifiedTime string
	Size         string
	Icon         string
}

type Breadcrumb struct {
	Name string
	Link string
}

type DirectoryListing struct {
	Path        string
	Files       []FileInfo
	Breadcrumbs []Breadcrumb
}

func generateThumbnail(srcPath, dstPath string) error {
	img, err := imaging.Open(srcPath)
	if err != nil {
		return err
	}

	// Open the image file to read EXIF data.
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode EXIF data.
	exifData, err := exif.Decode(file)
	if err == nil {
		// Get orientation tag.
		orientation, err := exifData.Get(exif.Orientation)
		if err == nil {
			orientValue, _ := orientation.Int(0)
			switch orientValue {
			case 3:
				img = imaging.Rotate180(img)
			case 6:
				img = imaging.Rotate270(img)
			case 8:
				img = imaging.Rotate90(img)
			}
		}
	}

	thumbnail := imaging.Thumbnail(img, 100, 100, imaging.Lanczos)

	err = os.MkdirAll(filepath.Dir(dstPath), 0755)
	if err != nil {
		return err
	}

	return imaging.Save(thumbnail, dstPath)
}

func generateBreadcrumbs(path string) []Breadcrumb {
	var breadcrumbs []Breadcrumb

	// Add the "Home" breadcrumb.
	breadcrumbs = append(breadcrumbs, Breadcrumb{
		Name: "Home",
		Link: "/",
	})

	parts := strings.Split(path, "/")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		link := strings.Join(parts[:i+1], "/")
		breadcrumbs = append(breadcrumbs, Breadcrumb{
			Name: parts[i],
			Link: link,
		})
	}
	return breadcrumbs
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}

	return "", fmt.Errorf("no IP address found")
}

func getFileIcon(ext string) string {
	switch ext {
	case ".mp3":
		return "/icons/mp3.png"
	case ".txt", ".rtf":
		return "/icons/txt.png"
	case ".pdf":
		return "/icons/pdf.png"
	case ".xls":
		return "/icons/xls.png"
	case ".zip":
		return "/icons/zip.png"
	case ".mp4", ".avi", ".mov", ".mkv":
		return "/icons/video.png"
	case ".doc":
		return "/icons/doc.png"
	default:
		return "/icons/file.png"
	}
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func formatTimestamp(t time.Time) string {
	return t.Format("02 Jan 2006 15:04")
}

func main() {
	flag.Parse()

	// Load the template files
	var err error
	tmpl, err = template.ParseFiles("template.html")
	if err != nil {
		log.Fatalf("%sError loading template.html: %s%s", Red, err, Reset)
	}

	tmpl404, err = template.ParseFiles("404.html")
	if err != nil {
		log.Fatalf("%sError loading 404.html: %s%s", Red, err, Reset)
	}

	tmpl500, err = template.ParseFiles("500.html")
	if err != nil {
		log.Fatalf("%sError loading 500.html: %s%s", Red, err, Reset)
	}

	// Ensure the thumbnails directory is deleted on exit.
	thumbnailDir := "thumbnails"
	cleanup := func() {
		os.RemoveAll(thumbnailDir)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cleanup()
		os.Exit(0)
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Clean the path to prevent path traversal attacks.
		cleanPath := filepath.Clean(r.URL.Path)
		fullPath := filepath.Join(dir, cleanPath)
		if logRequests {
			log.Printf("Requested: %s, Full path: %s, Client: %s", r.URL.Path, fullPath, r.RemoteAddr)
		}

		// Check if the requested path is a hidden file or directory.
		if strings.Contains(r.URL.Path, "..") || strings.Contains(r.URL.Path, "/.") {
			log.Printf("%sPath traversal or hidden file detected: %s%s", Red, r.URL.Path, Reset)
			w.WriteHeader(http.StatusNotFound)
			tmpl404.Execute(w, nil)
			return
		}

		// Ensure the path is within the base directory.
		if !strings.HasPrefix(fullPath, dir) {
			log.Printf("%sPath is outside the base directory: %s%s", Red, fullPath, Reset)
			w.WriteHeader(http.StatusNotFound)
			tmpl404.Execute(w, nil)
			return
		}

		// Check if the path is a file or a directory.
		info, err := os.Stat(fullPath)
		if err != nil {
			log.Printf("%sError stating path: %s%s", Red, err, Reset)
			w.WriteHeader(http.StatusNotFound)
			tmpl404.Execute(w, nil)
			return
		}

		if info.IsDir() {
			// Read the directory contents.
			files, err := os.ReadDir(fullPath)
			if err != nil {
				log.Printf("%sError reading directory: %s%s", Red, err, Reset)
				w.WriteHeader(http.StatusNotFound)
				tmpl404.Execute(w, nil)
				return
			}

			var fileInfos []FileInfo
			for _, file := range files {
				// Skip hidden files and directories
				if strings.HasPrefix(file.Name(), ".") {
					continue
				}

				info, err := file.Info()
				if err != nil {
					log.Printf("%sError getting file info: %s%s", Red, err, Reset)
					w.WriteHeader(http.StatusInternalServerError)
					tmpl500.Execute(w, nil)
					return
				}

				fileInfo := FileInfo{
					Name:         file.Name(),
					Link:         filepath.Join("/", cleanPath, file.Name()), // Absolute path from root without URL encoding
					IsDir:        file.IsDir(),
					CreatedTime:  formatTimestamp(info.ModTime()), // Use the formatted creation time
					ModifiedTime: formatTimestamp(info.ModTime()), // Use the formatted modification time
					Size:         formatFileSize(info.Size()),     // Use the formatted size
				}

				if !file.IsDir() {
					ext := strings.ToLower(filepath.Ext(file.Name()))
					fileInfo.IsImage = ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
					fileInfo.Icon = getFileIcon(ext)
				}

				fileInfos = append(fileInfos, fileInfo)
			}

			// Separate directories and files.
			var dirs, regularFiles []FileInfo

			for _, fileInfo := range fileInfos {
				if fileInfo.IsDir {
					dirs = append(dirs, fileInfo)
				} else {
					regularFiles = append(regularFiles, fileInfo)
				}
			}

			// Combine directories and files, with directories on top.
			fileInfos = append(dirs, regularFiles...)

			// Generate breadcrumbs.
			breadcrumbs := generateBreadcrumbs(cleanPath)

			// Render the template.
			err = tmpl.Execute(w, DirectoryListing{
				Path:        cleanPath,
				Files:       fileInfos,
				Breadcrumbs: breadcrumbs,
			})
			if err != nil {
				log.Printf("%sError rendering template: %s%s", Red, err, Reset)
				w.WriteHeader(http.StatusInternalServerError)
				tmpl500.Execute(w, nil)
			}
		} else {
			// Serve the file directly.
			http.ServeFile(w, r, fullPath)
		}
	})

	http.HandleFunc("/thumbnail/", func(w http.ResponseWriter, r *http.Request) {
		// Clean the path to prevent path traversal attacks.
		cleanPath := filepath.Clean(r.URL.Path[len("/thumbnail/"):])
		fullPath := filepath.Join(dir, cleanPath)

		// Ensure the path is within the base directory.
		if !strings.HasPrefix(fullPath, dir) || strings.Contains(cleanPath, "/.") {
			w.WriteHeader(http.StatusNotFound)
			tmpl404.Execute(w, nil)
			return
		}

		// Generate and serve the thumbnail.
		thumbnailPath := filepath.Join(thumbnailDir, cleanPath)
		if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
			err = generateThumbnail(fullPath, thumbnailPath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				tmpl500.Execute(w, nil)
				return
			}
		}

		http.ServeFile(w, r, thumbnailPath)
	})

	// Serve static files from the /icons/ directory.
	http.Handle("/icons/", http.StripPrefix("/icons/", http.FileServer(http.Dir("./icons"))))

	// Get the local IP address.
	ip, err := getLocalIP()
	if err != nil {
		log.Fatalf("%sError getting local IP address: %s%s", Red, err, Reset)
	}

	// Log the IP address and the URL to access the file list.
	url := fmt.Sprintf("http://%s:%d/", ip, port)
	log.Printf("%sServer is running at %s%s", Green, url, Reset)

	// Generate and print the QR code.
	qrCode, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		log.Fatalf("%sError generating QR code: %s%s", Red, err, Reset)
	}

	log.Printf("%sServing files from %s on port %d%s", Green, dir, port, Reset)

	fmt.Println("Scan the QR code below to access the file list:")
	fmt.Println(qrCode.ToString(true))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
