package main

import (
	"flag"
	"fmt"

	"github.com/spaam/vscoio/internal/collect"
	"github.com/spaam/vscoio/internal/download"
)

func main() {
	var nometadata bool
	var threads int
	flag.BoolVar(&nometadata, "no-metadata", false, "Dont create metadata files")
	flag.IntVar(&threads, "threads", 10, "Number of concurrent requests")
	flag.Parse()
	profiles := flag.Args()

	if len(profiles) == 0 {
		fmt.Println("vscoio: error: need profiles as arguments")
	}
	fmt.Println("scrape profiles for pictures")
	users := collect.Scrape(profiles)

	download.Download(nometadata, threads, users)
}
