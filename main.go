package main

import (
	"crawler/crawler"
)

func main() {
	crawler.Start("https://www.ebay.com/sch/garlandcomputer/m.html", "all")
}
