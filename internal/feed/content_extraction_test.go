package feed

import (
	"testing"

	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
)

// TestExtractContentWithPriority tests the content extraction logic
// ensuring it follows the correct priority order
func TestExtractContentWithPriority(t *testing.T) {
	tests := []struct {
		name        string
		item        *gofeed.Item
		expected    string
		description string
	}{
		{
			name: "Content:encoded takes priority over description",
			item: &gofeed.Item{
				Description: "Short summary",
				Content:     "<p>Full article content from content:encoded</p>",
			},
			expected:    "<p>Full article content from content:encoded</p>",
			description: "When both description and content:encoded exist, should use content:encoded",
		},
		{
			name: "Media:description takes priority over content:encoded",
			item: &gofeed.Item{
				Description: "Short summary",
				Content:     "<p>Content field</p>",
				Extensions: ext.Extensions{
					"media": {
						"group": []ext.Extension{
							{
								Name:  "group",
								Value: "",
								Children: map[string][]ext.Extension{
									"description": {
										{
											Name:  "description",
											Value: "Media description from YouTube feed",
										},
									},
								},
							},
						},
					},
				},
			},
			expected:    "Media description from YouTube feed",
			description: "Media RSS description should take highest priority (for YouTube feeds)",
		},
		{
			name: "Falls back to description when no content:encoded",
			item: &gofeed.Item{
				Description: "Only description available",
				Content:     "",
			},
			expected:    "Only description available",
			description: "Should fall back to description when content:encoded is empty",
		},
		{
			name: "Direct media:description structure",
			item: &gofeed.Item{
				Description: "Short summary",
				Content:     "Some content",
				Extensions: ext.Extensions{
					"media": {
						"description": []ext.Extension{
							{
								Name:  "description",
								Value: "Direct media description",
							},
						},
					},
				},
			},
			expected:    "Direct media description",
			description: "Should handle direct media:description structure",
		},
		{
			name: "Empty article",
			item: &gofeed.Item{
				Description: "",
				Content:     "",
			},
			expected:    "",
			description: "Should return empty string when no content available",
		},
		{
			name: "Only content:encoded available",
			item: &gofeed.Item{
				Description: "",
				Content:     "<p>Only content:encoded exists</p>",
			},
			expected:    "<p>Only content:encoded exists</p>",
			description: "Should use content:encoded when description is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractContent(tt.item)
			if result != tt.expected {
				t.Errorf("ExtractContent() = %q, want %q\nDescription: %s",
					result, tt.expected, tt.description)
			}
		})
	}
}

// TestRealWorldRSSExamples tests with actual RSS feed examples from the issue
func TestRealWorldRSSExamples(t *testing.T) {
	tests := []struct {
		name     string
		item     *gofeed.Item
		hasContent bool
		description string
	}{
		{
			name: "RSS with description only (Pranav Gade's blog)",
			item: &gofeed.Item{
				Description: "I love compilers. They make modern software possible, by taking source code humans can work with, and transforming it into code computers can actually execute.",
				Content:     "",
			},
			hasContent: true,
			description: "When only description exists, it should be used as content",
		},
		{
			name: "RSS with both description and content:encoded (Chinese blog)",
			item: &gofeed.Item{
				Description: "<h2>医疗之困</h2><p>时至今日，中文社交平台上仍流传着大量关于加拿大医疗体系的吐槽...</p>",
				Content:     "<h2>医疗之困</h2><p>时至今日，中文社交平台上仍流传着大量关于加拿大医疗体系的吐槽，核心槽点就是一个字：「慢」。各种「人没了，号还没排到」的故事广为流传...</p><p>Full article continues here...</p>",
			},
			hasContent: true,
			description: "When both exist, should prefer content:encoded (longer full content)",
		},
		{
			name: "Atom feed with content (Jaskey's blog)",
			item: &gofeed.Item{
				Description: "<p>在真实的业务场景中，我们业务的数据——例如订单、会员、支付等——都是持久化到数据库中的...</p>",
				Content:     "<p>在真实的业务场景中，我们业务的数据——例如订单、会员、支付等——都是持久化到数据库中的，因为数据库能有很好的事务保证、持久化保证。但是，正因为数据库要能够满足这么多优秀的功能特性...</p><h2>缓存的意义</h2><p>所谓缓存，实际上就是用空间换时间...</p>",
			},
			hasContent: true,
			description: "Atom feeds with <content> should extract the full content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractContent(tt.item)
			
			if tt.hasContent && result == "" {
				t.Errorf("ExtractContent() returned empty string, but expected content\nDescription: %s", tt.description)
			}
			
			// When both description and content exist, should prefer content (longer)
			if tt.item.Content != "" && tt.item.Description != "" {
				if result == tt.item.Description && result != tt.item.Content {
					t.Errorf("ExtractContent() returned description instead of content:encoded\nDescription: %s", tt.description)
				}
			}
		})
	}
}
