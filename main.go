package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

func main() {
	var opts = struct {
		Verbose    bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
		Engine     string `short:"e" long:"engine" description:"Search word in specific engine" choice:"youdao" choice:"google" default:"youdao"`
		Language   string `short:"l" long:"lang" description:"destination language" choice:"en" choice:"ch" choice:"jp" choice:"de" default:"ch"`
		Positional struct {
			Words []string `description:"word(s) to search for" required:"yes"`
		} `positional-args:"yes"`
	}{}

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	// verbosity
	if !opts.Verbose {
		log.SetLevel(log.PanicLevel)
	}
	log.Infof("verbosity: %v\n", opts.Verbose)

	// engine
	log.Infof("use engine: %v", opts.Engine)
	e := NewEngine(opts.Engine)

	// language
	log.Infof("language: %v", opts.Language)

	// words
	log.Info(opts.Positional.Words)

	q := Query{
		Lang:  opts.Language,
		Words: opts.Positional.Words,
	}

	log.Infof("words: %v", opts.Positional.Words)

	// work
	log.Infof("query: %v", e.URL(q))
	e.Execute(q)
}
