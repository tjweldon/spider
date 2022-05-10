package crawler

import (
	"fmt"
	"log"
	"time"
	"tjweldon/spider/src/util"
)

const SwarmSize = 2

type Spawner func(job string) *Crawler

type Swarm struct {
	Spawner  Spawner
	Crawlers [SwarmSize]*Crawler
	Signals  [SwarmSize]chan Signal
	Jobs     *util.Deque[string]
}

func NewSwarm() *Swarm {
	log.Println("Initialising Swarm")
	swarm := &Swarm{
		Jobs:     util.NewDeque[string](),
		Crawlers: [SwarmSize]*Crawler{},
		Spawner:  NewCrawler,
		Signals: [2]chan Signal{
			make(chan Signal),
			make(chan Signal),
		},
	}
	return swarm
}

func (s *Swarm) Kill() {
	for _, sigChan := range s.Signals {
		close(sigChan)
	}
}

func (s *Swarm) Spawn() {
	log.Println("Spawning & consuming")
	var (
		crawlerId int
	)
	next := func(cId int) int {
		return (cId + 1) % len(s.Signals)
	}

	go s.startSignals()
	for {
		crawlerId = next(crawlerId)

		fmt.Printf("Crawler Id %d\n", crawlerId)
		time.Sleep(time.Second / 10)
		select {
		case <-s.Signals[crawlerId]:
			if s.Crawlers[crawlerId] != nil {
				log.Printf("Successfully Crawled %s\n", s.Crawlers[crawlerId].Target)
			}
			s.Crawlers[crawlerId] = nil
			if !s.Jobs.IsEmpty() {
				s.refreshCrawler(crawlerId)
			}
		default:
			if s.Jobs.IsEmpty() && s.Population() == 0 {
				return
			}
		}
	}
}

func (s *Swarm) startSignals() {
	for _, sChan := range s.Signals {
		sChan <- Signal{}
	}
}

func (s *Swarm) Population() (count int) {
	for _, crawler := range s.Crawlers {
		if crawler != nil {
			count++
		}
	}
	return count
}

func (s *Swarm) refreshCrawler(crawlerId int) {
	log.Println("Refreshing...")
	s.Crawlers[crawlerId] = s.Spawner(s.Jobs.TakeOne())
	s.Crawlers[crawlerId].Crawl(s.Signals[crawlerId])
}

func (s *Swarm) SetSpawner(c Spawner) *Swarm {
	s.Spawner = c
	return s
}

func (s *Swarm) SeedJobs(jobs ...string) *Swarm {
	for _, j := range jobs {
		s.Jobs.Insert(j)
	}
	return s
}
