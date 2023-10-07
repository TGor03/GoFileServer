package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const directoryPath = "/mnt/volume_sfo3_01"

func main() {
	http.HandleFunc("/", listFiles)
	http.HandleFunc("/download/", downloadFile)

	fmt.Println("Server started at :1337")
	http.ListenAndServe(":1337", nil)
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	// Get the requested directory path from the URL
	requestedPath := filepath.Join(directoryPath, r.URL.Path[len("/"):])

	// Ensure that the requested path is within the allowed directory
	if !strings.HasPrefix(requestedPath, directoryPath) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get a list of all files and subdirectories in the requested directory
	files, err := getFilesAndSubdirectoriesInDirectory(requestedPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Generate an HTML list of file and folder links
	html := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Downloads</title>
			<style>
				body {
					background-color: #1a1a1a;
					color: #fff;
					font-family: Arial, sans-serif;
					padding: 20px;
				}
				h1 {
					color: #f3f3f3;
					text-align: center; /* Center-align the text */
				}
				ul {
					list-style: none;
					padding: 0;
				}
				li {
					margin: 10px 0;
				}
				a {
					color: #f3f3f3;
					text-decoration: none;
				}
				a:visited {
					color: #f3f3f3;
					text-decoration: none;
				}
				a:hover {
					color: #f07fae;
				}
			</style>
		</head>
		<body>
			<h1>Downloads - ` + strings.Split(requestedPath+"/", "/mnt/volume_sfo3_01")[1] + `</h1>
			<ul>
	`

	// Add a link to navigate to the parent directory (if not at the root)
	if requestedPath != directoryPath {
		parentDirectory := filepath.Dir(requestedPath)
		relativePath := strings.TrimPrefix(parentDirectory, directoryPath)
		html += fmt.Sprintf("<li> - <a href=\"%s\">Parent Directory</a></li>", strings.Replace("/"+relativePath, "//", "/", -1))
	}

	for _, item := range files {
		if item.IsDir() {
			// For subdirectories, create a link to navigate into the directory
			relativePath := strings.TrimPrefix(filepath.Join(requestedPath, item.Name()), directoryPath)
			html += fmt.Sprintf("<li> - <a href=\"%s\">%s</a></li>", strings.Replace(relativePath, "//", "", -1), item.Name())
		} else {
			// For files, create a link to download the file
			relativePath := strings.TrimPrefix(filepath.Join(requestedPath, item.Name()), directoryPath)
			html += fmt.Sprintf("<li> - <a download href=\"/download/%s\">%s</a></li>", relativePath, item.Name())
		}
	}

	html += `
			</ul>
		</body>
		</html>
	`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	// Extract the requested file path from the URL
	filePath := filepath.Join(directoryPath, r.URL.Path[len("/download/"):])

	// Ensure that the requested file path is within the allowed directory
	if !strings.HasPrefix(filePath, directoryPath) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if the file exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
}

func getFilesAndSubdirectoriesInDirectory(dir string) ([]os.DirEntry, error) {
	var entries []os.DirEntry

	// Open the directory
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// Collect the directory entries (both files and subdirectories)
	entries = append(entries, dirEntries...)

	return entries, nil
}
