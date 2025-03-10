package assert

import "testing"

func Equal(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
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
