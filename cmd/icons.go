package main

import (
	_ "embed"
	"net/http"
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

const (
	iconPath = "/icons/"
	mp3      = "mp3.png"
	pdf      = "pdf.png"
	txt      = "txt.png"
	xls      = "xls.png"
	wmv      = "wmv.png"
	mov      = "mov.png"
	mpg      = "mpg.png"
	avi      = "avi.png"
	file     = "file.png"
)

func serveEmbeddedIcon(w http.ResponseWriter, iconName string) {
	var data []byte
	switch iconName {
	case mp3:
		data = mp3Icon
	case pdf:
		data = pdfIcon
	case txt:
		data = txtIcon
	case xls:
		data = xlsIcon
	case wmv:
		data = wmvIcon
	case mov:
		data = movIcon
	case mpg:
		data = mpgIcon
	case avi:
		data = aviIcon
	case file:
		data = fileIcon

	default:
		http.NotFound(w, nil)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(data)
}

func getFileIcon(ext string) string {
	switch ext {
	case ".txt":
		return iconPath + txt
	case ".rtf":
		return iconPath + txt
	case ".pdf":
		return iconPath + pdf
	case ".xls":
		return iconPath + xls
	// case ".doc":
	// 	return "/icons/doc.png"
	case ".wmv":
		return iconPath + wmv
	case ".mov":
		return iconPath + mov
	case ".mpg":
		return iconPath + mpg
	case ".avi":
		return iconPath + avi
	// case ".mkv":
	// return "/icons/mkv.png"
	// case ".mp4":
	// return "/icons/mp4.png"
	case ".mp3":
		return iconPath + mp3
	// case ".csv":
	// return "/icons/csv.png"
	// case ".aac", ".flac", ".m4a", ".ogg", ".wav":
	// return "/icons/audio.png"
	// case ".zip":
	// return "/icons/zip.png"
	default:
		return iconPath + file
	}
}
