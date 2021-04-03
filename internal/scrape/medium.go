package scrape

import (
	"encoding/json"
	"github.com/gocolly/colly"
	"regexp"
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
}

type Data struct {
	Text string `json:"text,omitempty"`
}

func (m *mediumResponse) parse(s string) (content string, err error) {
	if re := regexp.MustCompile(`edit$`); re.MatchString(s) {
		s = re.ReplaceAllString(s, "")
	}
	c := colly.NewCollector()
	c.OnRequest(func(r *colly.Request) {
		r.Method = "GET"
		r.Headers.Add("Accept", "application/json")
	})
	c.OnResponse(func(r *colly.Response) {
		if e := json.Unmarshal(r.Body[16:], &m); e != nil {
			err = e
		}
	})
	if err != nil {
		return "", err
	}
	if err := c.Visit(s); err != nil {
		return "", err
	}
	for _, ps := range m.Payload.Value.Content.Article.Paragraphs {
		for _, p := range ps.Text {
			content += string(p)
		}
	}
	return content, nil
}
