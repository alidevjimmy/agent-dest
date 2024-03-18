package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

var (
	moveMsg = "worker %d x: %d, y: %d; dst x: %d, y: %d\n"
)

type Location = struct {
	X int
	Y int
}

type Worker interface {
	Start(delay time.Duration)
	Assign(dst Location)
	CanAssign() bool
	GetPriority() int
	GetLocation() Location
	Moves(dst Location) int
}

type Agent struct {
	id       int
	priority int
	location Location
	mu       sync.Mutex
	wg       *sync.WaitGroup
	dest     chan Location
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

func (a *Agent) Start(delay time.Duration) {
	go func() {
		for dest := range a.dest {
			moveCount := math.Min(math.Abs(float64(dest.X-a.location.X)), math.Abs(float64(dest.Y-a.location.Y)))
			fmt.Printf(moveMsg, a.id, a.location.X, a.location.Y, dest.X, dest.Y)
			nx := 1
			ny := 1
			switch {
			case (dest.X-a.location.X < 0) && (dest.Y-a.location.Y < 0):
				nx = -1
				ny = -1
			case dest.X-a.location.X < 0:
				nx = -1
			case dest.Y-a.location.Y < 0:
				ny = -1
			}
			for i := 0; i < int(moveCount); i++ {
				a.location.X += nx
				a.location.Y += ny
				fmt.Printf(moveMsg, a.id, a.location.X, a.location.Y, dest.X, dest.Y)
				time.Sleep(delay)
			}
			if m := dest.X - a.location.X; m != 0 {
				xm := 1
				if m < 0 {
					xm = -1
				}
				for i := 0; i < int(math.Abs(float64(m))); i++ {
					a.location.X += xm
					fmt.Printf(moveMsg, a.id, a.location.X, a.location.Y, dest.X, dest.Y)
					time.Sleep(delay)

				}
			} else if m = dest.Y - a.location.Y; m != 0 {
				ym := 1
				if m < 0 {
					ym = -1
				}
				for i := 0; i < int(math.Abs(float64(m))); i++ {
					a.location.Y += ym
					fmt.Printf(moveMsg, a.id, a.location.X, a.location.Y, dest.X, dest.Y)
					time.Sleep(delay)

				}
			}

			fmt.Printf("worker %d reached the destination x:%d, y:%d\n", a.id, a.location.X, a.location.Y)
			time.Sleep(delay)

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

func (a *Agent) Moves(dst Location) int {
	moveCount := math.Min(math.Abs(float64(dst.X-a.location.X)), math.Abs(float64(dst.Y-a.location.Y)))
	tx := moveCount
	ty := moveCount
	if m := math.Abs(tx - float64(dst.X)); m != 0 {
		moveCount += m
	} else if m = math.Abs(ty - float64(dst.Y)); m != 0 {
		moveCount += m
	}
	return int(moveCount)
}

type WorkerPool interface {
	Run(delay time.Duration)
	Wait()
	AddTask(Location)
	Stop()
}

type AgentPool struct {
	Agents []Worker
	wg     *sync.WaitGroup
	stop   chan struct{}
}

func NewAgentPool(numAgents int) WorkerPool {
	var wg sync.WaitGroup
	var agents []Worker

	for i := 0; i < numAgents; i++ {
		agents = append(agents, NewAgent(i+1, i, &wg))
	}

	return &AgentPool{
		Agents: agents,
		wg:     &wg,
		stop:   make(chan struct{}),
	}
}

func (a AgentPool) Run(delay time.Duration) {
	for _, agent := range a.Agents {
		agent.Start(delay)
	}
}

func (a AgentPool) Wait() {
	a.wg.Wait()
}

func (a AgentPool) AddTask(t Location) {
	a.wg.Add(1)
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
			minPath := ready[0].Moves(t)

			for _, r := range ready {
				mc := r.Moves(t)
				if mc < minPath {
					minPath = mc
				}
			}

			nominated = ready[0]
			for _, ca := range canAssign {
				if ca.Moves(t) == minPath {
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
	numAgents := 3
	dsts := []Location{
		{-2, -3},
		{-2, 2},
		{3, -7},
		{7, 4},
		{5, -1},
		{-3, -6},
	}

	pool := NewAgentPool(numAgents)
	pool.Run(1 * time.Second)

	for _, dest := range dsts {
		pool.AddTask(dest)
	}

	pool.Wait()
}
