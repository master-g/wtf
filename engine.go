package main

import (
	"bufio"
	"fmt"
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
	Words    []string
	Lang     string
	WebTrans bool
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
	Title         string
	PronounceList []Pronounce
	Translates    []string
	WebTranslates []string
	WebPhrases    []string
	Origin        string
}

func (r Result) String() string {
	builder := &strings.Builder{}
	builder.WriteString("-----------------------\n")
	builder.WriteString(r.Title)
	builder.WriteString("\n-----------------------\n")
	if len(r.PronounceList) > 0 {
		for _, p := range r.PronounceList {
			builder.WriteString(p.String())
			builder.WriteString("\n")
		}
		builder.WriteString("-----------------------\n")
	}
	if len(r.Translates) > 0 {
		for _, t := range r.Translates {
			builder.WriteString(t)
			builder.WriteString("\n")
		}
		builder.WriteString("-----------------------\n")
	}
	if len(r.WebTranslates) > 0 {
		builder.WriteString("Web Translations\n----\n")
		for _, t := range r.WebTranslates {
			builder.WriteString(t)
		}
		builder.WriteString("-----------------------\n")
	}
	if len(r.WebPhrases) > 0 {
		builder.WriteString("Web Phrases\n----\n")
		for _, t := range r.WebPhrases {
			builder.WriteString(t)
		}
		builder.WriteString("-----------------------\n")
	}
	if r.Origin != "" {
		builder.WriteString(r.Origin)
		builder.WriteString("\n-----------------------")
	}
	return builder.String()
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
	l := ""
	if q.Lang != "chs" {
		l = q.Lang + "/"
	}
	return "http://www.youdao.com/w/" + l + strings.Join(q.Words, "%20")
}

// Execute the query
func (e YDEngine) Execute(q Query) (result Result) {
	result.Title = strings.Join(q.Words, " ")
	result.Origin = e.URL(q)

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

	if q.Lang != "chs" {
		switch q.Lang {
		case "eng":
			result.Translates = e.extractEngTranslate(doc)
		case "jap":
			result.Translates = e.extractJapTranslate(doc)
		}
	} else if len(q.Words) == 1 {
		result.PronounceList = e.extractPronounce(doc)
		result.Translates = e.extractTranslate(doc)
	}

	if q.WebTrans {
		result.WebTranslates = e.extractWebTranslate(doc)
		result.WebPhrases = e.extractWebPhrase(doc)
	}

	return
}

func (e YDEngine) extractPronounce(doc *goquery.Document) (pronounces []Pronounce) {
	doc.Find("div.baav").Each(func(i int, s *goquery.Selection) {
		s.Find("span.pronounce").Each(func(j int, ss *goquery.Selection) {
			str := strings.TrimSpace(ss.Text())
			scanner := bufio.NewScanner(strings.NewReader(str))
			pronounce := Pronounce{}
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.Index(line, "[") != -1 {
					pronounce.Phonetic = line
					pronounces = append(pronounces, pronounce)
				} else {
					pronounce.Name = line
				}
			}
			if err := scanner.Err(); err != nil {
				log.Error("unable to scan pronounce: ", err.Error())
				return
			}
		})
	})

	return
}

func (e YDEngine) extractTranslate(doc *goquery.Document) (translates []string) {
	doc.Find("div#phrsListTab ul li").Each(func(i int, s *goquery.Selection) {
		translates = append(translates, s.Text())
	})

	return
}

func (e YDEngine) extractEngTranslate(doc *goquery.Document) (translates []string) {
	builder := &strings.Builder{}
	doc.Find("div#phrsListTab div.trans-container ul p.wordGroup").Each(func(i int, s *goquery.Selection) {
		elSpans := s.Find("span")
		if elSpans == nil {
			return
		}
		elFirst := elSpans.First()
		if elFirst == nil {
			return
		}
		first := strings.TrimSpace(elFirst.Text())
		titles := make([]string, 0)
		s.Find("span.contentTitle").Each(func(j int, ss *goquery.Selection) {
			title := ss.Find("a")
			if title != nil {
				titles = append(titles, strings.TrimSpace(title.Text()))
			}
		})
		builder.WriteString(fmt.Sprintf("%-8s %s", first, strings.Join(titles, "; ")))
		translates = append(translates, builder.String())
		builder.Reset()
	})

	return
}

func (e YDEngine) extractJapTranslate(doc *goquery.Document) (translates []string) {
	elContainer := doc.Find("div#results-contents .trans-container").First()
	if elContainer == nil {
		return
	}

	// ol
	elContainer.Find("ul.ol li").Each(func(i int, s *goquery.Selection) {
		t := s.Find("p.sense-title")
		if t != nil {
			str := strings.TrimSpace(t.Text())
			if len(str) > 0 {
				translates = append(translates, fmt.Sprintf("%d. %s", i+1, str))
			}
		}
	})

	// ul
	elContainer.Find("ul.ul>li").Each(func(i int, s *goquery.Selection) {
		elTitle := s.Find("p.sense-title")
		if elTitle != nil {
			translates = append(translates, strings.TrimSpace(elTitle.Text()))
		}
		elExample := s.Find("ul.sense-ex")
		if elExample != nil {
			elExample.Find("li").Each(func(j int, ss *goquery.Selection) {
				exTitle := ss.Find("p").First()
				if exTitle == nil {
					return
				}
				str := strings.TrimSpace(exTitle.Text())
				scanner := bufio.NewScanner(strings.NewReader(str))
				titleComp := make([]string, 0)
				for scanner.Scan() {
					content := strings.TrimSpace(scanner.Text())
					if len(content) > 0 {
						titleComp = append(titleComp, content)
					}
				}
				if err := scanner.Err(); err != nil {
					log.Error("unable to scan example: ", err.Error())
					return
				}
				translates = append(translates, strings.Join(titleComp, ""))
				exContent := ss.Find("p.exam-sen")
				if exContent == nil {
					return
				}
				translates = append(translates, "     "+strings.TrimSpace(exContent.Text()))
			})
		}
	})
	return
}

func (e YDEngine) extractWebTranslate(doc *goquery.Document) (translates []string) {
	doc.Find("div#tWebTrans div.wt-container").Each(func(i int, s *goquery.Selection) {
		title := s.Find("div.title span")
		if title == nil {
			return
		}
		content := s.Find("p.collapse-content")
		if content == nil {
			return
		}
		if first := content.First(); first != nil {
			builder := &strings.Builder{}
			builder.WriteString(strings.TrimSpace(title.Text()))
			builder.WriteString("\n")
			str := strings.TrimSpace(first.Text())
			scanner := bufio.NewScanner(strings.NewReader(str))
			for scanner.Scan() {
				builder.WriteString("    ")
				builder.WriteString(strings.TrimSpace(scanner.Text()))
				builder.WriteString("\n")
			}
			if err := scanner.Err(); err != nil {
				log.Error("unable to scan web translate: ", err.Error())
				return
			}
			translates = append(translates, builder.String())
		}
	})
	return
}

func (e YDEngine) extractWebPhrase(doc *goquery.Document) (translates []string) {
	builder := &strings.Builder{}
	doc.Find("div#webPhrase p.wordGroup").Each(func(i int, s *goquery.Selection) {
		content := strings.TrimSpace(s.Text())
		scanner := bufio.NewScanner(strings.NewReader(content))
		usages := make([]string, 0)
		for scanner.Scan() {
			usage := strings.TrimSpace(scanner.Text())
			if usage != "" && len(usage) > 1 {
				usages = append(usages, usage)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Error("unable to scan web phrase: ", err.Error())
			return
		}
		if len(usages) <= 1 {
			return
		}
		builder.WriteString(fmt.Sprintf("%-8s %s\n", usages[0], strings.Join(usages[1:], "; ")))
		translates = append(translates, builder.String())
		builder.Reset()
	})
	return
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
