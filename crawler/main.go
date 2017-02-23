package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/gocrawl"
	"github.com/PuerkitoBio/goquery"
)

type Ext struct {
	*gocrawl.DefaultExtender
	PageChannel chan Page
}

type Page struct {
	Url string
	Body string
}

func (e *Ext) Visit(ctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) (interface{}, bool) {
	fmt.Printf("Visit: %s\n", ctx.URL())
	url := fmt.Sprintf("%s", ctx.URL())
	body := fmt.Sprintf("%s", res.Body)
	e.PageChannel <- Page{url, body}
	return nil, true
}

func (e *Ext) Filter(ctx *gocrawl.URLContext, isVisited bool) bool {
	if isVisited {
		return false
	}
	if ctx.URL().Host == "github.com" || ctx.URL().Host == "golang.org" || ctx.URL().Host == "0value.com" {
		return true
	}
	return false
}

func main() {
	pagechan := make(chan Page)
	//crawl
	ext := &Ext{&gocrawl.DefaultExtender{}, pagechan}
	// Set custom options
	opts := gocrawl.NewOptions(ext)
	opts.CrawlDelay = 1 * time.Second
	opts.LogFlags = gocrawl.LogError
	opts.SameHostOnly = false
	opts.MaxVisits = 100

	c := gocrawl.NewCrawlerWithOptions(opts)
	go c.Run("http://0value.com")

	for p := range pagechan {
		fmt.Println("from channel: ", p.Url)
	}
}
