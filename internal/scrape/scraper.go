package scrape

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gocolly/colly"
)

const (
	// MEDIUM ...
	MEDIUM = "medium"
)

// Scraper ...
type Scraper struct {
	config *Config
}

// WPResponse ...
type WPResponse struct {
	Content struct {
		Rendered string
	}
}

// New ...
func New(config *Config) *Scraper {
	return &Scraper{
		config: config,
	}
}

// GetCountSymbols ...
func (s *Scraper) GetCountSymbols(url string) (int, error) {
	var content string
	var err error
	callerScraper := map[string]func(string) (string, error){
		MEDIUM:           s.scrapeMedium,
		s.config.WebSite: s.scrapeSite,
	}
	d := func(url string) string {
		line := strings.Split(url, "://")
		return strings.Split(line[1], ".")[0]
	}(url)
	if content, err = callerScraper[d](url); err != nil {
		return 0, err
	}
	text, err := s.parse(content)
	if err != nil {
		return 0, err
	}
	size := utf8.RuneCountInString(text)
	return size, nil
}

func (s *Scraper) scrapeMedium(url string) (string, error) {
	c := colly.NewCollector(
		colly.Async(true),
	)
	if comp := regexp.MustCompile("edit$"); comp.MatchString(url) {
		url = comp.ReplaceAllString(url, "")
	}
	var querySelectors []string = []string{`article`, "section"}
	var text string
	var querySelector string = querySelectors[0]
	c.OnHTML(querySelector, func(e *colly.HTMLElement) {
		tag := querySelectors[1]
		text = e.ChildText(tag)
	})
	c.Visit(url)
	c.Wait()

	if text == "" {
		return "", fmt.Errorf("Ошибка скрэппинга")
	}
	return text, nil
}

func (s *Scraper) scrapeSite(url string) (string, error) {
	var number string
	re := regexp.MustCompile(`[0-9]+`)
	if re.MatchString(url) {
		number = string(re.Find([]byte(url)))
	}
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", s.config.API+number, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(s.config.Login, s.config.Password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var wordpressresponse WPResponse
	if err := json.Unmarshal(data, &wordpressresponse); err != nil {
		return "", fmt.Errorf("Не удалось извлечь данные из сайта")
	}
	return wordpressresponse.Content.Rendered, nil
}

func (s *Scraper) parse(cnt string) (string, error) {
	re, err := regexp.Compile("\\p{Cyrillic}")
	if err != nil {
		return "", fmt.Errorf("Ошибка парсинга")
	}
	tmp := re.FindAllString(cnt, -1)
	var tot string
	for _, t := range tmp {
		tot += t
	}
	return tot, nil
}
