package swarm

import (
	"golang.org/x/net/html"
	"log"
	"time"
	"tjweldon/spider/src/messaging"
	"tjweldon/spider/src/util"
)

// FilteredScraper is the object
type FilteredScraper struct {
	Filter NodeFilter
	Scrape NodeScraper
}

// Crawler is the object that encapsulates the recursive walk over
// the html node tree
type Crawler struct {
	// Scrapers are a slice of scraping functions that are applied
	// in sequence to each node. They each have a filter that discards
	// irrelevant nodes in advance, and a Scrape function that takes
	// in a node and produces some side effect, like logging or putting
	// a message on the queue.
	Scrapers []FilteredScraper

	// Root is the Crawler local reference to the parent node of the
	// html tree
	Root *html.Node

	// Done is the channel that a crawler uses to indicate that it has scraped
	// a node
	Done chan Signal

	// Ready is a flag that is set to true if the crawler is ready for more work
	Ready bool
}

// NewCrawler creates a Crawler and hands us a pointer to it
func NewCrawler() *Crawler {
	done := make(chan Signal)
	return &Crawler{
		Done:  done,
		Ready: true,
	}
}

// Scrape iterates over each FilteredScraper in Crawler.Scrapers, applies that
// FilteredScraper's filter to ignore irrelevant nodes and then if not filtered
// out, it scrapes the node
func (c *Crawler) Scrape(node *html.Node) {
	for _, scraper := range c.Scrapers {
		if scraper.Filter(node) {
			scraper.Scrape(node)
		}
	}
}

// AddScraper provides a fluent interface to add new FilteredScrapers for the
// Crawler to apply to each node
func (c *Crawler) AddScraper(s NodeScraper, f NodeFilter) *Crawler {
	c.Scrapers = append(
		c.Scrapers, FilteredScraper{Scrape: s, Filter: f},
	)

	return c
}

// CrawlNow is a blocking recursive walk over the node tree. Each node is passed
// to the configured Scrapers. If there is an error retrieving the response,
// CrawlNow just returns so it can be made ready to pick up another job.
func (c *Crawler) CrawlNow(target string) {
	var f NodeScraper
	c.Root = nil
	f = func(n *html.Node) {
		c.Scrape(n)
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			f(child)
		}
	}
	tree := c.populateNodeTree(target)
	if tree != nil {
		f(tree)
	}
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

// populateNodeTree retrieves the html from the target URL and parses it
// into a node tree. It then stores it in Crawler.Root.
func (c *Crawler) populateNodeTree(target string) *html.Node {
	resp := util.GetOrNil(target)
	if resp == nil {
		return nil
	}
	parentNode, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	c.Root = parentNode

	return parentNode
}

// Die is the crawler teardown function
func (c *Crawler) Die() {
	close(c.Done)
}

// getWorker returns the worker for this crawler
func (c *Crawler) getWorker(incoming messaging.Backlog[string], id int) *Worker {
	return &Worker{id, c, incoming, make(chan Signal)}
}

// Work is a convenience method that encapsulates getting the Worker, setting
// it running and returning a pointer to it back to the calling scope
func (c *Crawler) Work(incoming messaging.Backlog[string], id int) *Worker {
	worker := c.getWorker(incoming, id)
	go worker.Run()
	log.Printf("Worker %d: Worker Started", worker.id)
	return worker
}

// Worker is the wrapper that manages a single crawler,
// allowing it to keep picking up jobs as long as they
// are available.
type Worker struct {
	id       int
	crawler  *Crawler
	incoming messaging.Backlog[string]
	done     chan Signal
}

// Run is the worker function that runs in a goroutine to
// actually do the work.
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
	}
	close(w.done)
}

// hasNoWork returns true if there is no work after 5 seconds
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

// IsDone returns true if the worker backlog channel has been closed
func (w *Worker) IsDone() bool {
	return util.IsClosed[Signal](w.done)
}

// AwaitCompletion is a blocking call that returns when the worker is done
func (w *Worker) AwaitCompletion() {
	util.AwaitClosure[Signal](w.done)
}

// Die is our teardown function
func (w *Worker) Die() {
	w.crawler.Die()
}
