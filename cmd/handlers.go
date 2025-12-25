package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Clean the path to prevent path traversal attacks.
	cleanPath := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
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
				// Use mimetype to detect if the file is an image.
				mtype, err := mimetype.DetectFile(filepath.Join(fullPath, file.Name()))
				if err == nil && strings.HasPrefix(mtype.String(), "image/") {
					fileInfo.IsImage = true
				} else {
					fileInfo.IsImage = false
				}

				ext := strings.ToLower(filepath.Ext(file.Name()))
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

			return
		}
	} else {
		// Detect MIME type and set Content-Type header before serving the file.
		mtype, err := mimetype.DetectFile(fullPath)
		if err != nil {
			log.Printf("%sError detecting MIME type: %s%s", Red, err, Reset)
			w.WriteHeader(http.StatusInternalServerError)
			tmpl500.Execute(w, nil)

			return
		}

		w.Header().Set("Content-Type", mtype.String())
		http.ServeFile(w, r, fullPath)
	}
}

// Use github.com/gabriel-vasile/mimetype to detect MIME type for thumbnails as well.
func thumbnailHandler(w http.ResponseWriter, r *http.Request) {
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

	// Set Content-Type for thumbnails.
	mtype, err := mimetype.DetectFile(thumbnailPath)

	if err == nil {
		w.Header().Set("Content-Type", mtype.String())
	}

	http.ServeFile(w, r, thumbnailPath)
}
