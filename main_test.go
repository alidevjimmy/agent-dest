package main

import (
	"testing"
	"time"
)

func TestWorkersDestinations(t *testing.T) {
	locs := []Location{
		{-2, -3},
		{-2, 2},
		{3, -7},
		{7, 4},
		{5, -1},
		{-3, -6},
	}
	tests := []struct {
		name       string
		numWorkers int
		dsts       []Location
		delay      time.Duration
	}{
		{
			name:       "3 workers 6 dsts",
			numWorkers: 3,
			delay:      0,
			dsts:       locs,
		},
		{
			name:       "1 workers 6 dsts",
			numWorkers: 1,
			delay:      0,
			dsts:       locs,
		},
		{
			name:       "10 workers 6 dsts",
			numWorkers: 3,
			delay:      0,
			dsts:       locs,
		},
		{
			name:       "3 workers 6 dsts with 1 sec delay",
			numWorkers: 3,
			delay:      0,
			dsts:       locs,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}
