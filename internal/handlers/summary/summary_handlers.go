package summary

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"MrRSS/internal/feed"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/summary"
	"MrRSS/internal/utils"
)

// HandleSummarizeArticle generates a summary for an article's content.
func HandleSummarizeArticle(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ArticleID int64  `json:"article_id"`
		Length    string `json:"length"` // "short", "medium", "long"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate length parameter
	summaryLength := summary.Medium
	switch req.Length {
	case "short":
		summaryLength = summary.Short
	case "long":
		summaryLength = summary.Long
	case "medium", "":
		summaryLength = summary.Medium
	default:
		http.Error(w, "Invalid length parameter. Use 'short', 'medium', or 'long'", http.StatusBadRequest)
		return
	}

	// Get the article content
	content, err := getArticleContent(h, req.ArticleID)
	if err != nil {
		log.Printf("Error getting article content for summary: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if content == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"summary":      "",
			"is_too_short": true,
			"error":        "No content available for this article",
		})
		return
	}

	// Get summary provider from settings (with default)
	provider, err := h.DB.GetSetting("summary_provider")
	if err != nil || provider == "" {
		provider = "local" // Default to local algorithm
	}

	var result summary.SummaryResult
	usedFallback := false

	if provider == "ai" {
		// Check if AI usage limit is reached - fallback to local if so
		if h.AITracker.IsLimitReached() {
			log.Printf("AI usage limit reached, falling back to local summarization")
			summarizer := summary.NewSummarizer()
			result = summarizer.Summarize(content, summaryLength)
			usedFallback = true
		} else {
			// Use AI summarization (use encrypted method for API key)
			apiKey, err := h.DB.GetEncryptedSetting("ai_api_key")
			if err != nil || apiKey == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "missing_ai_api_key",
				})
				return
			}

			// Apply rate limiting for AI requests
			h.AITracker.WaitForRateLimit()

			// Get global AI settings
			endpoint, _ := h.DB.GetSetting("ai_endpoint")
			model, _ := h.DB.GetSetting("ai_model")
			systemPrompt, _ := h.DB.GetSetting("ai_summary_prompt")

			aiSummarizer := summary.NewAISummarizer(apiKey, endpoint, model)
			if systemPrompt != "" {
				aiSummarizer.SetSystemPrompt(systemPrompt)
			}
			aiResult, err := aiSummarizer.Summarize(content, summaryLength)
			if err != nil {
				log.Printf("Error generating AI summary: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			result = aiResult

			// Track AI usage
			h.AITracker.TrackSummary(content, result.Summary)
		}
	} else {
		// Use local algorithm
		summarizer := summary.NewSummarizer()
		result = summarizer.Summarize(content, summaryLength)
	}

	response := map[string]interface{}{
		"summary":        result.Summary,
		"sentence_count": result.SentenceCount,
		"is_too_short":   result.IsTooShort,
	}
	if usedFallback {
		response["used_fallback"] = true
	}

	json.NewEncoder(w).Encode(response)
}

// getArticleContent fetches the content of an article by ID
func getArticleContent(h *core.Handler, articleID int64) (string, error) {
	// Get the article directly by ID (more efficient and includes hidden articles)
	article, err := h.DB.GetArticleByID(articleID)
	if err != nil {
		return "", nil
	}

	// Get the feed (including script path for custom script feeds)
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		return "", err
	}

	var targetFeed *struct {
		URL        string
		ScriptPath string
	}
	for _, f := range feeds {
		if f.ID == article.FeedID {
			targetFeed = &struct {
				URL        string
				ScriptPath string
			}{
				URL:        f.URL,
				ScriptPath: f.ScriptPath,
			}
			break
		}
	}

	if targetFeed == nil {
		return "", nil
	}

	// Parse the feed to get fresh content (handles both regular URLs and custom scripts)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parsedFeed, err := h.Fetcher.ParseFeedWithScript(ctx, targetFeed.URL, targetFeed.ScriptPath)
	if err != nil {
		return "", err
	}

	// Find the article in the feed by URL (use normalized comparison for robustness)
	for _, item := range parsedFeed.Items {
		if utils.URLsMatch(item.Link, article.URL) {
			// Use the centralized content extraction logic to ensure consistency
			content := feed.ExtractContent(item)
			return utils.CleanHTML(content), nil
		}
	}

	return "", nil
}
