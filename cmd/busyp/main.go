// busyp generates random task sets and checks if they are EDF schedulable.
package main

import (
	"fmt"
	"os"
	"sync"
	"flag"

	"github.com/schmr/busyperiod/taskset"
)

// set by linker
var revisiondate string
var revision string

func main() {
	var tries = flag.Int("n", 100000, "number of attempts to find a counterexample")
	var showversion = flag.Bool("v", false, "show version information and exit")
	flag.Parse()

	if *showversion {
		fmt.Println("busyp", revision, revisiondate)
		os.Exit(0)
	}

	type violator struct {
		TaskSet taskset.DualCritMin
		Checkpoints []int
		ViolatedAt int
	}
	violators := make(chan violator)
	var wg sync.WaitGroup
	for i := 0; i < *tries; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Generator
			d := taskset.CreateRandomDualCritMin()
			// EDF-NUVD check and scaling
			ok := d.ScaleTasksetEDFNUVD()
			if !ok {
				return
			}
			// Busy period check
			checkpoints, ok := taskset.BuildCheckpoints(&d)
			if !ok {
				return
			}
			for _, t := range checkpoints {
				w := taskset.WorkBound(&d, float64(t))
				if w > float64(t) { // found counterexample
					var vr violator
					vr.TaskSet = d
					vr.Checkpoints = checkpoints
					vr.ViolatedAt = t
					violators <- vr
				}
			}
		}()
	}

	// check in parallel if we already have a result and print it
	// don't want to wait for all goroutines to finish prior to printing
	// don't track in WaitGroup wg, if no counterexample is found this
	// routine never returns and would stall the WaitGroup
	go func() {
		var first violator
		// unbuffered channel waits until receive completes?
		first = <-violators
		fmt.Printf("\ntaskset not EDF schedulable according to busy period check:\n%v\n", first.TaskSet)
		fmt.Printf("checkpoints: %v\n", first.Checkpoints)
		fmt.Printf("checked t: %v\n", first.ViolatedAt)
	}()

	// wait for all to finish
	go func() {
		wg.Wait()
		close(violators)
	}()

	// drain the channel; might include further counterexamples
	fmt.Print("\n")
	for v := range violators {
		fmt.Printf("\ntaskset not EDF schedulable according to busy period check:\n%v\n", v.TaskSet)
		fmt.Printf("checkpoints: %v\n", v.Checkpoints)
		fmt.Printf("checked t: %v\n", v.ViolatedAt)
	}
}
