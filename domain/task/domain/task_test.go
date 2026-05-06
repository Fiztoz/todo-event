package domain

import "testing"

func TestStatus_IsValid(t *testing.T) {
	cases := []struct {
		s    Status
		want bool
	}{
		{StatusPending, true},
		{StatusInProgress, true},
		{StatusDone, true},
		{Status("anything-else"), false},
	}
	for _, tc := range cases {
		if got := tc.s.IsValid(); got != tc.want {
			t.Errorf("Status(%q).IsValid() = %v, want %v", tc.s, got, tc.want)
		}
	}
}
