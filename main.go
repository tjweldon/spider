package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"log"
	"regexp"
	"strings"
	"tjweldon/spider/src/crawler"
	"tjweldon/spider/src/util"
)

var args struct {
	Target string `arg:"positional"`
}

var crawlUrlPattern = regexp.MustCompile(
	`(?m)https?://[a-zA-Z\d./:]+/[a-zA-Z\d]*(\.html)?$`,
)

func main() {
	arg.MustParse(&args)
	DoCrawl()
}

func RegexpTesting() {
	f := func(item string) string {
		return crawlUrlPattern.
			ReplaceAllStringFunc(
				item,
				func(s string) string {
					return strings.ToUpper(s)
				},
			)
	}
	tests := []string{
		"http://127.0.0.1:8000/url-encode.c",
		"http://127.0.0.1:8000/urlenc",
		"http://127.0.0.1:8000/a.out",
		"http://127.0.0.1:8000/",
	}

	for _, t := range tests {
		fmt.Println("Matches " + f(t))
	}
}

func DoCrawl() {
	dispatcher, backlog := util.NewQ[string]()
	withDeDuplication := util.
		WithDeDuplication[string](dispatcher).
		SetMaxJobs(50)
	withValidation := AddValidation(withDeDuplication)
	withPreProcessors := AddPreProcessors(withValidation)

	log.Printf("%+v\n", withPreProcessors)
	swarm := crawler.
		NewSwarm(NewSpawner(withPreProcessors).Spawn).
		SetIncoming(backlog).
		SeedJobs(args.Target)
	defer CleanUp(withDeDuplication)

	swarm.Spawn()
}

func CleanUp(dispatcher *util.DeDuplicatingDispatcher[string]) {
	for _, job := range dispatcher.ReportDispatched() {
		fmt.Println(job)
	}
}

func AddPreProcessors(dispatcher util.Dispatcher[string]) util.Dispatcher[string] {
	dispatcher = util.WithPreProcessing[string](
		dispatcher,
		func(item string) string {
			return strings.TrimLeft(item, "./")
		},
		func(item string) string {
			if strings.HasPrefix(item, "http") {
				return item
			}
			target := strings.TrimRight(args.Target, "/")
			alt_target := strings.Join(strings.Split(target, "www."), "")
			if !(strings.HasPrefix(item, target) || strings.HasPrefix(item, alt_target)) {
				return strings.Join([]string{target, item}, "/")
			}
			return item
		},
	)
	return dispatcher
}

func AddValidation(dispatcher util.Dispatcher[string]) util.Dispatcher[string] {
	dispatcher = util.WithValidation[string](
		dispatcher,
		func(item string) bool {
			return !strings.Contains(item, "#")
		},
		func(item string) bool {
			return crawlUrlPattern.MatchString(item)
		},
	)
	return dispatcher
}

type Spawner struct {
	dispatcher util.Dispatcher[string]
}

func NewSpawner(dispatcher util.Dispatcher[string]) *Spawner {
	return &Spawner{dispatcher: dispatcher}
}

func (s *Spawner) Spawn() *crawler.Crawler {
	log.Println("Spawning Crawler")
	HasLinks := crawler.HasAttrs("src", "href")
	return crawler.NewCrawler().
		AddScraper(crawler.RecoverUrls(s.dispatcher), HasLinks).
		AddScraper(crawler.Dump, HasLinks.And(crawler.IsLeafNode))
}
