package main

import (
	"github.com/alexflint/go-arg"
	"log"
	"tjweldon/spider/src/crawler"
)

var args struct {
	Target string `arg:"positional"`
}

var urlChan = make(chan string)

func main() {
	arg.MustParse(&args)
	swarm := crawler.
		NewSwarm().
		SetSpawner(CrawlerSpawner).
		SeedJobs(args.Target)
	defer swarm.Kill()
	swarm.Spawn()
}

func CrawlerSpawner(target string) *crawler.Crawler {
	log.Println("Spawning Crawler")
	HasLinks := crawler.HasAttrs("src", "href")
	return crawler.NewCrawler(target).
		AddScraper(crawler.RecoverUrls(urlChan), HasLinks).
		AddScraper(crawler.Dump, HasLinks.And(crawler.IsLeafNode))
}
