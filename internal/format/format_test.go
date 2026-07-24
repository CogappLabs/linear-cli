package format

import (
	"testing"
	"time"
)

func TestPriority(t *testing.T) {
	cases := map[int]string{
		0: "None", 1: "Urgent", 2: "High", 3: "Medium", 4: "Low", 9: "None",
	}
	for in, want := range cases {
		if got := Priority(in); got != want {
			t.Errorf("Priority(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestTimeAgo(t *testing.T) {
	if got := TimeAgo(""); got != "never" {
		t.Errorf(`TimeAgo("") = %q, want "never"`, got)
	}
	if got := TimeAgo("not-a-date"); got != "not-a-date" {
		t.Errorf("TimeAgo(bad) should echo the input, got %q", got)
	}
	recent := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	if got := TimeAgo(recent); got != "2h ago" {
		t.Errorf("TimeAgo(2h) = %q, want %q", got, "2h ago")
	}
}

func TestPadRightIgnoresANSI(t *testing.T) {
	// Coloured string wider than its visible text should still pad to width by
	// visible length, not byte length.
	coloured := Cyan("abc") // visible width 3
	got := PadRight(coloured, 5)
	// 2 trailing spaces expected (5 - 3 visible).
	if stripANSI(got) != "abc  " {
		t.Errorf("PadRight visible = %q, want %q", stripANSI(got), "abc  ")
	}
}
