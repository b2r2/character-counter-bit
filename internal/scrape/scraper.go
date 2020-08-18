package scrape

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gocolly/colly"
)

const (
	// MEDIUM ...
	MEDIUM = "medium"
)

// New ...
func New(config *Config) *Scraper {
	return &Scraper{
		config: config,
	}
}

// GetCountSymbols ...
func (s *Scraper) GetCountSymbols(url string) (int, error) {
	caller := map[string]func(string) string{
		MEDIUM:           s.parseMedium,
		s.config.WebSite: s.parseWp,
	}
	name := func(url string) string {
		line := strings.Split(url, "://")
		return strings.Split(line[1], ".")[0]
	}(url)

	raw := caller[name](url)
	text, err := s.getCyrillicText(raw)
	if err != nil {
		return 0, err
	}
	return utf8.RuneCountInString(text), nil
}

func (s *Scraper) parseMedium(url string) string {
	if re := regexp.MustCompile("edit$"); re.MatchString(url) {
		url = re.ReplaceAllString(url, "")
	}

	var mr MediumResponse
	c := colly.NewCollector()
	c.OnRequest(func(r *colly.Request) {
		r.Method = "GET"
		r.Headers.Add("Accept", "application/json")
	})
	c.OnResponse(func(r *colly.Response) {
		json.Unmarshal(r.Body[16:], &mr)
	})
	c.Visit(url)
	return handleBody(mr)
}

func (s *Scraper) parseWp(url string) string {
	item := handleLink(url)
	url = s.config.API + item

	var wr WPResponse
	c := colly.NewCollector()
	c.OnRequest(func(r *colly.Request) {
		auth := base64.StdEncoding.EncodeToString([]byte(s.config.Login + ":" + s.config.Password))
		r.Method = "GET"
		r.Headers.Set("Authorization", "Basic "+auth)
		r.Headers.Set("Content-Type", "application/json")
	})
	c.OnResponse(func(r *colly.Response) {
		json.Unmarshal(r.Body, &wr)
	})
	c.Visit(url)
	return handleBody(wr)
}

func handleBody(v interface{}) string {
	var t string
	switch b := v.(type) {
	case MediumResponse:
		for _, ps := range b.Payload.Value.Content.Article.Paragraphs {
			for _, p := range ps.Text {
				t += string(p)
			}
		}
	case WPResponse:
		t = b.Content.Rendered
	}
	return t
}

func (s *Scraper) getCyrillicText(cnt string) (string, error) {
	re, err := regexp.Compile("\\p{Cyrillic}")
	if err != nil {
		return "", fmt.Errorf("Ошибка парсинга")
	}
	tmp := re.FindAllString(cnt, -1)
	var total string
	for _, t := range tmp {
		total += t
	}
	return total, nil
}

func handleLink(s string) string {
	re := regexp.MustCompile(`[0-9]+`)
	return string(re.Find([]byte(s)))
}
