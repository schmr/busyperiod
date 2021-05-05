package taskset

import (
	"fmt"
	"testing"
)

/*
var lowcrit Task = Task{CompLow: 2.0, Period: 12.0, Deadline: 12.0}
var highcrit Task = Task{CompLow: 1.0, CompHigh: 5.0, Period: 10.0,
        Deadline: 10.0, Scale: 0.78}
*/

func TestHighTaskIsHigh(t *testing.T) {
	highcrit := Task{CompLow: 1.0, CompHigh: 5.0, Period: 10.0, Deadline: 10.0, Scale: 0.78}
	got := highcrit.IsHigh()
	want := true

	if got != want {
		t.Errorf("got %t want %t", got, want)
	}
}

func TestLowTaskIsHigh(t *testing.T) {
	lowcrit := Task{CompLow: 2.0, Period: 12.0, Deadline: 12.0}
	got := lowcrit.IsHigh()
	want := false

	if got != want {
		t.Errorf("got %t want %t", got, want)
	}
}

func TestLowTaskString(t *testing.T) {
	lowcrit := Task{CompLow: 2.0, Period: 12.8, Deadline: 12.456}
	got := lowcrit.String()
	want := "2.00\t-\t12.80\t12.46\t-\t-"

	if got != want {
		t.Errorf("got:\n%s\nwant\n%s", got, want)
	}
}

func TestHighTaskString(t *testing.T) {
	highcrit := Task{CompLow: 1.3, CompHigh: 5.786, Period: 10.9, Deadline: 10.0, Scale: 0.78}
	got := highcrit.String()
	want := "1.30\t5.79\t10.90\t10.00\t0.78\t7.80"

	if got != want {
		t.Errorf("got:\n%s\nwant\n%s", got, want)
	}
}

func TestDualCritMinString(t *testing.T) {
	highcrit0 := Task{CompLow: 1.3, CompHigh: 5.786, Period: 10.9, Deadline: 10.0, Scale: 0.78}
	highcrit1 := Task{CompLow: 1.4, CompHigh: 5.786, Period: 10.9, Deadline: 10.0, Scale: 0.78}
	lowcrit := Task{CompLow: 2.0, Period: 12.8, Deadline: 12.456}
	d := DualCritMin{highcrit0, highcrit1, lowcrit}
	got := d.String()
	header := "c_L\tc_H\tp\td\ts\tvd"
	wanthi0 := "1.30\t5.79\t10.90\t10.00\t0.78\t7.80"
	wanthi1 := "1.40\t5.79\t10.90\t10.00\t0.78\t7.80"
	wantlo := "2.00\t-\t12.80\t12.46\t-\t-"
	want := fmt.Sprintf("%s\n%s\n%s\n%s\n", header, wanthi0, wanthi1, wantlo)

	if got != want {
		t.Errorf("got:\n%s\nwant\n%s", got, want)
	}
}
