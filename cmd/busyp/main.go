// busyp generates random task sets and checks if they are EDF schedulable.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/schmr/busyperiod/taskset"
)

// set by linker
var revisiondate string = "unknown"
var revision string = "unknown"

type violator struct {
	TaskSet     taskset.DualCritMin
	Checkpoints []int
	ViolatedAt  int
}

func main() {
	var tries = flag.Int("n", 100000, "number of attempts to find a counterexample or zero for endless search")
	var showversion = flag.Bool("v", false, "show version information and exit")
	flag.Parse()

	if *showversion {
		fmt.Println("busyp", revision, revisiondate)
		os.Exit(0)
	}

	rand.Seed(time.Now().UnixNano())
	numWorkers := runtime.NumCPU()
	wp := workerpool.New(numWorkers)
	violators := make(chan violator, numWorkers)

	// Check in parallel if we already have a result and print it.
	// Don't want to wait for all goroutines to finish prior to printing.
	// Don't track in WaitGroup or run in workerpool, if no counterexample
	// is found this routine never returns and would stall the program.
	go func() {
		for {
			v, ok := <-violators
			if !ok {
				break
			}
			fmt.Printf("\ntaskset not EDF schedulable according to busy period check:\n%v\n", v.TaskSet)
			fmt.Printf("checkpoints: %v\n", v.Checkpoints)
			fmt.Printf("checked t: %v\n", v.ViolatedAt)
		}
	}()

	// put all jobs in queue; how many jobs can I put here without problems?
	if *tries > 0 {
		for i := 0; i < *tries; i++ {
			wp.Submit(func() {
				searchCounterExample(violators)
			})
		}
	} else { // endless generation
		for {
			wp.Submit(func() {
				searchCounterExample(violators)
			})
		}
	}

	wp.StopWait()
}

// a complete chain to discover a counterexample
func searchCounterExample(out chan<- violator) {
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
			out <- vr
		}
	}
}
