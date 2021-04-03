package scrape

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gocolly/colly"
	"regexp"
)

type wordpressResponse struct {
	Content struct {
		Rendered string
	}
	config Config
}

func NewWordpressResponse(config Config) *wordpressResponse {
	return &wordpressResponse{
		config: config,
	}
}

func (w *wordpressResponse) parse(s string) (content string, err error) {
	item := func(s string) string {
		re := regexp.MustCompile(`[0-9]+`)
		return string(re.Find([]byte(s)))
	}(s)
	url := w.config.API + item
	c := colly.NewCollector()
	c.OnRequest(func(r *colly.Request) {
		auth := base64.StdEncoding.EncodeToString([]byte(w.config.Login + ":" + w.config.Password))
		r.Method = "GET"
		r.Headers.Set("Authorization", "Basic"+auth)
		r.Headers.Set("Content-Type", "application/json")
	})
	c.OnResponse(func(r *colly.Response) {
		if e := json.Unmarshal(r.Body, &w); err != nil {
			err = e
		}
	})
	if err != nil {
		return "", err
	}
	if err := c.Visit(url); err != nil {
		return "", err
	}
	return w.Content.Rendered, nil
}
