// Package taskset models computations with real-time constraints for schedulability checks.
package taskset

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// Task is a model of a recurring computation.
// It reduces the execution time distribution to two estimates:
// CompLow in the regular case and CompHigh in the worst case.
// Deadline and Period are relative values; the absolute deadline of a job
// would be Deadline time units after its arrival.
type Task struct {
	CompLow  float64
	CompHigh float64
	Period   float64
	Deadline float64
	Scale    float64
}

// DualCritMin is a minimal taskset of two high criticality tasks and one
// low criticality task.
type DualCritMin [3]Task

const float64EqualityThreshold = 1e-9

// Is this best practice?
// https://stackoverflow.com/a/47969546
func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

// IsHigh returns true if the task is of high criticality.
func (t Task) IsHigh() bool {
	return (t.Scale != 0) && (t.CompHigh > 0)
}

// VirtualDeadline calculates the earlier, virtual deadline.
// If the task is of low criticality it returns an error.
func (t Task) VirtualDeadline() (float64, error) {
	if t.IsHigh() {
		return t.Deadline * t.Scale, nil
	}
	return math.NaN(), fmt.Errorf("low criticality task has no virtual deadline")
}

// LowUtilization calculates the ratio of low computation bound to deadline.
func (t Task) LowUtilization() float64 {
	return t.CompLow / t.Deadline
}

// HighUtilization calculates the ratio of high computation bound to deadline.
// If the task is of low criticality it returns an error.
func (t Task) HighUtilization() (float64, error) {
	if t.IsHigh() {
		return t.CompHigh / t.Deadline, nil
	}
	return math.NaN(), fmt.Errorf("low criticality task has no high utilization")
}

// String presents a task depending on its criticality.
func (t Task) String() string {
	v, err := t.VirtualDeadline()
	if err != nil {
		return fmt.Sprintf("%3.2f\t-\t%3.2f\t%3.2f\t-\t-",
			t.CompLow, t.Period, t.Deadline)
	}
	return fmt.Sprintf("%3.2f\t%3.2f\t%3.2f\t%3.2f\t%3.2f\t%3.2f",
		t.CompLow, t.CompHigh, t.Period, t.Deadline, t.Scale, v)
}

// String presents the task set as a table.
func (d DualCritMin) String() string {
	header := "c_L\tc_H\tp\td\ts\tvd"
	return fmt.Sprintf("%s\n%s\n%s\n%s\n", header, d[0], d[1], d[2])
}

// BusyPeriod recursively computes the length of the synchronous busy period
// for a synchronous job release pattern in low criticality mode.
// It is the ASAP job release pattern from t=0 up to the first idle time.
func BusyPeriod(d *DualCritMin) (float64, error) {
	// Calculate initial busy period length
	var bL float64
	for _, v := range *d {
		bL += v.CompLow
	}
	return busyPeriodTail(d, bL)
}

func busyPeriodTail(d *DualCritMin, bL float64) (float64, error) {
	if bL > 1e6 {
		return math.NaN(), fmt.Errorf("can't handle possible endless busy period")
	}
	var bLNext float64
	for _, v := range *d {
		bLNext += math.Ceil(bL/v.Period) * v.CompLow
	}
	//fmt.Println(bL,bLNext)
	if int(bL) == int(bLNext) {
		//fmt.Println("Returning", bL)
		return bL, nil
	}
	return busyPeriodTail(d, bLNext)
}

// WorkBound calculates the maximum possible computation request up to t.
// It considers the criticality, and the resulting bound is for the low
// criticality mode.
func WorkBound(d *DualCritMin, t float64) float64 {
	var accum float64
	var deadl float64
	for _, v := range *d {
		vdeadl, err := v.VirtualDeadline()
		if err != nil {
			deadl = v.Deadline
		} else {
			deadl = vdeadl
		}
		accum += math.Max(0.0, 1.0+
			math.Floor((t-deadl)/v.Period)) * v.CompLow
	}
	return accum
}

// uniq removes all duplicates from an integer slice.
// Argument slice a is not modified, instead a new slice is returned.
func uniq(a []int) []int {
	b := a
	if len(b) == 0 {
		return b
	}
	sort.Ints(b)
	var s []int
	var last int = b[0]
	s = append(s, last)
	for _, v := range b {
		if v != last {
			s = append(s, v)
		}
		last = v
	}
	return s
}

// BuildCheckpoints creates a set of checkworthy time points t where the
// WorkBound function h(t) should be below t for EDF schedulability.
func BuildCheckpoints(d *DualCritMin) ([]int, bool) {
	var s []int
	busyperiod, err := BusyPeriod(d)
	if err != nil {
		return s, false
	}
	bL := int(busyperiod)
	for _, v := range *d {
		vdeadl, err := v.VirtualDeadline()
		var deadl float64
		if err != nil {
			deadl = v.Deadline
		} else {
			deadl = vdeadl
		}
		var tp int = 0
		for k := 0; tp < bL; k++ {
			tp = k*int(v.Period) + int(deadl)
			if tp < bL {
				s = append(s, tp)
			}
		}
	}
	return uniq(s), true
}

// EDFNUVDSchedulable checks if a task set is schedulable under EDF-NUVD.
func EDFNUVDSchedulable(d *DualCritMin) bool {
	_, _, ok := CalculateEDFNUVDLambda(d)
	return ok
}

// CalculateEDFNUVDLambda tries to calculate an intermediate result
// required to calculate virtual deadline scales under EDF-NUVD.
// Success implies that the task set is schedulable under EDF-NUVD.
func CalculateEDFNUVDLambda(d *DualCritMin) (float64, float64, bool) {
	var s12 float64
	var uhh float64 // utilization of all high tasks in high mode
	var uhl float64 // utilization of all high tasks in low mode
	var ull float64 // utilization of all low tasks in low mode
	for _, v := range *d {
		if v.IsHigh() {
			ul := v.LowUtilization()
			uh, _ := v.HighUtilization()
			s12 += math.Sqrt(ul * uh)
			uhh += uh
			uhl += ul
		}
	} // Theorem 6.2
	if (uhh >= 1.0) || (s12 == 0.0) {
		// for supply of 1, uhh=1 results in divide by zero
		return math.NaN(), math.NaN(), false
	}
	for _, v := range *d {
		if !v.IsHigh() {
			ull += v.LowUtilization()
		}
	}
	lowlamb := s12 / (1.0 - uhh)
	hilamb := (1.0 - ull - uhl) / s12
	if lowlamb > hilamb {
		return math.NaN(), math.NaN(), false
	}
	return lowlamb, hilamb, true
}

// ScaleTasksetEDFNUVD calculates and updates the taskset's scales according to EDFNUVD.
func (d DualCritMin) ScaleTasksetEDFNUVD() bool {
	ll, _, ok := CalculateEDFNUVDLambda(&d)
	if !ok {
		return false
	}
	scaleformula := func(l, uh, ul float64) float64 {
		return 1.0 / (1 + l*math.Sqrt(uh/ul))
	}
	for i, v := range d {
		if v.IsHigh() {
			ul := v.LowUtilization()
			uh, _ := v.HighUtilization()
			d[i].Scale = scaleformula(ll, uh, ul)
		}
	}
	return true
}

// CreateRandomDualCritMin creates a random implicit deadline dual criticality task set
// of one low criticality task and two high criticality tasks:
// forall task CompLow <= CompHigh <= Deadline = Period
func CreateRandomDualCritMin() DualCritMin {
	var d DualCritMin
	var tl, th1, th2 Task
	var cl, ch, period float64

	gen := func() (period, ch, cl float64) {
		for (ch == 0) || (cl == 0) { // avoid zero cl or ch
			period = float64(1 + rand.Intn(20))
			ch = float64(int(rand.Float64() * period))
			cl = float64(int(rand.Float64() * ch))
		}
		return period, ch, cl
	}

	period, _, cl = gen()
	tl.CompLow = cl
	tl.Period = period
	tl.Deadline = period
	d[0] = tl

	period, ch, cl = gen()
	th1.CompLow = cl
	th1.CompHigh = ch
	th1.Period = period
	th1.Deadline = period
	th1.Scale = 1.0 // get recognized as high crit task
	d[1] = th1

	period, ch, cl = gen()
	th2.CompLow = cl
	th2.CompHigh = ch
	th2.Period = period
	th2.Deadline = period
	th2.Scale = 1.0 // get recognized as high crit task
	d[2] = th2

	return d
}
