package opml

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/opml"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// HandleOPMLImport handles OPML file import.
func HandleOPMLImport(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleOPMLImport: ContentLength: %d", r.ContentLength)
	contentType := r.Header.Get("Content-Type")
	log.Printf("HandleOPMLImport: Content-Type: %s", contentType)

	var file io.Reader

	if strings.Contains(contentType, "multipart/form-data") {
		f, header, err := r.FormFile("file")
		if err != nil {
			log.Printf("Error getting form file: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer f.Close()
		log.Printf("HandleOPMLImport: Received file %s, size: %d", header.Filename, header.Size)

		if header.Size == 0 {
			http.Error(w, "Uploaded file is empty", http.StatusBadRequest)
			return
		}
		file = f
	} else {
		// Handle raw body upload
		file = r.Body
		defer r.Body.Close()
	}

	feeds, err := opml.Parse(file)
	if err != nil {
		log.Printf("Error parsing OPML: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Import feeds synchronously so they appear in the sidebar immediately
	var feedIDs []int64
	for _, f := range feeds {
		var feedID int64
		var err error

		// Check if feed has XPath configuration
		if f.Type == "HTML+XPath" || f.Type == "XML+XPath" {
			feedID, err = h.Fetcher.AddXPathSubscription(
				f.URL, f.Category, f.Title, f.Type,
				f.XPathItem, f.XPathItemTitle, f.XPathItemContent, f.XPathItemUri,
				f.XPathItemAuthor, f.XPathItemTimestamp, f.XPathItemTimeFormat,
				f.XPathItemThumbnail, f.XPathItemCategories, f.XPathItemUid,
			)
		} else {
			feedID, err = h.Fetcher.ImportSubscription(f.Title, f.URL, f.Category)
		}

		if err != nil {
			log.Printf("Error importing feed %s: %v", f.Title, err)
			continue
		}
		feedIDs = append(feedIDs, feedID)
	}

	// Fetch articles for the newly imported feeds asynchronously with progress tracking
	if len(feedIDs) > 0 {
		go func() {
			h.Fetcher.FetchFeedsByIDs(context.Background(), feedIDs)
		}()
	}

	w.WriteHeader(http.StatusOK)
}

// HandleOPMLExport handles OPML file export.
func HandleOPMLExport(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := opml.Generate(feeds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=subscriptions.opml")
	w.Header().Set("Content-Type", "text/xml")
	w.Write(data)
}

// HandleOPMLImportDialog opens a file dialog to select OPML file for import.
func HandleOPMLImportDialog(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if h.App == nil {
		log.Printf("App instance not available for dialog integration")
		http.Error(w, "Dialog integration not available", http.StatusInternalServerError)
		return
	}

	// Open file dialog to select OPML file
	filePath, err := h.App.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title: "Select OPML File",
		Filters: []application.FileFilter{
			{
				DisplayName: "OPML Files",
				Pattern:     "*.opml;*.xml",
			},
			{
				DisplayName: "All Files",
				Pattern:     "*",
			},
		},
	}).PromptForSingleSelection()
	if err != nil {
		log.Printf("Error opening file dialog: %v", err)
		http.Error(w, "Failed to open file dialog", http.StatusInternalServerError)
		return
	}

	if filePath == "" {
		// User cancelled the dialog
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
		return
	}

	// Read the selected file
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening selected file: %v", err)
		http.Error(w, "Failed to open selected file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Parse OPML content
	feeds, err := opml.Parse(file)
	if err != nil {
		log.Printf("Error parsing OPML: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Import feeds synchronously so they appear in the sidebar immediately
	var feedIDs []int64
	for _, f := range feeds {
		var feedID int64
		var err error

		// Check if feed has XPath configuration
		if f.Type == "HTML+XPath" || f.Type == "XML+XPath" {
			feedID, err = h.Fetcher.AddXPathSubscription(
				f.URL, f.Category, f.Title, f.Type,
				f.XPathItem, f.XPathItemTitle, f.XPathItemContent, f.XPathItemUri,
				f.XPathItemAuthor, f.XPathItemTimestamp, f.XPathItemTimeFormat,
				f.XPathItemThumbnail, f.XPathItemCategories, f.XPathItemUid,
			)
		} else {
			feedID, err = h.Fetcher.ImportSubscription(f.Title, f.URL, f.Category)
		}

		if err != nil {
			log.Printf("Error importing feed %s: %v", f.Title, err)
			continue
		}
		feedIDs = append(feedIDs, feedID)
	}

	// Fetch articles for the newly imported feeds asynchronously with progress tracking
	if len(feedIDs) > 0 {
		go func() {
			h.Fetcher.FetchFeedsByIDs(context.Background(), feedIDs)
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"feedCount": len(feeds),
		"filePath":  filePath,
	})
}

// HandleOPMLExportDialog opens a save dialog to export OPML file.
func HandleOPMLExportDialog(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if h.App == nil {
		log.Printf("App instance not available for dialog integration")
		http.Error(w, "Dialog integration not available", http.StatusInternalServerError)
		return
	}

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

	// Open save dialog
	filePath, err := h.App.Dialog.SaveFileWithOptions(&application.SaveFileDialogOptions{
		Title:    "Save OPML File",
		Filename: "subscriptions.opml",
		Filters: []application.FileFilter{
			{
				DisplayName: "OPML Files",
				Pattern:     "*.opml",
			},
			{
				DisplayName: "XML Files",
				Pattern:     "*.xml",
			},
			{
				DisplayName: "All Files",
				Pattern:     "*",
			},
		},
	}).PromptForSingleSelection()
	if err != nil {
		log.Printf("Error opening save dialog: %v", err)
		http.Error(w, "Failed to open save dialog", http.StatusInternalServerError)
		return
	}

	if filePath == "" {
		// User cancelled the dialog
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
		return
	}

	// Write OPML content to selected file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		log.Printf("Error writing OPML file: %v", err)
		http.Error(w, "Failed to write OPML file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"filePath": filePath,
	})
}
