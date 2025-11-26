package handlers

import (
	"encoding/json"
	"net/http"

	"MrRSS/internal/rules"
)

// HandleApplyRule applies a rule to matching articles
func (h *Handler) HandleApplyRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var rule rules.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(rule.Actions) == 0 {
		http.Error(w, "No actions specified", http.StatusBadRequest)
		return
	}

	engine := rules.NewEngine(h.DB)
	affected, err := engine.ApplyRule(rule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Success  bool `json:"success"`
		Affected int  `json:"affected"`
	}{
		Success:  true,
		Affected: affected,
	}
	json.NewEncoder(w).Encode(response)
}
