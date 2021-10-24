package main

import (
	"flag"
	"fmt"

	"github.com/spaam/vscoio/internal/collect"
	"github.com/spaam/vscoio/internal/download"
)

func main() {
	var nometadata bool
	flag.BoolVar(&nometadata, "no-metadata", false, "Dont create metadata files")
	flag.Parse()
	profiles := flag.Args()

	if len(profiles) == 0 {
		fmt.Println("vscoio: error: need profiles as arguments")
	}
	users := collect.Scrape(profiles)

	download.Download(nometadata, users)
}
