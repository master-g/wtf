package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

func main() {
	var opts struct {
		Verbose bool `short:"v" long:"verbose" description:"Show verbose debug information"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	log.Infof("verbosity: %v\n", opts.Verbose)
}
