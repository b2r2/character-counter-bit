package scrape

// Scraper ...
type Scraper struct {
	config *Config
}

// MediumResponse ...
type MediumResponse struct {
	Payload struct {
		Value struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Content struct {
				Subtitle string `json:"subtitle"`
				Article  struct {
					Paragraphs []Data `json:"paragraphs"`
				} `json:"bodyModel"`
			} `json:"content"`
		} `json:"value"`
	} `json:"payload"`
}

type Data struct {
	Text string `json:"text,omitempty"`
}

// WPResponse ...
type WPResponse struct {
	Content struct {
		Rendered string
	}
}
