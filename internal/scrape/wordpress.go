package scrape

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/pkg/errors"

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

func NewWordpressResponse(config Config) Parser {
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
	if item == "" {
		return "", errors.New("invalid URL")
	}

	url := w.config.API + item
	return w.Content.Rendered, w.visit(url)
}

func (w *wordpressResponse) visit(s string) error {
	w.collector.OnRequest(func(r *colly.Request) {
		r.Headers.Add("Authorization", w.authHeaderValue)
	})

	if err := func() (err error) {
		w.collector.OnResponse(func(r *colly.Response) {
			err = json.Unmarshal(r.Body, &w)
		})
		return
	}(); err != nil {
		return err
	}

	return errors.Wrap(w.collector.Visit(s), "on visit")
}
