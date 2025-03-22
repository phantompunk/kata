package models

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestDB(t *testing.T) *sql.DB {
	dbPath := filepath.Join("./testdata", "test.db")
	setupPath := filepath.Join("./testdata", "setup.sql")
	teardownPath := filepath.Join("./testdata", "teardown.sql")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Error("could not open file")
		t.Fatal(err)
	}

	script, err := os.ReadFile(setupPath)
	if err != nil {
		t.Error("could not read file")
		db.Close()
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		t.Error("could not exec file", err.Error())
		db.Close()
	}

	t.Cleanup(func() {
		script, err := os.ReadFile(teardownPath)
		if err != nil {
			t.Logf("Teardown script read error: %v", err) // Log the error
			t.Fail()
		} else {
			_, err = db.Exec(string(script))
			if err != nil {
				t.Logf("Teardown script exec error: %v", err) // Log the error
				t.Fail()
			}
		}

		db.Close()
		os.Remove(dbPath)
	})

	return db
}
