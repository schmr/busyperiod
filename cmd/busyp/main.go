// busyp generates random task sets and checks if they are EDF schedulable.
package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/schmr/busyperiod/taskset"
)

// set by linker
var revisiondate string
var revision string

func main() {
	// Maybe there is a package for this I am not aware of?
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-v":
			fmt.Println("busyp", revision, revisiondate)
			os.Exit(0)
		case "-h":
			fmt.Println("busyp [-v] [-h]")
			fmt.Println("\t-v Show version information")
			fmt.Println("\t-h Show this help")
			os.Exit(0)
		default:
			fmt.Println("Ignoring unknown argument.")
		}
	}
	tries := 100000

	type violator struct {
		TaskSet taskset.DualCritMin
		Checkpoints []int
		ViolatedAt int
	}
	violators := make(chan violator)
	var wg sync.WaitGroup
	for i := 0; i < tries; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf(" %d ", id)
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
				if w > float64(t) {
					var vr violator
					vr.TaskSet = d
					vr.Checkpoints = checkpoints
					vr.ViolatedAt = t
					violators <- vr
				}
			}
			fmt.Printf(" %d(E) ", id)
		}(i)
	}

	// wait for all to finish; rewrite to end program on first found counterexample
	go func() {
		wg.Wait()
		close(violators)
	}()

	fmt.Print("\n")
	for v := range violators {
		fmt.Printf("\ntaskset not EDF schedulable according to busy period check:\n%v\n", v.TaskSet)
		fmt.Printf("checkpoints: %v\n", v.Checkpoints)
		fmt.Printf("checked t: %v\n", v.ViolatedAt)
	}
}
