package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

var (
	moveMsg = "worker %d x: %d, y: %d\n"
)

type Location = struct {
	X int
	Y int
}

type Worker interface {
	Start()
	Assign(Location)
	CanAssign() bool
	GetPriority() int
	GetLocation() Location
}

type Agent struct {
	id       int
	priority int
	mu       sync.Mutex
	location Location
	dest     chan Location
	wg       *sync.WaitGroup
}

func NewAgent(id int, priority int, wg *sync.WaitGroup) Worker {
	return &Agent{
		id:       id,
		location: Location{0, 0},
		priority: priority,
		dest:     make(chan Location, 1),
		mu:       sync.Mutex{},
		wg:       wg,
	}
}

func (a *Agent) Start() {
	go func() {
		for dest := range a.dest {
			moveCount := math.Min(math.Abs(float64(dest.X-a.location.X)), math.Abs(float64(dest.Y-a.location.Y)))
			fmt.Printf(moveMsg, a.id, a.location.X, a.location.Y)
			for i := 0; i < int(moveCount); i++ {
				a.location.X += 1
				a.location.Y += 1
				fmt.Printf(moveMsg, a.id, a.location.X, a.location.Y)
				time.Sleep(1 * time.Second)
			}
			if m := a.location.X - dest.X; m != 0 {
				for i := 0; i < m; i++ {
					a.location.X += 1
					fmt.Printf(moveMsg, a.id, a.location.X, a.location.Y)
					time.Sleep(1 * time.Second)

				}
			} else if m := a.location.Y - dest.Y; m != 0 {
				for i := 0; i < m; i++ {
					a.location.Y += 1
					fmt.Printf(moveMsg, a.id, a.location.X, a.location.Y)
					time.Sleep(1 * time.Second)

				}
			} else {
				fmt.Printf("worker %d reached the destination x:%d, y:%d\n", a.id, a.location.X, a.location.Y)
				time.Sleep(1 * time.Second)

			}

			a.wg.Done()
			a.mu.Unlock()
		}
	}()
}

func (a *Agent) CanAssign() bool {
	if a.mu.TryLock() {
		defer a.mu.Unlock()
		return true
	}
	return false
}

func (a *Agent) Assign(t Location) {
	a.mu.Lock()
	a.dest <- t
}

func (a *Agent) GetPriority() int {
	return a.priority
}
func (a *Agent) GetLocation() Location {
	return a.location
}

type WorkerPool interface {
	Run()
	AddTask(Location)
	Stop()
}

type AgentPool struct {
	Agents []Worker
	stop   chan struct{}
}

func NewAgentPool(numAgents int, wg *sync.WaitGroup) WorkerPool {
	var agents []Worker

	for i := 0; i < numAgents; i++ {
		agents = append(agents, NewAgent(i+1, i, wg))
	}

	return &AgentPool{
		Agents: agents,
	}
}

func (a AgentPool) Run() {
	for _, agent := range a.Agents {
		agent.Start()
	}
}

func (a AgentPool) AddTask(t Location) {
	for {
		var nominated Worker
		var ready []Worker
		var canAssign []Worker
		for _, agent := range a.Agents {
			if agent.CanAssign() {
				ready = append(ready, agent)
			}
		}
		if len(ready) > 0 {
			loc := ready[0].GetLocation()
			minPath := moves(loc, t)

			for _, r := range ready {
				loc := r.GetLocation()
				mc := moves(loc, t)
				if mc < minPath {
					minPath = mc
				}
			}

			nominated = ready[0]
			for _, ca := range canAssign {
				if moves(ca.GetLocation(), t) == minPath {
					nominated = ca

				}
			}
			for _, ca := range canAssign {
				if ca.GetPriority() < nominated.GetPriority() {
					nominated = ca
				}
			}
			nominated.Assign(t)
			break
		}
	}
}

func (a AgentPool) Stop() {
	a.stop <- struct{}{}
}

func main() {
	numAgents := 2
	dests := []Location{
		{6, 8},
		{2, 2},
		{3, 7},
		{7, 4},
	}
	var wg sync.WaitGroup
	pool := NewAgentPool(numAgents, &wg)
	pool.Run()

	for _, dest := range dests {
		wg.Add(1)
		pool.AddTask(dest)
	}

	wg.Wait()
}

func moves(src Location, dst Location) int {
	moveCount := math.Min(math.Abs(float64(dst.X-src.X)), math.Abs(float64(dst.Y-src.Y)))
	src.X += int(moveCount)
	src.Y += int(moveCount)
	if m := src.X - dst.X; m != 0 {
		moveCount += float64(m)
	} else if m := src.Y - dst.Y; m != 0 {
		moveCount += float64(m)
	}
	return int(moveCount)
}
