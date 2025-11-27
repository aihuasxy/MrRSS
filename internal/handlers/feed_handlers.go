package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// HandleFeeds returns all feeds.
func (h *Handler) HandleFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(feeds)
}

// HandleAddFeed adds a new feed subscription.
func (h *Handler) HandleAddFeed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL        string `json:"url"`
		Category   string `json:"category"`
		Title      string `json:"title"`
		ScriptPath string `json:"script_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var err error
	if req.ScriptPath != "" {
		// Add feed using custom script
		err = h.Fetcher.AddScriptSubscription(req.ScriptPath, req.Category, req.Title)
	} else {
		// Add feed using URL
		err = h.Fetcher.AddSubscription(req.URL, req.Category, req.Title)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleDeleteFeed deletes a feed subscription.
func (h *Handler) HandleDeleteFeed(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.DB.DeleteFeed(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleUpdateFeed updates a feed's properties.
func (h *Handler) HandleUpdateFeed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID         int64  `json:"id"`
		Title      string `json:"title"`
		URL        string `json:"url"`
		Category   string `json:"category"`
		ScriptPath string `json:"script_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.DB.UpdateFeed(req.ID, req.Title, req.URL, req.Category, req.ScriptPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
