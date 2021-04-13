package scrape

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/gocolly/colly"
)

type wordpressResponse struct {
	Content struct {
		Rendered string
	}
	config          Config
	collector       *colly.Collector
	authHeaderValue string
}

func NewWordpressResponse(config Config) Scrape {
	return &wordpressResponse{
		config:          config,
		collector:       colly.NewCollector(),
		authHeaderValue: "Basic " + basicAuth(config.Login, config.Password),
	}
}

func basicAuth(login, password string) string {
	auth := fmt.Sprintf("%s:%s", login, password)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (w *wordpressResponse) parse(s string) (string, error) {
	item := func(s string) string {
		re := regexp.MustCompile(`[0-9]+`)
		return string(re.Find([]byte(s)))
	}(s)
	url := w.config.API + item
	w.collector.OnRequest(func(r *colly.Request) {
		r.Headers.Add("Authorization", w.authHeaderValue)
	})

	var stage bool
	w.collector.OnResponse(func(r *colly.Response) {
		if err := json.Unmarshal(r.Body, &w); err != nil {
			stage = true
		}
	})
	if stage {
		return "", ErrUnmarshal
	}
	if err := w.collector.Visit(url); err != nil {
		return "", err
	}
	return w.Content.Rendered, nil
}
