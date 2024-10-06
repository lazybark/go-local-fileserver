package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

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
