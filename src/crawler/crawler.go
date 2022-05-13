package crawler

import (
	"golang.org/x/net/html"
	"log"
	"tjweldon/spider/src/util"
)

type FilteredScraper struct {
	Filter NodeFilter
	Scrape NodeScraper
}

type Crawler struct {
	Scrapers []FilteredScraper
	Root     *html.Node
	Done     chan Signal
	Ready    bool
}

func NewCrawler() *Crawler {
	done := make(chan Signal)
	return &Crawler{
		Done:  done,
		Ready: true,
	}
}

func (c *Crawler) Scrape(node *html.Node) {
	for _, scraper := range c.Scrapers {
		if scraper.Filter(node) {
			scraper.Scrape(node)
		}
	}
}

func (c *Crawler) AddScraper(s NodeScraper, f NodeFilter) *Crawler {
	c.Scrapers = append(
		c.Scrapers, FilteredScraper{Scrape: s, Filter: f},
	)

	return c
}

// CrawlNow is a blocking recursive walk over the node tree. Each node is passed
// to the configured Scrapers.
func (c *Crawler) CrawlNow(target string) {
	var f NodeScraper
	f = func(n *html.Node) {
		c.Scrape(n)
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			f(child)
		}
	}
	f(c.GetNodeTree(target))
}

type Signal struct{}

// Crawl is non-blocking. Will report completion on the chan Signal
// passed if not nil.
func (c *Crawler) Crawl(target string) {
	log.Println("Beginning crawl...")
	go func(d chan Signal, t string) {
		c.CrawlNow(t)
		d <- Signal{}
	}(c.Done, target)
}

func (c *Crawler) GetNodeTree(target string) *html.Node {
	resp := util.MustGet(target)
	parentNode, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	c.Root = parentNode

	return parentNode
}

func (c *Crawler) Die() {
	close(c.Done)
}
