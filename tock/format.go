package tock

import (
	"fmt"
	"strings"
)

// FormatResponse converts a Response into a plain-text / Markdown string
// suitable for display in an MCP text content block.
func FormatResponse(r *Response) string {
	var sb strings.Builder

	for _, msg := range r.Responses {
		if msg.Text != "" {
			sb.WriteString(msg.Text)
			sb.WriteString("\n")
		}

		if msg.Card != nil {
			formatCard(&sb, msg.Card)
		}

		// Flatten carousels: each card is rendered sequentially.
		if msg.Carousel != nil {
			for _, card := range msg.Carousel.Cards {
				formatCard(&sb, &card)
			}
		}

		for _, btn := range msg.Buttons {
			formatButton(&sb, btn)
		}
	}

	return strings.TrimSpace(sb.String())
}

// formatCard appends a card's title, subtitle, optional file link and buttons to sb.
func formatCard(sb *strings.Builder, card *Card) {
	if card.Title != "" {
		sb.WriteString(fmt.Sprintf("**%s**\n", card.Title))
	}
	if card.SubTitle != "" {
		sb.WriteString(fmt.Sprintf("%s\n", card.SubTitle))
	}
	if card.File != nil && card.File.URL != "" {
		sb.WriteString(fmt.Sprintf(" %s : %s\n", card.File.Name, card.File.URL))
	}
	for _, btn := range card.Buttons {
		formatButton(sb, btn)
	}
}

// formatButton appends a single button line to sb.
// web_url buttons are rendered as Markdown links; others as list items.
func formatButton(sb *strings.Builder, btn Button) {
	switch btn.Type {
	case "web_url":
		sb.WriteString(fmt.Sprintf(" [%s](%s)\n", btn.Title, btn.URL))
	case "postback", "quick_reply":
		sb.WriteString(fmt.Sprintf("▶ %s\n", btn.Title))
	default:
		sb.WriteString(fmt.Sprintf("• %s\n", btn.Title))
	}
}
