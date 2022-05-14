package crawler

import (
	"fmt"
	"log"
	"time"
	"tjweldon/spider/src/util"
)

const SwarmSize = 20

type Spawner func() *Crawler

type Swarm struct {
	Spawner  Spawner
	Crawlers [SwarmSize]*Crawler
	Jobs     *util.Deque[string]
	incoming util.Backlog[string]
}

// NewSwarm returns a pointer to a new spawn instance
func NewSwarm(spawner Spawner) *Swarm {
	log.Println("Initialising Swarm")

	crawlers := [SwarmSize]*Crawler{}
	for i := range crawlers {
		crawlers[i] = spawner()
	}

	swarm := &Swarm{
		Jobs:     util.NewDeque[string](),
		Crawlers: crawlers,
		Spawner:  spawner,
	}
	return swarm
}

// Spawn starts the swarm at the
func (s *Swarm) Spawn() {
	var restartCount int
	log.Println("Spawning & Consuming")
	for {
		time.Sleep(time.Second / 10)

		s.pollIncomingJobs()

		// if there's no jobs and no running crawlers
		// after polling, we're done
		if s.Jobs.IsEmpty() && s.countRunning() == 0 {
			fmt.Println("returning")
			return
		}

		// Check all the crawlers and refresh
		// any that are reporting completion.
		restartCount = 0
		for crawlerId := range s.Crawlers {
			select {
			// Mark done crawlers as ready
			case <-s.Crawlers[crawlerId].Done:
				s.Crawlers[crawlerId].Ready = true
			default:
			}

			// Wrangle any Crawlers that are ready and give them
			// a job to do
			if !s.Jobs.IsEmpty() && s.Crawlers[crawlerId].Ready {
				restartCount++
				s.refreshCrawler(crawlerId)
			}
		}
		log.Printf("%d crawlers restarted\n", restartCount)
	}
}

// pollIncomingJobs for jobs, break loop when
// the next job isn't immediately available
func (s *Swarm) pollIncomingJobs() {
inner:
	for {
		time.Sleep(time.Millisecond * 10)
		select {
		case job := <-s.incoming.Channel():
			s.Jobs.Insert(job)
			break
		default:
			break inner
		}
	}
}

// refreshCrawler gives the identified crawler a new job to do
func (s *Swarm) refreshCrawler(crawlerId int) {
	job := s.Jobs.TakeOne()
	s.Crawlers[crawlerId].Crawl(job)
	log.Printf("Refreshing... Job: %s\n", job)
}

// countRunning returns the number of currently running crawlers
func (s *Swarm) countRunning() int {
	count := 0
	for _, crawler := range s.Crawlers {
		if !crawler.Ready {
			count++
		}
	}

	return count
}

// SetIncoming fluently sets the backlog of work for the swarm
func (s *Swarm) SetIncoming(incoming util.Backlog[string]) *Swarm {
	s.incoming = incoming
	return s
}

// SeedJobs initialises the Swarm. Make sure to adequately feed the
// Swarm, it dies without food!
func (s *Swarm) SeedJobs(jobs ...string) *Swarm {
	for _, j := range jobs {
		s.Jobs.Insert(j)
	}
	return s
}

type SwarmReport struct {
	Duration time.Duration
}
