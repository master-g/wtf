package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/master-g/wtf/pkg/engine"
	log "github.com/sirupsen/logrus"
)

func main() {
	var opts = struct {
		Verbose      bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
		Engine       string `short:"e" long:"engine" description:"Search word in specific engine" choice:"youdao" choice:"google" default:"youdao"`
		Language     string `short:"l" long:"lang" description:"destination language" choice:"eng" choice:"chs" choice:"jap" default:"chs"`
		WebTranslate bool   `short:"w" long:"web" description:"enable web translate based on website data"`
		Positional   struct {
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
	log.Infof("verbosity: %v", opts.Verbose)

	// engine
	log.Infof("use engine: %v", opts.Engine)
	e := engine.NewEngine(opts.Engine)

	// language
	log.Infof("language: %v", opts.Language)

	// words
	q := Query{
		Lang:     opts.Language,
		Words:    opts.Positional.Words,
		WebTrans: opts.WebTranslate,
	}

	log.Infof("words: %v", opts.Positional.Words)

	// work
	log.Infof("query: %v", e.URL(q))
	r := e.Execute(q)
	fmt.Println(r.String())
}
