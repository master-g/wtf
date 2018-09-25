package main

import (
	"bufio"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/htmlindex"
)

// Query represent a word query
type Query struct {
	Words []string
	Lang  string
}

// Pronounce structure
type Pronounce struct {
	Name     string
	Phonetic string
}

func (p Pronounce) String() string {
	return p.Name + p.Phonetic
}

// Result of query the dictionary
type Result struct {
	PronounceList []Pronounce
}

// Engine interface
type Engine interface {
	URL(q Query) string
	Execute(q Query) Result
}

// NewEngine creates and returns a engine implementation of type specified
func NewEngine(e string) Engine {
	switch e {
	case "youdao":
		return YDEngine{}
	case "google":
		return GoogleEngine{}
	default:
		return nil
	}
}

// GoogleEngine for google
type GoogleEngine struct {
}

// URL implementation
func (e GoogleEngine) URL(q Query) string {
	return ""
}

// Execute the query
func (e GoogleEngine) Execute(q Query) (result Result) {
	return
}

// YDEngine for youdao dictionary
type YDEngine struct {
}

// URL implementation
func (e YDEngine) URL(q Query) string {
	return "http://www.youdao.com/w/" + strings.Join(q.Words, "%20")
}

// Execute the query
func (e YDEngine) Execute(q Query) (result Result) {
	res, err := http.Get(e.URL(q))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code: %d %s", res.StatusCode, res.Status)
	}

	// convert to utf-8
	utfBody, err := decodeHTMLBody(res.Body, "utf-8")
	if err != nil {
		log.Fatalf("unable to convert html to utf-8: %s", err.Error())
	}

	// load html document
	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		log.Fatal(err)
	}

	log.Info(doc)
	// find pronounce
	if len(q.Words) == 1 {
		result.PronounceList = e.extractPronounce(doc)
	}

	return
}

func (e YDEngine) extractPronounce(doc *goquery.Document) (pronounces []Pronounce) {
	doc.Find("div.baav").Each(func(i int, s *goquery.Selection) {
		log.Info(s.Find("span.pronounce").Text())
	})

	return nil
}

func detectContentCharset(body io.Reader) string {
	r := bufio.NewReader(body)
	if data, err := r.Peek(1024); err == nil {
		if _, name, ok := charset.DetermineEncoding(data, ""); ok {
			return name
		}
	}
	return "utf-8"
}

// decodeHTMLBody returns an decoding reader of the html Body for the specified `charset`
// If `charset` is empty, DecodeHTMLBody tries to guess the encoding from the content
func decodeHTMLBody(body io.Reader, charset string) (io.Reader, error) {
	if charset == "" {
		charset = detectContentCharset(body)
	}
	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, err
	}
	if name, _ := htmlindex.Name(e); name != "utf-8" {
		body = e.NewDecoder().Reader(body)
	}
	return body, nil
}
