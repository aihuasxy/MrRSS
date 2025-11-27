package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"MrRSS/internal/utils"
)

// HandleGetScriptsDir returns the path to the scripts directory
func (h *Handler) HandleGetScriptsDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scriptsDir, err := utils.GetScriptsDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"scripts_dir": scriptsDir,
	})
}

// HandleOpenScriptsDir opens the scripts directory in the system file explorer
func (h *Handler) HandleOpenScriptsDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scriptsDir, err := utils.GetScriptsDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Open the directory based on the OS
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", scriptsDir)
	case "darwin":
		cmd = exec.Command("open", scriptsDir)
	case "linux":
		cmd = exec.Command("xdg-open", scriptsDir)
	default:
		http.Error(w, "Unsupported platform", http.StatusBadRequest)
		return
	}

	if err := cmd.Start(); err != nil {
		http.Error(w, "Failed to open directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":      "opened",
		"scripts_dir": scriptsDir,
	})
}

// HandleListScripts returns a list of available scripts in the scripts directory
func (h *Handler) HandleListScripts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scriptsDir, err := utils.GetScriptsDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Valid script extensions
	validExtensions := map[string]bool{
		".py":  true,
		".sh":  true,
		".ps1": true,
		".js":  true,
		".rb":  true,
	}

	var scripts []map[string]string

	err = filepath.Walk(scriptsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path from scripts directory
		relPath, err := filepath.Rel(scriptsDir, path)
		if err != nil {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))
		scriptType := ""

		if validExtensions[ext] {
			switch ext {
			case ".py":
				scriptType = "Python"
			case ".sh":
				scriptType = "Shell"
			case ".ps1":
				scriptType = "PowerShell"
			case ".js":
				scriptType = "Node.js"
			case ".rb":
				scriptType = "Ruby"
			}

			scripts = append(scripts, map[string]string{
				"name": info.Name(),
				"path": relPath,
				"type": scriptType,
			})
		}

		return nil
	})

	if err != nil {
		http.Error(w, "Error listing scripts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if scripts == nil {
		scripts = []map[string]string{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"scripts":     scripts,
		"scripts_dir": scriptsDir,
	})
}
