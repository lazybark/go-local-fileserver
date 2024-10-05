package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	dir  string
	port int
	user string
	pass string
)

func init() {
	flag.StringVar(&dir, "dir", "./public", "Directory to serve files from")
	flag.IntVar(&port, "port", 8080, "Port to run the server on")
	flag.StringVar(&user, "user", "", "Username for basic authentication")
	flag.StringVar(&pass, "pass", "", "Password for basic authentication")
}

func main() {
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if user != "" && pass != "" {
			u, p, ok := r.BasicAuth()
			if !ok || u != user || p != pass {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)

				return
			}
		}

		decodedPath, err := url.PathUnescape(r.URL.Path)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)

			return
		}

		// Check if the requested path is a hidden file or directory.
		if strings.HasPrefix(filepath.Base(decodedPath), ".") {
			http.Error(w, "Forbidden", http.StatusForbidden)

			return
		}

		// Clean the path to prevent path traversal attacks.
		cleanPath := filepath.Clean(decodedPath)
		path := filepath.Join(dir, cleanPath)

		// Ensure the path is within the base directory.
		baseDir, err := filepath.Abs(dir)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}

		if !strings.HasPrefix(absPath, baseDir) {
			http.Error(w, "Forbidden", http.StatusForbidden)

			return
		}

		info, err := os.Stat(path)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)

			return
		}

		if info.IsDir() {
			files, err := os.ReadDir(path)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, "<h1>Directory listing for %s</h1>", r.URL.Path)
			fmt.Fprintln(w, "<ul>")

			for _, file := range files {
				name := file.Name()

				// Skip hidden files and directories.
				if strings.HasPrefix(name, ".") {
					continue
				}

				encodedName := url.PathEscape(name)

				link := filepath.Join(r.URL.Path, encodedName)
				if file.IsDir() {
					link += "/"
				}

				fmt.Fprintf(w, `<li><a href="%s">%s</a></li>`, link, name)
			}

			fmt.Fprintln(w, "</ul>")
		} else {
			mimeType := mime.TypeByExtension(filepath.Ext(path))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}

			w.Header().Set("Content-Type", mimeType)
			http.ServeFile(w, r, path)
		}
	})

	log.Printf("Serving %s on HTTP port: %d\n", dir, port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
