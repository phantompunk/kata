package kata

import (
	"fmt"
	"testing"
)

func TestSol(t *testing.T) {
	testCases := []struct {
		strA string
		strB string
		want bool
	}{
		{"a", "b", false},
		{"a", "a", true},
		{"a", "aa", false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s -> %s", tc.strA, tc.strB), func(t *testing.T) {
			got := Sol(tc.strA, tc.strB)
			if got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}
