package assert

import "testing"

func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}

func NilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got %v; expected: nil", actual)
	}
}

// func Exists(t *testing.T, stubFilePath string) {
// 	t.Helper()
//
// 	_, err := testFS.Stat(stubFilePath)
// 	if err != nil {
// 		t.Error("sample template not found")
// 	}
// }
