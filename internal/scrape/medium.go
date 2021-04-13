package scrape

import (
	"encoding/json"
	"regexp"

	"github.com/gocolly/colly"
)

type mediumResponse struct {
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
	collector *colly.Collector
}

type Data struct {
	Text string `json:"text,omitempty"`
}

func NewMediumResponse() Scrape {
	return &mediumResponse{
		collector: colly.NewCollector(),
	}
}

func (m *mediumResponse) parse(s string) (string, error) {
	if re := regexp.MustCompile(`edit$`); re.MatchString(s) {
		s = re.ReplaceAllString(s, "")
	}
	m.collector.OnRequest(func(r *colly.Request) {
		r.Method = "GET"
		r.Headers.Add("Accept", "application/json")
	})
	var stage bool
	m.collector.OnResponse(func(r *colly.Response) {
		if err := json.Unmarshal(r.Body[16:], &m); err != nil {
			stage = true
		}
	})
	if stage {
		return "", ErrUnmarshal
	}
	if err := m.collector.Visit(s); err != nil {
		return "", err
	}
	var content string
	for _, ps := range m.Payload.Value.Content.Article.Paragraphs {
		for _, p := range ps.Text {
			content += string(p)
		}
	}
	return content, nil
}
