package main

import (
	"flag"
	"html/template"
)

var (
	dir          string
	port         int
	user         string
	pass         string
	logRequests  bool
	tmpl         *template.Template
	tmpl404      *template.Template
	tmpl500      *template.Template
	thumbnailDir string
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
