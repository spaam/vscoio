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
	var fastupdate bool
	flag.BoolVar(&nometadata, "no-metadata", false, "Dont create metadata files")
	flag.BoolVar(&fastupdate, "fast-update", false, "Stop downloading files when we detect existing one")
	flag.IntVar(&threads, "threads", 10, "Number of concurrent requests")
	flag.Parse()
	profiles := flag.Args()

	if len(profiles) == 0 {
		fmt.Println("vscoio: error: need profiles as arguments")
		return
	}
	fmt.Println("scrape profiles for pictures")
	users := collect.Scrape(fastupdate, profiles, threads)

	total := download.Download(nometadata, fastupdate, threads, users)
	fmt.Printf("Total new pictures: %d\n", total)
}
