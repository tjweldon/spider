package main

import (
	"github.com/alexflint/go-arg"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"os"
)

var args struct {
	Target string `arg:"positional"`
}

var urls []string

func main() {
	arg.MustParse(&args)
	c := NewCrawler(args.Target).
		AddScraper(SelectiveExtractUrls, HasAttrs("src", "href")).
		AddScraper(Dump, HasAttrs("src", "href").And(IsLeaf))
	c.Crawl()
	// fmt.Println(strings.Join(urls, "\n"))
}

type NodeFilter func(node *html.Node) bool

func (f1 NodeFilter) And(f2 NodeFilter) NodeFilter {
	return func(node *html.Node) bool {
		return f1(node) && f2(node)
	}
}

type NodeScraper func(node *html.Node)

type Scraper struct {
	Filter NodeFilter
	Scrape NodeScraper
}

type Crawler struct {
	Target   string
	Scrapers []Scraper
	Root     *html.Node
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
		c.Scrapers, Scraper{Scrape: s, Filter: f},
	)

	return c
}

func (c *Crawler) Crawl() {
	c.TreeRecurse(c.GetNodeTree())
}

func (c *Crawler) GetNodeTree() *html.Node {
	resp := mustGet(c.Target)
	parentNode, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	c.Root = parentNode

	return parentNode
}

func (c *Crawler) TreeRecurse(doc *html.Node) {
	var f func(*html.Node)
	f = func(n *html.Node) {
		c.Scrape(n)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

func SelectiveExtractUrls(n *html.Node) {
	for _, attr := range n.Attr {
		if attr.Key == "src" || attr.Key == "href" {
			urls = append(urls, attr.Val)
		}
	}
}

func Dump(n *html.Node) {
	err := html.Render(os.Stdout, n)
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stdout.Write([]byte("\n"))
	if err != nil {
		log.Fatal(err)
	}
}

func IgnoreNil(node *html.Node) bool {
	return node != nil
}

func HasAttrs(includedAttrs ...string) NodeFilter {
	return func(node *html.Node) bool {
		var (
			nodeAttrKeys []string
			attrsFound   bool
		)

		for _, attr := range node.Attr {
			nodeAttrKeys = append(nodeAttrKeys, attr.Key)
		}

		for _, inclusion := range includedAttrs {
			attrsFound = slices.Index(nodeAttrKeys, inclusion) != -1
			if attrsFound {
				return true
			}
		}
		return false
	}
}

func IsLeaf(node *html.Node) bool {
	return node.FirstChild == nil
}

func mustGet(url string) *http.Response {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	return resp
}
