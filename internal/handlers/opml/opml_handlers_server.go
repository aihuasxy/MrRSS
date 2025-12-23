//go:build server

package opml

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/opml"
)

// HandleOPMLImport handles OPML file import for server mode.
func HandleOPMLImport(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleOPMLImport: ContentLength: %d", r.ContentLength)

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the file
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error getting file: %v", err)
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received file: %s, size: %d", header.Filename, header.Size)

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Parse OPML
	feeds, err := opml.Parse(strings.NewReader(string(content)))
	if err != nil {
		log.Printf("Error parsing OPML: %v", err)
		http.Error(w, "Failed to parse OPML file", http.StatusBadRequest)
		return
	}

	log.Printf("Parsed %d feeds from OPML", len(feeds))

	// Import feeds
	imported := 0
	for _, feed := range feeds {
		_, err := h.DB.AddFeed(&feed)
		if err != nil {
			log.Printf("Error importing feed %s: %v", feed.URL, err)
			continue
		}
		imported++
	}

	log.Printf("Successfully imported %d feeds", imported)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"imported": imported,
		"total":    len(feeds),
	})
}

// HandleOPMLImportDialog is not available in server mode.
func HandleOPMLImportDialog(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	log.Printf("File dialog operations are not available in server mode")
	http.Error(w, "File dialog operations are not available in server mode. Use /api/opml/import endpoint with file upload instead.", http.StatusNotImplemented)
}

// HandleOPMLExport handles OPML export for server mode.
func HandleOPMLExport(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	// Get feeds data
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate OPML content
	data, err := opml.Generate(feeds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", "attachment; filename=subscriptions.opml")

	// Write OPML content
	w.Write(data)
}

// HandleOPMLExportDialog is not available in server mode.
func HandleOPMLExportDialog(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	log.Printf("File dialog not available in server mode")
	http.Error(w, "File dialog not available in server mode. Use /api/opml/export endpoint with direct download instead.", http.StatusNotImplemented)
}
