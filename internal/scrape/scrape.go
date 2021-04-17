package scrape

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Parser interface {
	parse(string) (string, error)
}

type Scrape struct {
	config *Config
}

func New(config *Config) *Scrape {
	return &Scrape{
		config: config,
	}
}

func (s *Scrape) GetCountSymbols(url string) (int, error) {
	get := map[string]Parser{
		s.config.Medium:  NewMediumResponse(),
		s.config.WebSite: NewWordpressResponse(*s.config),
	}
	rawData, err := parse(get[func(url string) string {
		line := strings.Split(url, "://")
		return strings.Split(line[1], ".")[0]
	}(url)], url)
	if err != nil {
		return -1, fmt.Errorf("parsing error: %w\n", err)
	}
	text, err := s.getCyrillicText(rawData)
	if err != nil {
		return -1, fmt.Errorf("get cyrillic error: %w\n", err)
	}
	return utf8.RuneCountInString(text), nil
}

func (s *Scrape) getCyrillicText(content string) (string, error) {
	re, err := regexp.Compile(`\p{Cyrillic}`)
	if err != nil {
		return "", err
	}
	var text string
	for _, t := range re.FindAllString(content, -1) {
		text += t
	}
	return text, nil
}

func parse(p Parser, url string) (string, error) {
	return p.parse(url)
}
