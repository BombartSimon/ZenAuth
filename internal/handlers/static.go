package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StaticFileHandler handles requests for static files with proper MIME types
func StaticFileHandler(w http.ResponseWriter, r *http.Request) {
	// Remove "/admin/" prefix from the path
	path := r.URL.Path[7:]

	// Log the requested path for debugging
	log.Printf("Static file requested: %s", path)

	// Default to index.html if path is empty
	if path == "" {
		path = "index.html"
	}

	// Special handling for /admin/assets path
	if strings.HasPrefix(path, "assets/") {
		serveAssetFile(w, r, path)
		return
	}

	// Special handling for JavaScript files
	if strings.HasPrefix(path, "js/") {
		serveWebFile(w, r, path)
		return
	}

	// Construct the full file path
	filePath := filepath.Join("./web", path)
	log.Printf("Looking for file: %s", filePath)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File not found: %s", filePath)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		log.Printf("Error opening file: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Explicitly set content-type for specific file types
	ext := filepath.Ext(path)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	}

	// Serve the file with the correct content type
	http.ServeContent(w, r, path, info.ModTime(), file)
}

// serveAssetFile serves files from the assets directory
func serveAssetFile(w http.ResponseWriter, r *http.Request, path string) {
	// Strip the "assets/" prefix
	assetPath := path[len("assets/"):]

	// Construct the full asset path
	filePath := filepath.Join("./assets", assetPath)
	log.Printf("Looking for asset: %s", filePath)

	// Open the asset file
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Asset not found: %s", filePath)
			http.Error(w, "Asset not found", http.StatusNotFound)
			return
		}
		log.Printf("Error opening asset file: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		log.Printf("Error getting asset file info: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set content-type based on file extension
	ext := filepath.Ext(assetPath)
	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	}

	// Serve the asset file
	http.ServeContent(w, r, assetPath, info.ModTime(), file)
}

// serveWebFile serves files from the web directory, preserving the directory structure
func serveWebFile(w http.ResponseWriter, r *http.Request, path string) {
	// Construct the full web file path
	filePath := filepath.Join("./web", path)
	log.Printf("Looking for web file: %s", filePath)

	// Open the web file
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Web file not found: %s", filePath)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		log.Printf("Error opening web file: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		log.Printf("Error getting web file info: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set content-type based on file extension
	ext := filepath.Ext(path)
	switch ext {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	}

	// Serve the web file
	http.ServeContent(w, r, path, info.ModTime(), file)
}
