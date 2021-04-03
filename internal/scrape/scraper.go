package scrape

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Scraper struct {
	config *Config
}

type Scrape interface {
	parse(string) (string, error)
}

func New(config *Config) *Scraper {
	return &Scraper{
		config: config,
	}
}

func (s *Scraper) GetCountSymbols(url string) (int, error) {
	get := map[string]Scrape{
		s.config.Medium:  &mediumResponse{},
		s.config.WebSite: NewWordpressResponse(*s.config),
	}
	rawData, err := parse(get[func(url string) string {
		line := strings.Split(url, "://")
		return strings.Split(line[1], ".")[0]
	}(url)], url)
	if err != nil {
		return -1, err
	}
	text, err := s.getCyrillicText(rawData)
	if err != nil {
		return -1, err
	}
	return utf8.RuneCountInString(text), nil
}

func parse(s Scrape, url string) (string, error) {
	content, err := s.parse(url)
	if err != nil {
		return "", err
	}
	return content, nil
}

func (s *Scraper) getCyrillicText(content string) (string, error) {
	re, err := regexp.Compile(`\p{Cyrillic}`)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга: %v\n", err)
	}
	var text string
	for _, t := range re.FindAllString(content, -1) {
		text += t
	}
	return text, nil
}
