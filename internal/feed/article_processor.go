package feed

import (
	"MrRSS/internal/models"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

// processArticles processes RSS feed items and converts them to Article models
func (f *Fetcher) processArticles(feed models.Feed, items []*gofeed.Item) []*models.Article {
	// Check translation settings
	translationEnabledStr, _ := f.db.GetSetting("translation_enabled")
	targetLang, _ := f.db.GetSetting("target_language")
	translationEnabled := translationEnabledStr == "true"

	var articles []*models.Article

	for _, item := range items {
		published := time.Now()
		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
		}

		imageURL := extractImageURL(item)
		audioURL := extractAudioURL(item)

		// Extract Media RSS content (YouTube feeds)
		mediaTitle := extractMediaTitle(item)
		mediaDescription := extractMediaDescription(item)

		// Extract content from RSS item (prefer media:description, then Content, then Description)
		content := mediaDescription
		if content == "" {
			content = item.Content
		}
		if content == "" {
			content = item.Description
		}

		// Determine title: prefer media:title if available, then item.Title, then generate from content
		title := item.Title
		if mediaTitle != "" {
			title = mediaTitle
		}
		if title == "" {
			// Fallback to generating from the processed content
			title = generateTitleFromContent(content)
		}

		translatedTitle := ""
		if translationEnabled && f.translator != nil {
			t, err := f.translator.Translate(title, targetLang)
			if err == nil {
				translatedTitle = t
			}
		}

		article := &models.Article{
			FeedID:          feed.ID,
			Title:           title,
			URL:             item.Link,
			ImageURL:        imageURL,
			AudioURL:        audioURL,
			Content:         content,
			PublishedAt:     published,
			TranslatedTitle: translatedTitle,
		}
		articles = append(articles, article)
	}

	return articles
}

// extractImageURL extracts the image URL from a feed item
func extractImageURL(item *gofeed.Item) string {
	// Try item.Image first
	if item.Image != nil {
		return item.Image.URL
	}

	// Try Media RSS thumbnail (YouTube feeds use this)
	if thumbnailURL := extractMediaThumbnail(item); thumbnailURL != "" {
		return thumbnailURL
	}

	// Try enclosures for images (check various image MIME types)
	for _, enc := range item.Enclosures {
		if strings.HasPrefix(enc.Type, "image/") {
			return enc.URL
		}
	}

	// Fallback: Try to find image in description/content
	content := item.Content
	if content == "" {
		content = item.Description
	}

	re := regexp.MustCompile(`<img[^>]+src="([^">]+)"`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// extractAudioURL extracts the audio URL from a feed item (for podcasts)
func extractAudioURL(item *gofeed.Item) string {
	// Try enclosures for audio files
	for _, enc := range item.Enclosures {
		// Check for audio MIME types
		if strings.HasPrefix(enc.Type, "audio/") {
			return enc.URL
		}
	}

	return ""
}

// generateTitleFromContent generates a title from content when title is missing
func generateTitleFromContent(content string) string {
	if content == "" {
		return "Untitled Article"
	}

	// Remove HTML tags
	htmlTagRegex := regexp.MustCompile(`<[^>]+>`)
	plainText := htmlTagRegex.ReplaceAllString(content, "")

	// Trim whitespace
	plainText = strings.TrimSpace(plainText)

	// Limit to 100 characters
	if len(plainText) > 100 {
		plainText = plainText[:100] + "..."
	}

	// If still empty after cleaning, use default
	if plainText == "" {
		return "Untitled Article"
	}

	return plainText
}

// extractMediaThumbnail extracts the thumbnail URL from Media RSS extensions (used by YouTube)
func extractMediaThumbnail(item *gofeed.Item) string {
	if item.Extensions == nil {
		return ""
	}

	// Check for media:group extension (YouTube uses this structure)
	if mediaExt, ok := item.Extensions["media"]; ok {
		if groupExts, ok := mediaExt["group"]; ok && len(groupExts) > 0 {
			// Navigate to media:group's children
			if groupExts[0].Children != nil {
				if thumbnailExts, ok := groupExts[0].Children["thumbnail"]; ok && len(thumbnailExts) > 0 {
					// Get the URL from the thumbnail's attributes
					if thumbnailExts[0].Attrs != nil {
						if url, ok := thumbnailExts[0].Attrs["url"]; ok {
							return url
						}
					}
				}
			}
		}

		// Also check for direct media:thumbnail (some feeds use this)
		if thumbnailExts, ok := mediaExt["thumbnail"]; ok && len(thumbnailExts) > 0 {
			if thumbnailExts[0].Attrs != nil {
				if url, ok := thumbnailExts[0].Attrs["url"]; ok {
					return url
				}
			}
		}
	}

	return ""
}

// extractMediaTitle extracts the title from Media RSS extensions (used by YouTube)
func extractMediaTitle(item *gofeed.Item) string {
	if item.Extensions == nil {
		return ""
	}

	// Check for media:group extension (YouTube uses this structure)
	if mediaExt, ok := item.Extensions["media"]; ok {
		if groupExts, ok := mediaExt["group"]; ok && len(groupExts) > 0 {
			// Navigate to media:group's children
			if groupExts[0].Children != nil {
				if titleExts, ok := groupExts[0].Children["title"]; ok && len(titleExts) > 0 {
					return titleExts[0].Value
				}
			}
		}

		// Also check for direct media:title (some feeds use this)
		if titleExts, ok := mediaExt["title"]; ok && len(titleExts) > 0 {
			return titleExts[0].Value
		}
	}

	return ""
}

// extractMediaDescription extracts the description from Media RSS extensions (used by YouTube)
func extractMediaDescription(item *gofeed.Item) string {
	if item.Extensions == nil {
		return ""
	}

	// Check for media:group extension (YouTube uses this structure)
	if mediaExt, ok := item.Extensions["media"]; ok {
		if groupExts, ok := mediaExt["group"]; ok && len(groupExts) > 0 {
			// Navigate to media:group's children
			if groupExts[0].Children != nil {
				if descExts, ok := groupExts[0].Children["description"]; ok && len(descExts) > 0 {
					return descExts[0].Value
				}
			}
		}

		// Also check for direct media:description (some feeds use this)
		if descExts, ok := mediaExt["description"]; ok && len(descExts) > 0 {
			return descExts[0].Value
		}
	}

	return ""
}
