package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/skip2/go-qrcode"
)

func main() {
	flag.Parse()

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
