package main

import (
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

//go:embed icons/mp3.png
var mp3Icon []byte

//go:embed icons/pdf.png
var pdfIcon []byte

//go:embed icons/txt.png
var txtIcon []byte

//go:embed icons/xls.png
var xlsIcon []byte

//go:embed icons/wmv.png
var wmvIcon []byte

//go:embed icons/mov.png
var movIcon []byte

//go:embed icons/mpg.png
var mpgIcon []byte

//go:embed icons/avi.png
var aviIcon []byte

//go:embed icons/file.png
var fileIcon []byte

func serveEmbeddedIcon(w http.ResponseWriter, iconName string) {
	var data []byte
	switch iconName {
	case "mp3.png":
		data = mp3Icon
	case "pdf.png":
		data = pdfIcon
	case "txt.png":
		data = txtIcon
	case "xls.png":
		data = xlsIcon
	case "wmv.png":
		data = wmvIcon
	case "mov.png":
		data = movIcon
	case "mpg.png":
		data = mpgIcon
	case "avi.png":
		data = aviIcon
	case "file.png":
		data = fileIcon

	default:
		http.NotFound(w, nil)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(data)
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

	err = os.MkdirAll(filepath.Dir(dstPath), 0o755)
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
	case ".txt":
		return "/icons/txt.png"
	case ".rtf":
		return "/icons/txt.png"
	case ".pdf":
		return "/icons/pdf.png"
	case ".xls":
		return "/icons/xls.png"
	// case ".doc":
	// 	return "/icons/doc.png"
	case ".wmv":
		return "/icons/wmv.png"
	case ".mov":
		return "/icons/mov.png"
	case ".mpg":
		return "/icons/mpg.png"
	case ".avi":
		return "/icons/avi.png"
	// case ".mkv":
	// return "/icons/mkv.png"
	// case ".mp4":
	// return "/icons/mp4.png"
	case ".mp3":
		return "/icons/mp3.png"
	// case ".csv":
	// return "/icons/csv.png"
	// case ".aac", ".flac", ".m4a", ".ogg", ".wav":
	// return "/icons/audio.png"
	// case ".zip":
	// return "/icons/zip.png"
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
