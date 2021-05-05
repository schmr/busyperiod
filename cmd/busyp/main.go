package main

import (
	"fmt"
	"os"

	"github.com/schmr/busyperiod/taskset"
)

func main() {
	fmt.Println(os.Args)
	for k := 0; k < 100; k++ {
		// Generator
		d := taskset.CreateRandomDualCritMin()
		// EDF-NUVD check and scaling
		ok := d.ScaleTasksetEDFNUVD()
		if !ok {
			continue
		}
		// Busy period check
		checkpoints, ok := taskset.BuildCheckpoints(&d)
		if !ok {
			continue
		}
		for _, t := range checkpoints {
			w := taskset.WorkBound(&d, float64(t))
			if w > float64(t) {
				fmt.Printf("\ntaskset not EDF schedulable according to busy period check:\n%v\n", d)
				fmt.Printf("checkpoints: %v\n", checkpoints)
				fmt.Printf("checked t: %v\n", t)
				break
			}
		}
		fmt.Print(".")
	}
}
