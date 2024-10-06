# go-local-fileserver

`go-local-fileserver` is a lightweight, cross-platform local file server built in Golang with directory listing, file downloads, and extensive customization options. This server is perfect for sharing files on a local network or hosting files on your machine with minimal setup. It supports a range of useful features, from QR code file access to media previews and basic authentication.

## Features

* <b>File Downloads:</b> Quickly and easily download any file listed in the served directory.
* <b>QR Code Links:</b> Generate and display QR codes for each file, allowing easy access via mobile devices.
* <b>Thumbnail Previews:</b> Automatic thumbnail generation for image files, with efficient cleanup on server shutdown.
* <b>Subdirectory Support:</b> Seamlessly browse through directories and subdirectories.
* <b>Basic Authentication:</b> Secure your directories with optional basic authentication.
* <b>CLI Logging:</b> Real-time logging of server activity and requests.
* <b>Sorting & Grouping:</b> Sort files by name or date, and group them by file type.
* <b>Hidden Files:</b> Files prefixed with . are hidden from the listing and cannot be accessed.
* <b>File Type Icons:</b> Visual indicators for different file types.
* <b>Customizable Interface:</b> Easily modify the server’s HTML template to change the appearance of the file listing.

## Planned Features

* <b>User Management:</b> Support for multiple users with custom permissions.
* <b>Advanced Logging:</b> Track full request history with detailed analytics.
* <b>File Type Whitelisting/Blacklisting:</b> Control allowed file types for download or restrict specific formats.
* <b>Directory Whitelisting/Blacklisting:</b> Limit access to specific directories based on settings.
* <b>ZIP Compression:</b> Download entire directories as compressed ZIP files via the interface.
* <b>Smart Search:</b> Advanced search options, allowing filtering by media types (videos, text, audio, etc.).
* <b>File Uploads:</b> Enable file uploads via the web interface.
* <b>Admin panel:</b> To control users, statistics and config server in real time.
* <b>Rate limiting:</b> Limit the number of requests or file downloads per user / IP to prevent abuse.
* <b>External storages support:</b> If you have a lot of files, you can use external storage like S3, Various Clouds, etc.
* <b>Fancy audio player:</b> Not sure if I have enough skill & time.
* <b>Fancy video player:</b> Not sure if I have enough skill & time (2).

## Installation

### Via executable

Just download the executable for your platform from the [releases page](https://github.com/lazybark/go-local-fileserver/releases) and run it in the directory you want to serve.

Add to PATH for global access.

### Via Go

```bash
git clone https://github.com/yourusername/go-local-fileserver.git
cd go-local-fileserver

go build
```

## Usage

```bash
go-local-fileserver [flags]
```

### Flags

```bash
  -dir string
        Directory to serve files from (default "./public")
  -log-requests
        Log HTTP requests
  -user string
        Username for basic authentication
  -pass string
        Password for basic authentication
  -port int
        Port to run the server on (default 8080)
```

### Interface

* <b>Sorting & Grouping:</b> Sort files by name or date, group them by file type to simplify browsing.
* <b>QR Codes:</b> Click on the QR icon next to any file to get a scannable link for mobile download.
* <b>Authentication:</b> Use the -auth flag to enable basic authentication for sensitive directories.
* <b>Custom Themes:</b> Modify the template.html file in the templates/ folder to change the look and feel.
