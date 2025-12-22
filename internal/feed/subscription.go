package feed

import (
	"MrRSS/internal/models"
	"MrRSS/internal/utils"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xmlquery"
	"github.com/mmcdole/gofeed"
)

// AddSubscription adds a new feed subscription and returns the feed ID.
func (f *Fetcher) AddSubscription(url string, category string, customTitle string) (int64, error) {
	parsedFeed, err := f.fp.ParseURL(url)
	if err != nil {
		return 0, err
	}

	title := parsedFeed.Title
	if customTitle != "" {
		title = customTitle
	}

	feed := &models.Feed{
		Title:       title,
		URL:         url,
		Link:        parsedFeed.Link,
		Description: parsedFeed.Description,
		Category:    category,
	}

	if parsedFeed.Image != nil {
		feed.ImageURL = parsedFeed.Image.URL
	}

	return f.db.AddFeed(feed)
}

// AddScriptSubscription adds a new feed subscription that uses a custom script
// and returns the feed ID.
func (f *Fetcher) AddScriptSubscription(scriptPath string, category string, customTitle string) (int64, error) {
	// Validate script path
	if f.scriptExecutor == nil {
		return 0, &ScriptError{Message: "script executor not initialized"}
	}

	// Execute script to get initial feed info
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parsedFeed, err := f.scriptExecutor.ExecuteScript(ctx, scriptPath)
	if err != nil {
		return 0, err
	}

	title := parsedFeed.Title
	if customTitle != "" {
		title = customTitle
	}

	// Use a placeholder URL for script-based feeds
	url := "script://" + scriptPath

	feed := &models.Feed{
		Title:       title,
		URL:         url,
		Link:        parsedFeed.Link,
		Description: parsedFeed.Description,
		Category:    category,
		ScriptPath:  scriptPath,
	}

	if parsedFeed.Image != nil {
		feed.ImageURL = parsedFeed.Image.URL
	}

	return f.db.AddFeed(feed)
}

// AddXPathSubscription adds a new feed subscription that uses XPath expressions
// and returns the feed ID.
func (f *Fetcher) AddXPathSubscription(url string, category string, customTitle string, feedType string, xpathItem string, xpathItemTitle string, xpathItemContent string, xpathItemUri string, xpathItemAuthor string, xpathItemTimestamp string, xpathItemTimeFormat string, xpathItemThumbnail string, xpathItemCategories string, xpathItemUid string) (int64, error) {
	title := customTitle
	if title == "" {
		title = "XPath Feed"
	}

	feed := &models.Feed{
		Title:               title,
		URL:                 url,
		Category:            category,
		Type:                feedType,
		XPathItem:           xpathItem,
		XPathItemTitle:      xpathItemTitle,
		XPathItemContent:    xpathItemContent,
		XPathItemUri:        xpathItemUri,
		XPathItemAuthor:     xpathItemAuthor,
		XPathItemTimestamp:  xpathItemTimestamp,
		XPathItemTimeFormat: xpathItemTimeFormat,
		XPathItemThumbnail:  xpathItemThumbnail,
		XPathItemCategories: xpathItemCategories,
		XPathItemUid:        xpathItemUid,
	}

	return f.db.AddFeed(feed)
}

// ImportSubscription imports a feed subscription and returns the feed ID.
func (f *Fetcher) ImportSubscription(title, url, category string) (int64, error) {
	feed := &models.Feed{
		Title:    title,
		URL:      url,
		Link:     "", // Link will be fetched later when feed is refreshed
		Category: category,
	}
	return f.db.AddFeed(feed)
}

// ParseFeed parses an RSS feed from a URL and returns the parsed feed
func (f *Fetcher) ParseFeed(ctx context.Context, url string) (*gofeed.Feed, error) {
	return f.fp.ParseURLWithContext(url, ctx)
}

// ParseFeedWithScript parses an RSS feed, using a custom script or XPath if specified.
// If scriptPath is non-empty, it executes the script.
// If feed.Type is "HTML+XPath" or "XML+XPath", it uses XPath parsing.
// Otherwise, it fetches from the URL as normal.
// priority: true for high-priority requests (like article content fetching), false for normal requests (like feed refresh)
func (f *Fetcher) ParseFeedWithScript(ctx context.Context, url string, scriptPath string, priority bool) (*gofeed.Feed, error) {
	return f.ParseFeedWithFeed(ctx, &models.Feed{URL: url, ScriptPath: scriptPath}, priority)
}

// ParseFeedWithFeed parses a feed using the feed configuration (script or XPath)
func (f *Fetcher) ParseFeedWithFeed(ctx context.Context, feed *models.Feed, priority bool) (*gofeed.Feed, error) {
	if priority {
		// High priority requests get dedicated processing without interference from low priority operations
		f.priorityMu.Lock()
		defer f.priorityMu.Unlock()

		return f.parseFeedWithFeedInternal(ctx, feed, true)
	}

	// Normal priority requests
	return f.parseFeedWithFeedInternal(ctx, feed, false)
}

// parseFeedWithFeedInternal does the actual parsing work
func (f *Fetcher) parseFeedWithFeedInternal(ctx context.Context, feed *models.Feed, priority bool) (*gofeed.Feed, error) {
	if feed.ScriptPath != "" {
		// Execute the custom script to fetch feed
		if f.scriptExecutor == nil {
			return nil, &ScriptError{Message: "Script executor not initialized"}
		}

		// For high priority requests, use shorter timeout
		scriptCtx := ctx
		if priority {
			var cancel context.CancelFunc
			scriptCtx, cancel = context.WithTimeout(ctx, 15*time.Second) // Shorter timeout for content fetching
			defer cancel()
		}

		return f.scriptExecutor.ExecuteScript(scriptCtx, feed.ScriptPath)
	}

	// Check if this is an XPath-based feed
	if feed.Type == "HTML+XPath" || feed.Type == "XML+XPath" {
		// For high priority requests, use shorter timeout
		xpathCtx := ctx
		if priority {
			var cancel context.CancelFunc
			xpathCtx, cancel = context.WithTimeout(ctx, 15*time.Second) // Shorter timeout for content fetching
			defer cancel()
		}

		return f.parseFeedWithXPath(xpathCtx, feed)
	}

	// Use traditional URL-based fetching
	// For high priority requests, use shorter timeout
	fetchCtx := ctx
	if priority {
		var cancel context.CancelFunc
		fetchCtx, cancel = context.WithTimeout(ctx, 15*time.Second) // Shorter timeout for content fetching
		defer cancel()
	}

	return f.fp.ParseURLWithContext(feed.URL, fetchCtx)
}

// parseFeedWithXPath parses a feed using XPath expressions
func (f *Fetcher) parseFeedWithXPath(_ context.Context, feed *models.Feed) (*gofeed.Feed, error) {
	if feed.XPathItem == "" {
		return nil, fmt.Errorf("XPath item expression is required for XPath-based feeds")
	}

	// Fetch the content
	httpClient, err := utils.CreateHTTPClient("", 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}
	resp, err := httpClient.Get(feed.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Create gofeed.Feed
	parsedFeed := &gofeed.Feed{
		Title:       feed.Title,
		Link:        feed.URL,
		Description: feed.Description,
		Items:       make([]*gofeed.Item, 0),
	}

	// Parse based on type
	switch feed.Type {
	case "HTML+XPath":
		doc, err := htmlquery.Parse(strings.NewReader(string(body)))
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML: %w", err)
		}
		items := htmlquery.Find(doc, feed.XPathItem)
		if len(items) == 0 {
			return nil, fmt.Errorf("no items found with XPath: %s", feed.XPathItem)
		}

		// Process HTML items
		for _, item := range items {
			gofeedItem := f.extractItemFromHTMLNode(item, feed)
			parsedFeed.Items = append(parsedFeed.Items, gofeedItem)
		}
	case "XML+XPath":
		doc, err := xmlquery.Parse(strings.NewReader(string(body)))
		if err != nil {
			return nil, fmt.Errorf("failed to parse XML: %w", err)
		}
		items := xmlquery.Find(doc, feed.XPathItem)
		if len(items) == 0 {
			return nil, fmt.Errorf("no items found with XPath: %s", feed.XPathItem)
		}

		// Process XML items
		for _, item := range items {
			gofeedItem := f.extractItemFromXMLNode(item, feed)
			parsedFeed.Items = append(parsedFeed.Items, gofeedItem)
		}
	default:
		return nil, fmt.Errorf("unsupported feed type: %s", feed.Type)
	}

	return parsedFeed, nil
}

// extractItemFromHTMLNode extracts a gofeed.Item from an HTML node
func (f *Fetcher) extractItemFromHTMLNode(item *html.Node, feed *models.Feed) *gofeed.Item {
	gofeedItem := &gofeed.Item{}

	// Extract title
	if feed.XPathItemTitle != "" {
		if titleNode := htmlquery.FindOne(item, feed.XPathItemTitle); titleNode != nil {
			gofeedItem.Title = strings.TrimSpace(htmlquery.InnerText(titleNode))
		}
	}

	// Extract content
	if feed.XPathItemContent != "" {
		if contentNode := htmlquery.FindOne(item, feed.XPathItemContent); contentNode != nil {
			gofeedItem.Content = htmlquery.OutputHTML(contentNode, true)
		}
	}

	// Extract URI
	if feed.XPathItemUri != "" {
		var link string
		// Special handling for @href XPath - get href attribute directly from the item (a tag)
		if feed.XPathItemUri == "./@href" || feed.XPathItemUri == "@href" || feed.XPathItemUri == "href" {
			link = htmlquery.SelectAttr(item, "href")
		} else {
			// For other XPath expressions
			if uriNode := htmlquery.FindOne(item, feed.XPathItemUri); uriNode != nil {
				// Check if this XPath ends with an attribute selector
				if strings.Contains(feed.XPathItemUri, "@") {
					// For attribute XPath expressions, get the text content directly
					link = strings.TrimSpace(htmlquery.InnerText(uriNode))
				} else {
					// Try to get href attribute first (for element nodes)
					if attr := htmlquery.SelectAttr(uriNode, "href"); attr != "" {
						link = attr
					} else {
						// Fallback to inner text for other XPath expressions
						link = strings.TrimSpace(htmlquery.InnerText(uriNode))
					}
				}
			}
		}

		// Additional fallback: if no link found and item is an <a> tag, get href directly
		if link == "" && item != nil && item.Data == "a" {
			link = htmlquery.SelectAttr(item, "href")
		}

		// Resolve relative URLs to absolute URLs
		if link != "" && !strings.HasPrefix(link, "http") {
			baseURL, err := url.Parse(feed.URL)
			if err == nil {
				if ref, err := url.Parse(link); err == nil {
					gofeedItem.Link = baseURL.ResolveReference(ref).String()
				} else {
					gofeedItem.Link = link
				}
			} else {
				gofeedItem.Link = link
			}
		} else {
			gofeedItem.Link = link
		}
	}

	// If no URI was extracted, generate a unique URL for this article
	// This ensures each XPath article has a unique URL to prevent database conflicts
	if gofeedItem.Link == "" {
		// Use feed URL as base and append a hash of the title or content
		uniqueID := gofeedItem.Title
		if uniqueID == "" {
			// Fallback to content or a timestamp-based ID
			if gofeedItem.Content != "" {
				uniqueID = gofeedItem.Content
			} else {
				uniqueID = fmt.Sprintf("xpath-article-%d", time.Now().UnixNano())
			}
		}
		// Create a simple hash of the unique identifier
		hash := fmt.Sprintf("%x", len(uniqueID)) // Simple length-based hash for uniqueness
		gofeedItem.Link = fmt.Sprintf("%s#xpath-%s", feed.URL, hash)
	}

	// Extract author
	if feed.XPathItemAuthor != "" {
		if authorNode := htmlquery.FindOne(item, feed.XPathItemAuthor); authorNode != nil {
			gofeedItem.Author = &gofeed.Person{
				Name: strings.TrimSpace(htmlquery.InnerText(authorNode)),
			}
		}
	}

	// Extract timestamp
	if feed.XPathItemTimestamp != "" {
		if timeNode := htmlquery.FindOne(item, feed.XPathItemTimestamp); timeNode != nil {
			timeStr := strings.TrimSpace(htmlquery.InnerText(timeNode))
			// Remove icon text if present (e.g., "calendar_month 2025-12" -> "2025-12")
			if strings.Contains(timeStr, " ") {
				parts := strings.Split(timeStr, " ")
				// Find the date part (usually the last part that looks like a date)
				for i := len(parts) - 1; i >= 0; i-- {
					part := strings.TrimSpace(parts[i])
					if part != "" && (strings.Contains(part, "-") || strings.Contains(part, "/") || len(part) >= 4) {
						timeStr = part
						break
					}
				}
			}
			if timeStr != "" {
				var parsedTime time.Time
				var err error
				if feed.XPathItemTimeFormat != "" {
					parsedTime, err = time.Parse(feed.XPathItemTimeFormat, timeStr)
				} else {
					// Try common formats
					formats := []string{
						time.RFC3339,
						time.RFC1123,
						"2006-01-02T15:04:05Z07:00",
						"2006-01-02 15:04:05",
						"2006-01-02",
						"2006/01/02",
						"01/02/2006",
						"2006-01",
					}
					for _, format := range formats {
						parsedTime, err = time.Parse(format, timeStr)
						if err == nil {
							break
						}
					}
				}
				if err == nil {
					gofeedItem.PublishedParsed = &parsedTime
				}
			}
		}
	}

	// Extract thumbnail
	if feed.XPathItemThumbnail != "" {
		if thumbNode := htmlquery.FindOne(item, feed.XPathItemThumbnail); thumbNode != nil {
			var imageURL string
			// Check if it's an img src or just text
			if thumbNode.Data == "img" {
				for _, attr := range thumbNode.Attr {
					if attr.Key == "src" {
						imageURL = attr.Val
						break
					}
				}
			} else {
				imageURL = strings.TrimSpace(htmlquery.InnerText(thumbNode))
			}
			// Resolve relative URLs to absolute URLs
			if imageURL != "" && !strings.HasPrefix(imageURL, "http") {
				baseURL, err := url.Parse(feed.URL)
				if err == nil {
					if ref, err := url.Parse(imageURL); err == nil {
						imageURL = baseURL.ResolveReference(ref).String()
					}
				}
			}
			if imageURL != "" {
				gofeedItem.Image = &gofeed.Image{URL: imageURL}
			}
		}
	}

	// Extract categories
	if feed.XPathItemCategories != "" {
		categories := htmlquery.Find(item, feed.XPathItemCategories)
		if len(categories) > 0 {
			gofeedItem.Categories = make([]string, 0, len(categories))
			for _, cat := range categories {
				catText := strings.TrimSpace(htmlquery.InnerText(cat))
				if catText != "" {
					gofeedItem.Categories = append(gofeedItem.Categories, catText)
				}
			}
		}
	}

	// Extract UID
	if feed.XPathItemUid != "" {
		if uidNode := htmlquery.FindOne(item, feed.XPathItemUid); uidNode != nil {
			gofeedItem.GUID = strings.TrimSpace(htmlquery.InnerText(uidNode))
		}
	}

	// If no UID, generate one from link or title
	if gofeedItem.GUID == "" {
		if gofeedItem.Link != "" {
			gofeedItem.GUID = gofeedItem.Link
		} else {
			gofeedItem.GUID = gofeedItem.Title
		}
	}

	return gofeedItem
}

// extractItemFromXMLNode extracts a gofeed.Item from an XML node
func (f *Fetcher) extractItemFromXMLNode(item *xmlquery.Node, feed *models.Feed) *gofeed.Item {
	gofeedItem := &gofeed.Item{}

	// Extract title
	if feed.XPathItemTitle != "" {
		if titleNode := xmlquery.FindOne(item, feed.XPathItemTitle); titleNode != nil {
			gofeedItem.Title = strings.TrimSpace(titleNode.InnerText())
		}
	}

	// Extract content
	if feed.XPathItemContent != "" {
		if contentNode := xmlquery.FindOne(item, feed.XPathItemContent); contentNode != nil {
			gofeedItem.Content = contentNode.OutputXML(true)
		}
	}

	// Extract URI
	if feed.XPathItemUri != "" {
		if uriNode := xmlquery.FindOne(item, feed.XPathItemUri); uriNode != nil {
			link := strings.TrimSpace(uriNode.InnerText())
			// Resolve relative URLs to absolute URLs
			if link != "" && !strings.HasPrefix(link, "http") {
				baseURL, err := url.Parse(feed.URL)
				if err == nil {
					if ref, err := url.Parse(link); err == nil {
						gofeedItem.Link = baseURL.ResolveReference(ref).String()
					} else {
						gofeedItem.Link = link
					}
				} else {
					gofeedItem.Link = link
				}
			} else {
				gofeedItem.Link = link
			}
		}
	}

	// If no URI was extracted, generate a unique URL for this article
	// This ensures each XPath article has a unique URL to prevent database conflicts
	if gofeedItem.Link == "" {
		// Use feed URL as base and append a hash of the title or content
		uniqueID := gofeedItem.Title
		if uniqueID == "" {
			// Fallback to content or a timestamp-based ID
			if gofeedItem.Content != "" {
				uniqueID = gofeedItem.Content
			} else {
				uniqueID = fmt.Sprintf("xpath-article-%d", time.Now().UnixNano())
			}
		}
		// Create a simple hash of the unique identifier
		hash := fmt.Sprintf("%x", len(uniqueID)) // Simple length-based hash for uniqueness
		gofeedItem.Link = fmt.Sprintf("%s#xpath-%s", feed.URL, hash)
	}

	// Extract author
	if feed.XPathItemAuthor != "" {
		if authorNode := xmlquery.FindOne(item, feed.XPathItemAuthor); authorNode != nil {
			gofeedItem.Author = &gofeed.Person{
				Name: strings.TrimSpace(authorNode.InnerText()),
			}
		}
	}

	// Extract timestamp
	if feed.XPathItemTimestamp != "" {
		if timeNode := xmlquery.FindOne(item, feed.XPathItemTimestamp); timeNode != nil {
			timeStr := strings.TrimSpace(timeNode.InnerText())
			if timeStr != "" {
				var parsedTime time.Time
				var err error
				if feed.XPathItemTimeFormat != "" {
					parsedTime, err = time.Parse(feed.XPathItemTimeFormat, timeStr)
				} else {
					// Try common formats
					formats := []string{
						time.RFC3339,
						time.RFC1123,
						"2006-01-02T15:04:05Z07:00",
						"2006-01-02 15:04:05",
						"2006-01-02",
					}
					for _, format := range formats {
						parsedTime, err = time.Parse(format, timeStr)
						if err == nil {
							break
						}
					}
				}
				if err == nil {
					gofeedItem.PublishedParsed = &parsedTime
				}
			}
		}
	}

	// Extract thumbnail
	if feed.XPathItemThumbnail != "" {
		if thumbNode := xmlquery.FindOne(item, feed.XPathItemThumbnail); thumbNode != nil {
			var imageURL string
			// For XML, we assume it's text content or attribute
			if thumbNode.Type == xmlquery.ElementNode && len(thumbNode.Attr) > 0 {
				// Check for src attribute
				for _, attr := range thumbNode.Attr {
					if attr.Name.Local == "src" || attr.Name.Local == "href" {
						imageURL = attr.Value
						break
					}
				}
			} else {
				imageURL = strings.TrimSpace(thumbNode.InnerText())
			}
			// Resolve relative URLs to absolute URLs
			if imageURL != "" && !strings.HasPrefix(imageURL, "http") {
				baseURL, err := url.Parse(feed.URL)
				if err == nil {
					if ref, err := url.Parse(imageURL); err == nil {
						imageURL = baseURL.ResolveReference(ref).String()
					}
				}
			}
			if imageURL != "" {
				gofeedItem.Image = &gofeed.Image{URL: imageURL}
			}
		}
	}

	// Extract categories
	if feed.XPathItemCategories != "" {
		categories := xmlquery.Find(item, feed.XPathItemCategories)
		if len(categories) > 0 {
			gofeedItem.Categories = make([]string, 0, len(categories))
			for _, cat := range categories {
				catText := strings.TrimSpace(cat.InnerText())
				if catText != "" {
					gofeedItem.Categories = append(gofeedItem.Categories, catText)
				}
			}
		}
	}

	// Extract UID
	if feed.XPathItemUid != "" {
		if uidNode := xmlquery.FindOne(item, feed.XPathItemUid); uidNode != nil {
			gofeedItem.GUID = strings.TrimSpace(uidNode.InnerText())
		}
	}

	// If no UID, generate one from link or title
	if gofeedItem.GUID == "" {
		if gofeedItem.Link != "" {
			gofeedItem.GUID = gofeedItem.Link
		} else {
			gofeedItem.GUID = gofeedItem.Title
		}
	}

	return gofeedItem
}

// ScriptError represents an error related to script execution
type ScriptError struct {
	Message string
}

func (e *ScriptError) Error() string {
	return e.Message
}
