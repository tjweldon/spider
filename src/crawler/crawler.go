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
	Target   string
	Scrapers []FilteredScraper
	Root     *html.Node
	Done     chan Signal
}

func NewCrawler(target string) *Crawler {
	return &Crawler{
		Target: target,
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

type Signal struct{}

// Crawl is non-blocking. Will report completion on the chan Signal
// passed if not nil.
func (c *Crawler) Crawl(done chan Signal) {
	log.Println("Beginning crawl...")
	if done != nil {
		c.Done = done
		go func() {
			c.CrawlNow()
			c.Done <- Signal{}
		}()
	} else {
		go c.CrawlNow()
	}
}

func (c *Crawler) GetNodeTree() *html.Node {
	resp := util.MustGet(c.Target)
	parentNode, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	c.Root = parentNode

	return parentNode
}

func (c *Crawler) CrawlNow() {
	var f NodeScraper
	f = func(n *html.Node) {
		c.Scrape(n)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(c.GetNodeTree())
}
