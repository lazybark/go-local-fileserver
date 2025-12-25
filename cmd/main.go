package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/mdp/qrterminal/v3"
)

//go:embed template.html
var mainTemplate string

//go:embed 404.html
var notFoundTemplate string

//go:embed 500.html
var errorTemplate string

func main() {
	flag.Parse()

	// Resolve the directory to an absolute path.
	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("Error resolving directory: %s", err)
	}
	dir = dirAbs

	// Set rootName to base name of dir if not provided.
	if rootName == "" {
		rootName = filepath.Base(dir)
	}

	// Check if the configured directory exists before starting the server.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatalf("%sConfigured directory does not exist: %s%s", Red, dir, Reset)
	}

	// TODO: use https://github.com/charmbracelet/bubbletea for a better TUI
	// https://themarkokovacevic.com/posts/terminal-ui-with-bubbletea/
	// https://github.com/charmbracelet/lipgloss
	// https://github.com/charmbracelet/bubbles
	tmpl = template.Must(template.New("main").Parse(mainTemplate))
	tmpl404 = template.Must(template.New("404").Parse(notFoundTemplate))
	tmpl500 = template.Must(template.New("500").Parse(errorTemplate))

	// Ensure the thumbnails directory is deleted on exit.
	thumbnailDir = "thumbnails"
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

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/thumbnail/", thumbnailHandler)
	http.HandleFunc("/icons/mp3.png", func(w http.ResponseWriter, r *http.Request) { serveEmbeddedIcon(w, "mp3.png") })
	http.HandleFunc("/icons/pdf.png", func(w http.ResponseWriter, r *http.Request) { serveEmbeddedIcon(w, "pdf.png") })
	http.HandleFunc("/icons/txt.png", func(w http.ResponseWriter, r *http.Request) { serveEmbeddedIcon(w, "txt.png") })
	http.HandleFunc("/icons/xls.png", func(w http.ResponseWriter, r *http.Request) { serveEmbeddedIcon(w, "xls.png") })
	http.HandleFunc("/icons/wmv.png", func(w http.ResponseWriter, r *http.Request) { serveEmbeddedIcon(w, "wmv.png") })
	http.HandleFunc("/icons/mov.png", func(w http.ResponseWriter, r *http.Request) { serveEmbeddedIcon(w, "mov.png") })
	http.HandleFunc("/icons/mpg.png", func(w http.ResponseWriter, r *http.Request) { serveEmbeddedIcon(w, "mpg.png") })
	http.HandleFunc("/icons/avi.png", func(w http.ResponseWriter, r *http.Request) { serveEmbeddedIcon(w, "avi.png") })
	http.HandleFunc("/icons/file.png", func(w http.ResponseWriter, r *http.Request) { serveEmbeddedIcon(w, "file.png") })
	http.Handle("/icons/", http.StripPrefix("/icons/", http.FileServer(http.Dir("./icons")))) // fallback for other icons

	// Get the local IP address.
	ip, err := getLocalIP()
	if err != nil {
		log.Fatalf("%sError getting local IP address: %s%s", Red, err, Reset)
	}

	// Log the IP address and the URL to access the file list.
	url := fmt.Sprintf("http://%s:%d/", ip, port)
	log.Printf("%sServer is running at %s%s", Green, url, Reset)

	log.Printf("%sServing files from %s on port %d%s", Green, dir, port, Reset)

	fmt.Println("Scan the QR code below to access the file list:")
	qrterminal.GenerateHalfBlock(url, qrterminal.M, os.Stdout)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
