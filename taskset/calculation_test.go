package taskset

import (
	"testing"
)

func TestHighTaskLowUtilization(t *testing.T) {
	highcrit := Task{CompLow: 1.0, CompHigh: 5.0, Period: 10.0, Deadline: 10.0, Scale: 0.78}
	got := highcrit.LowUtilization()
	want := 0.1

	if !almostEqual(got, want) {
		t.Errorf("got %g want %g", got, want)
	}
}

func TestHighTaskHighUtilization(t *testing.T) {
	highcrit := Task{CompLow: 1.0, CompHigh: 5.0, Period: 10.0, Deadline: 10.0, Scale: 0.78}
	got, err := highcrit.HighUtilization()
	want := 0.5

	if !almostEqual(got, want) && err == nil {
		t.Errorf("got %g want %g", got, want)
	}
}

func TestLowTaskHighUtilization(t *testing.T) {
	lowcrit := Task{CompLow: 4.0, Period: 10.0, Deadline: 10.0}
	got, err := lowcrit.HighUtilization()

	if err == nil {
		t.Errorf("got %g and no error", got)
	}
}

func TestLowTaskLowUtilization(t *testing.T) {
	lowcrit := Task{CompLow: 4.0, Period: 10.0, Deadline: 10.0}
	got := lowcrit.LowUtilization()
	want := 0.4

	if !almostEqual(got, want) {
		t.Errorf("got %g want %g", got, want)
	}
}

func TestValidBusyPeriod(t *testing.T) {
	d := DualCritMin{
		Task{CompLow: 3.0, Period: 10.0, Deadline: 10.0},
		Task{CompLow: 3.0, CompHigh: 4.0, Period: 10.0, Deadline: 10.0, Scale: 1.0},
		Task{CompLow: 2.0, CompHigh: 3.0, Period: 10.0, Deadline: 10.0, Scale: 1.0},
	}
	got, err := BusyPeriod(&d)
	if err != nil {
		t.Errorf(err.Error())
	}
	want := 8.0

	if !almostEqual(got, want) {
		t.Errorf("got %g want %g", got, want)
	}
}

func TestValidWorkBound(t *testing.T) {
	d := DualCritMin{
		Task{CompLow: 3.0, Period: 10.0, Deadline: 10.0},
		Task{CompLow: 3.0, CompHigh: 4.0, Period: 10.0, Deadline: 10.0, Scale: 1.0},
		Task{CompLow: 2.0, CompHigh: 3.0, Period: 10.0, Deadline: 10.0, Scale: 1.0},
	}
	got := WorkBound(&d, 10.0)
	want := 8.0

	if !almostEqual(got, want) {
		t.Errorf("got %g want %g", got, want)
	}
}

// compare if two slices are equal
func equal(x, y []int) bool {
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func TestUniqSlice(t *testing.T) {
	a := []int{0, 1, 2, 3, 3, 3, 4, 6, 2, 10, -1}
	got := uniq(a)
	want := []int{-1, 0, 1, 2, 3, 4, 6, 10}

	if !equal(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestValidBuildCheckpoints(t *testing.T) {
	d := DualCritMin{
		Task{CompLow: 3.0, Period: 10.0, Deadline: 10.0},
		Task{CompLow: 3.0, CompHigh: 4.0, Period: 10.0, Deadline: 10.0, Scale: 1.0},
		Task{CompLow: 2.0, CompHigh: 3.0, Period: 10.0, Deadline: 10.0, Scale: 1.0},
	}
	got, ok := BuildCheckpoints(&d)
	if !ok {
		t.Errorf("not ok")
	}
	/*
	 * With a busy period of 8, and T=D=10 for each task,
	 * the intersection of {10} and {0,1,2,3,4,5,6,7} is the empty set.
	 */
	want := []int{}

	if !equal(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestValidBuildCheckpointsMultiplePeriods(t *testing.T) {
	d := DualCritMin{
		Task{CompLow: 4.0, Period: 7.0, Deadline: 7.0},
		Task{CompLow: 3.0, CompHigh: 4.0, Period: 15.0, Deadline: 15.0, Scale: 1.0},
		Task{CompLow: 2.0, CompHigh: 3.0, Period: 12.0, Deadline: 12.0, Scale: 1.0},
	}
	/*
	 * Given taskset d, we expect a busy period length of 28:
	 *
	 * Initial L0 = 4 + 3 + 2 = 9
	 * W(L0) = ceil(L_0 / 7)*4 + ceil(L_0 / 15)*3 + ceil(L_0 / 12)*2
	 *       = 2*4 + 1*3 + 1*2
	 *       = 13
	 * W(13) = ceil(13/7)*4 + ceil(13/15)*3 + ceil(13/12)*2
	 *       = 8 + 3 + 4 = 15
	 * W(15) = ceil(15/7)*4 + ceil(15/15)*3 + ceil(15/12)*2
	 *       = 3*4 + 3 + 4 = 19
	 * W(19) = ceil(19/7)*4 + ceil(19/15)*3 + ceil(19/12)*2
	 *       = 12 + 6 + 4 = 22
	 * W(22) = ceil(22/7)*4 + ceil(22/15)*3 + ceil(22/12)*2
	 *       = 16 + 6 + 4 = 26
	 * W(26) = ceil(26/7)*4 + ceil(26/15)*3 + ceil(26/12)*2
	 *       = 16 + 6 + 6 = 28
	 * W(28) = ceil(28/7)*4 + ceil(28/15)*3 + ceil(28/12)*2
	 *       = 16 + 6 + 6 = 28
	 * -> Busy period length is L = 28.
	 *
	 * Given d, periods are equal to deadlines.  Moreover, virtual deadlines
	 * are equal to regular deadlines due to a scale of 1.0.
	 * Checkworthy points are therefore kT+T = (k+1)T for all (k+1)T < L:
	 *
	 * Task | Period T |  k | (k+1)T
	 * -----+----------+----+-------
	 *    0 |        7 |  0 |      7
	 *    0 |        7 |  1 |     14
	 *    0 |        7 |  2 |     21
	 *    0 |        7 |  3 |     28
	 * -----+----------+----+-------
	 *    1 |       15 |  0 |     15
	 * -----+----------+----+-------
	 *    2 |       12 |  0 |     12
	 *    2 |       12 |  1 |     24
	 * -----+----------+----+-------
	 *
	 * The set of checkworthy points is {7, 12, 14, 15, 21, 24}.
	 * ( Intersection with {l | forall l in [0,L)} )
	 */
	got, ok := BuildCheckpoints(&d)
	if !ok {
		t.Errorf("not ok")
		return
	}
	want := []int{7, 12, 14, 15, 21, 24} // as explained above

	if !equal(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestEDFNUVDScales(t *testing.T) {
	d := DualCritMin{
		Task{CompLow: 1.0, Period: 5.0, Deadline: 5.0},
		Task{CompLow: 1.0, CompHigh: 2.0, Period: 5.0, Deadline: 5.0, Scale: 1.0},
		Task{CompLow: 1.0, CompHigh: 2.0, Period: 10.0, Deadline: 10.0, Scale: 1.0},
	}
	lowlamb, hilamb, ok := CalculateEDFNUVDLambda(&d)
	if !ok {
		t.Errorf("not ok")
		return
	}
	wantlow := 1.0606601717798212
	wanthi := 1.1785113019775793

	if !almostEqual(lowlamb, wantlow) {
		t.Errorf("got %g want %g", lowlamb, wantlow)
	}
	if !almostEqual(hilamb, wanthi) {
		t.Errorf("got %g want %g", lowlamb, wantlow)
	}
}

func TestEDFNUVDSchedulable(t *testing.T) {
	d := DualCritMin{
		Task{CompLow: 1.0, Period: 5.0, Deadline: 5.0},
		Task{CompLow: 1.0, CompHigh: 2.0, Period: 5.0, Deadline: 5.0, Scale: 1.0},
		Task{CompLow: 1.0, CompHigh: 2.0, Period: 10.0, Deadline: 10.0, Scale: 1.0},
	}
	got := EDFNUVDSchedulable(&d)
	want := true

	if got != want {
		t.Errorf("got %t want %t", got, want)
	}
}

func TestRandomDualCritMin(t *testing.T) {
	for k := 0; k < 100; k++ {
		d := CreateRandomDualCritMin()
		var numhi int
		var numlo int
		var prop [3]bool
		for i, v := range d {
			if v.IsHigh() {
				numhi++
				prop[i] = (v.Period == v.Deadline) && (v.CompLow <= v.CompHigh) && (v.CompHigh <= v.Period)
			} else {
				numlo++
				prop[i] = (v.Period == v.Deadline) && (v.CompLow <= v.Period)
			}
		}

		// we want implicit deadline tasksets with one low and two high criticality tasks
		if (numlo != 1) || (numhi != 2) || !prop[0] || !prop[1] || !prop[2] {
			t.Errorf("random taskset violates desired properties:\n%v", d)
		}
	}
}

func TestScaling(t *testing.T) {
	for k := 0; k < 100; k++ {
		d := CreateRandomDualCritMin()
		ok := d.ScaleTasksetEDFNUVD()
		if ok && ((d[0].Scale > 1) || (d[0].Scale < 0) ||
		          (d[1].Scale > 1) || (d[1].Scale < 0) ||
		          (d[2].Scale > 1) || (d[2].Scale < 0) ){
			t.Errorf("random taskset scaling failed:\n%v", d)
		}
	}
}
