package crawler

import (
	"golang.org/x/net/html"
	"log"
	"os"
)

var Urls []string

type NodeScraper func(node *html.Node)

func (ns1 NodeScraper) Then(ns2 NodeScraper) NodeScraper {
	return func(node *html.Node) {
		ns1(node)
		ns2(node)
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

func ScrapeUrls(n *html.Node) {
	for _, attr := range n.Attr {
		if attr.Key == "src" || attr.Key == "href" {
			Urls = append(Urls, attr.Val)
		}
	}
}

func RecoverUrls(generator chan<- string) NodeScraper {
	return func(n *html.Node) {
		for _, attr := range n.Attr {
			if attr.Key == "src" || attr.Key == "href" {
				generator <- attr.Val
			}
		}
	}
}
