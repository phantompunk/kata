package assert

import "testing"

func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}

func True(t *testing.T, actual bool) {
	t.Helper()
	if !actual {
		t.Errorf("got %v; expected: True", actual)
	}
}

func False(t *testing.T, actual bool) {
	t.Helper()
	if actual {
		t.Errorf("got %v; expected: True", actual)
	}
}

func NilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got %v; expected: nil", actual)
	}
}

func NotNil(t *testing.T, actual any) {
	t.Helper()

	if actual == nil {
		t.Errorf("got nil; expected: not nil")
	}
}
