package swarm

import (
	"golang.org/x/net/html"
	"log"
	"time"
	"tjweldon/spider/src/messaging"
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
	c.Root = nil
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
	log.Println("Beginning crawl... target: " + target)
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

func (c *Crawler) getWorker(incoming messaging.Backlog[string], id int) *Worker {
	return &Worker{id, c, incoming, make(chan Signal)}
}

func (c *Crawler) Work(incoming messaging.Backlog[string], id int) *Worker {
	worker := c.getWorker(incoming, id)
	go worker.Run()
	log.Printf("Worker %d: Worker Started", worker.id)
	return worker
}

type Worker struct {
	id       int
	crawler  *Crawler
	incoming messaging.Backlog[string]
	done     chan Signal
}

func (w *Worker) Run() {
	var done bool
	for !done {
		select {
		case job, ok := <-w.incoming.Channel():
			if !ok {
				done = true
				break
			}
			w.crawler.CrawlNow(job)
		default:
			if w.hasNoWork() {
				done = true
				break
			}
		}
		// log.Printf("Worker %d: Worker picked up incoming job %s", w.id, job)
	}
	close(w.done)
}

func (w *Worker) hasNoWork() bool {
	noJobs := false
	jobCount := 0

	for range [4]any{} {
		jobCount += w.incoming.Length()
		if jobCount == 0 {
			log.Printf("Worker %d: No jobs, waiting...", w.id)
			time.Sleep(time.Second)
			jobCount += w.incoming.Length()
		}
	}
	if jobCount == 0 {
		log.Printf("Worker %d: Still no jobs, done.", w.id)
		noJobs = true
	}

	return noJobs
}

func (w *Worker) IsDone() bool {
	return util.IsClosed[Signal](w.done)
}

func (w *Worker) AwaitCompletion() {
	util.AwaitClosure[Signal](w.done)
}

func (w *Worker) Die() {
	w.crawler.Die()
}
