// Package core contains the main Handler struct and core HTTP handlers for the application.
// It defines the Handler struct which holds dependencies like the database and fetcher.
package core

import (
	"sync"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/discovery"
	"MrRSS/internal/feed"
	"MrRSS/internal/translation"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Discovery timeout constants
const (
	// SingleFeedDiscoveryTimeout is the timeout for discovering feeds from a single source
	SingleFeedDiscoveryTimeout = 90 * time.Second
	// BatchDiscoveryTimeout is the timeout for discovering feeds from all sources
	BatchDiscoveryTimeout = 5 * time.Minute
)

// DiscoveryState represents the current state of a discovery operation
type DiscoveryState struct {
	IsRunning  bool                       `json:"is_running"`
	Progress   discovery.Progress         `json:"progress"`
	Feeds      []discovery.DiscoveredBlog `json:"feeds,omitempty"`
	Error      string                     `json:"error,omitempty"`
	IsComplete bool                       `json:"is_complete"`
}

// Handler holds all dependencies for HTTP handlers.
type Handler struct {
	DB               *database.DB
	Fetcher          *feed.Fetcher
	Translator       translation.Translator
	DiscoveryService *discovery.Service
	App              *application.App // Wails app instance for browser integration

	// Discovery state tracking for polling-based progress
	DiscoveryMu          sync.RWMutex
	SingleDiscoveryState *DiscoveryState
	BatchDiscoveryState  *DiscoveryState
}

// NewHandler creates a new Handler with the given dependencies.
func NewHandler(db *database.DB, fetcher *feed.Fetcher, translator translation.Translator) *Handler {
	return &Handler{
		DB:               db,
		Fetcher:          fetcher,
		Translator:       translator,
		DiscoveryService: discovery.NewService(),
	}
}

// SetApp sets the Wails application instance for browser integration.
// This is called after app initialization in main.go.
func (h *Handler) SetApp(app *application.App) {
	h.App = app
}
