package bonk

import (
	"math"
	"testing"
)

// TestLookupKey checks that characters and named keys get the code,
// virtual key code and Shift state a real US keyboard reports.
func TestLookupKey(t *testing.T) {
	cases := []struct {
		in    string
		code  string
		vk    int64
		shift bool
	}{
		{"a", "KeyA", 65, false},
		{"Q", "KeyQ", 81, true},
		{"7", "Digit7", 55, false},
		{"!", "Digit1", 49, true},
		{"@", "Digit2", 50, true},
		{".", "Period", 190, false},
		{" ", "Space", 32, false},
		{"Enter", "Enter", 13, false},
		{"ArrowDown", "ArrowDown", 40, false},
	}
	for _, c := range cases {
		d := lookupKey(c.in)
		if d.code != c.code || d.vk != c.vk || d.shift != c.shift {
			t.Errorf("lookupKey(%q) = %+v, want code %s vk %d shift %v",
				c.in, d, c.code, c.vk, c.shift)
		}
	}
}

// TestHumanPath checks that a pointer path has several steps, ends exactly
// on the target and never jumps further than a fraction of the distance.
func TestHumanPath(t *testing.T) {
	pts := humanPath(10, 10, 810, 410)
	if len(pts) < 8 {
		t.Fatalf("got %d points, want at least 8", len(pts))
	}
	last := pts[len(pts)-1]
	if last[0] != 810 || last[1] != 410 {
		t.Errorf("path ends at %v, want [810 410]", last)
	}
	px, py := 10.0, 10.0
	for _, p := range pts {
		if step := math.Hypot(p[0]-px, p[1]-py); step > 300 {
			t.Errorf("step of %.0fpx is a jump", step)
		}
		px, py = p[0], p[1]
	}
	if got := humanPath(5, 5, 5.5, 5); len(got) != 1 {
		t.Errorf("tiny move got %d points, want 1", len(got))
	}
}
