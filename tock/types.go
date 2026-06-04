package tock

// Query is the request payload sent to the Tock web-connector.
type Query struct {
	Query  string `json:"query"`  // Natural-language question from the user.
	UserID string `json:"userId"` // Identifier of the caller passed to Tock.
}

// Button represents a clickable action returned by Tock (link, postback, quick-reply).
type Button struct {
	Title   string `json:"title"`             // Display label shown to the user.
	Payload string `json:"payload,omitempty"` // Postback payload (postback/quick_reply buttons).
	URL     string `json:"url,omitempty"`     // Target URL (web_url buttons).
	Type    string `json:"type"`              // Button type: "web_url", "postback", or "quick_reply".
}

// File holds metadata for a file or image attachment returned by Tock.
type File struct {
	URL  string `json:"url"`  // Public URL of the file.
	Name string `json:"name"` // Display name of the file.
	Type string `json:"type"` // MIME type or file category.
}

// Card is a rich-content card (title, subtitle, optional image and buttons).
type Card struct {
	Title    string   `json:"title,omitempty"`
	SubTitle string   `json:"subTitle,omitempty"`
	File     *File    `json:"file,omitempty"`
	Buttons  []Button `json:"buttons,omitempty"`
}

// Carousel wraps a list of Card items returned as a carousel.
type Carousel struct {
	Cards []Card `json:"cards,omitempty"`
}

// Message is a single message element within a Tock response.
// A response can mix plain text, cards, carousels, and standalone buttons.
type Message struct {
	Text     string    `json:"text,omitempty"`
	Buttons  []Button  `json:"buttons,omitempty"`
	Card     *Card     `json:"card,omitempty"`
	Carousel *Carousel `json:"carousel,omitempty"`
}

// Response is the top-level payload returned by the Tock web-connector.
type Response struct {
	Responses []Message `json:"responses"` // Ordered list of message elements.
}
