package main

import (
	"fmt"
	"math"
	"sync"
)

var (
	moveMsg = "worker %d x: %d, y: %d"
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
}

type Agent struct {
	id       int
	priority int
	mu       sync.Mutex
	location Location
	dest     chan Location
}

func NewAgent(id int, priority int) Worker {
	return &Agent{
		id:       id,
		location: Location{0, 0},
		priority: priority,
		dest:     make(chan Location, 1),
		mu:       sync.Mutex{},
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
			}
			if m := a.location.X - dest.X; m != 0 {
				for i := 0; i < m; i++ {
					a.location.X += 1
					fmt.Printf(moveMsg, a.id, a.location.X, a.location.Y)
				}
			} else if m := a.location.Y - dest.Y; m != 0 {
				for i := 0; i < m; i++ {
					a.location.Y += 1
					fmt.Printf(moveMsg, a.id, a.location.X, a.location.Y)
				}
			} else {
				fmt.Printf("worker %d reached the destination x:%d, y:%d", a.id, a.location.X, a.location.Y)
			}
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

type WorkerPool interface {
	Run()
	AddTask(Location)
	Stop()
}

type AgentPool struct {
	Agents []Worker
	stop   chan struct{}
}

func NewAgentPool(numAgents int) WorkerPool {
	var agents []Worker

	for i := 0; i < numAgents; i++ {
		agents = append(agents, NewAgent(i, i))
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
			for _, r := range ready {
				canAssign = append(canAssign, r)
			}
			nominated = canAssign[0]
			for _, ca := range canAssign {
				if ca.GetPriority() < nominated.GetPriority() {
					nominated = ca
				}
			}
		}
		nominated.Assign(t)
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
		{3, 1},
		{7, 4},
	}
	done := make(chan struct{}, len(dests))
	pool := NewAgentPool(numAgents)
	pool.Run()

	for _, dest := range dests {
		pool.AddTask(dest)
	}
	//wait group for done
	for i := 0; i < len(dests); i++ {
		<-done
	}
}
