package main

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
