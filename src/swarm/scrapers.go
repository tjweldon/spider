package swarm

import (
	"golang.org/x/net/html"
	"log"
	"os"
	"tjweldon/spider/src/messaging"
)

var Urls []string

type NodeScraper func(node *html.Node)

// Then composes scraping operations sequentially:
//
//      var ns NodeScraper = ns1.Then(ns2).Then(ns3)
//
// The scraper assigned to ns executes the scrapers in
// the order ns1, ns2, ns3
func (ns1 NodeScraper) Then(ns2 NodeScraper) NodeScraper {
	return func(node *html.Node) {
		ns1(node)
		ns2(node)
	}
}

func DumpHtml(n *html.Node) {
	err := html.Render(os.Stdout, n)
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stdout.Write([]byte("\n"))
	if err != nil {
		log.Fatal(err)
	}
}

func ScrapeUrls(n *html.Node) {
	for _, attr := range n.Attr {
		if attr.Key == "src" || attr.Key == "href" {
			Urls = append(Urls, attr.Val)
		}
	}
}

func RecoverUrls(dispatcher messaging.Dispatcher[string]) NodeScraper {
	return func(n *html.Node) {
		for _, attr := range n.Attr {
			if attr.Key == "src" || attr.Key == "href" {
				if !dispatcher.Dispatch(attr.Val) {
					return
				}
			}
		}
	}
}
