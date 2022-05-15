package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"log"
	"regexp"
	"strings"
	"tjweldon/spider/src/messaging"
	"tjweldon/spider/src/reporting"
	"tjweldon/spider/src/swarm"
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

func DoCrawl() {
	dispatcher, backlog := messaging.NewQ[string](swarm.SwarmSize * 1024)
	withPreProcessors := ProvisionDispatcher(dispatcher)

	var (
		fork   messaging.Backlog[string]
		result <-chan string
	)

	backlog, fork = messaging.Fork(backlog)
	result = reporting.DomainsReport(fork)

	s := swarm.
		NewSwarm(NewSpawner(withPreProcessors).Create).
		SetIncoming(backlog).
		SetDispatcher(withPreProcessors, args.Target)
	defer CleanUp(result, withPreProcessors)

	s.Spawn()
}

func ProvisionDispatcher(dispatcher messaging.Dispatcher[string]) messaging.Dispatcher[string] {
	withDeDuplication := messaging.WithDeDuplication[string](dispatcher).
		SetMaxJobs(256)

	withValidation := AddValidation(withDeDuplication)
	withPreProcessors := AddPreProcessors(withValidation)
	return withPreProcessors
}

func CleanUp(result <-chan string, d messaging.Dispatcher[string]) {
	d.Close()
	fmt.Println(<-result)
}

func AddPreProcessors(dispatcher messaging.Dispatcher[string]) messaging.Dispatcher[string] {
	dispatcher = messaging.WithPreProcessing[string](
		dispatcher,
		func(item string) string {
			return strings.TrimLeft(item, "./")
		},
		func(item string) string {
			if strings.HasPrefix(item, "http") {
				return item
			}
			target := strings.TrimRight(args.Target, "/")
			altTarget := strings.Join(strings.Split(target, "www."), "")
			if !(strings.HasPrefix(item, target) || strings.HasPrefix(item, altTarget)) {
				return strings.Join([]string{target, item}, "/")
			}
			return item
		},
	)
	return dispatcher
}

func AddValidation(dispatcher messaging.Dispatcher[string]) messaging.Dispatcher[string] {
	dispatcher = messaging.WithValidation[string](
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
	dispatcher messaging.Dispatcher[string]
}

func NewSpawner(dispatcher messaging.Dispatcher[string]) *Spawner {
	return &Spawner{dispatcher: dispatcher}
}

func (s *Spawner) Create() *swarm.Crawler {
	log.Println("Spawning Crawler")
	HasLinks := swarm.HasAttrs("src", "href")
	return swarm.NewCrawler().
		AddScraper(swarm.RecoverUrls(s.dispatcher), HasLinks)
	// .AddScraper(swarm.DumpHtml, HasLinks.And(swarm.IsLeafNode))
}
