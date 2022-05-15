package swarm

import (
	"log"
	"time"
	"tjweldon/spider/src/messaging"
)

const SwarmSize = 5

type Spawner func() *Crawler

type Swarm struct {
	Spawner    Spawner
	Crawlers   [SwarmSize]*Crawler
	Jobs       []string
	incoming   messaging.Backlog[string]
	dispatcher messaging.Dispatcher[string]
}

// NewSwarm returns a pointer to a new spawn instance
func NewSwarm(spawner Spawner) *Swarm {
	log.Println("Initialising Swarm")

	crawlers := [SwarmSize]*Crawler{}
	for i := range crawlers {
		crawlers[i] = spawner()
	}

	swarm := &Swarm{
		Jobs:     []string{},
		Crawlers: crawlers,
		Spawner:  spawner,
	}
	return swarm
}

func (s *Swarm) Spawn() {
	workers := [SwarmSize]*Worker{}
	for i, crawler := range s.Crawlers {
		workers[i] = crawler.Work(s.incoming, i)
	}

	for _, worker := range workers {
		worker.AwaitCompletion()
		worker.Die()
	}
	s.dispatcher.Close()
}

func (s *Swarm) workersDone(workers [SwarmSize]*Worker) (count int) {
	for _, worker := range workers {
		if worker.IsDone() {
			count++
		}
	}
	return count
}

// countRunning returns the number of currently running crawlers
func (s *Swarm) countRunning() int {
	count := 0
	for _, crawler := range s.Crawlers {
		if !crawler.Ready {
			count++
		}
	}

	log.Printf("Workers Running: %d", count)
	return count
}

// SetIncoming fluently sets the backlog of work for the swarm
func (s *Swarm) SetIncoming(incoming messaging.Backlog[string]) *Swarm {
	s.incoming = incoming
	return s
}

func (s *Swarm) SetDispatcher(dispatcher messaging.Dispatcher[string], seedJobs ...string) *Swarm {
	s.dispatcher = dispatcher
	for _, job := range seedJobs {
		s.dispatcher.Dispatch(job)
	}
	return s
}

type SwarmReport struct {
	Duration time.Duration
}
