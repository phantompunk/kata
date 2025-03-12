package datastore

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	_ "github.com/mattn/go-sqlite3"
)

type Datastore struct {
	DB *sql.DB
}

func GetDbPath() string {
	return filepath.Join(xdg.DataHome, "kata", "kata.db")
}

func EnsureDB(dbPath string) (*Datastore, error) {
	if dbErr := os.MkdirAll(filepath.Dir(dbPath), os.ModePerm); dbErr != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", dbErr)
	}

	if _, err := os.Stat(dbPath); errors.Is(err, os.ErrNotExist) {
		db, err := createDatabaseFile(dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file: %w", err)
		}
		return &Datastore{DB: db}, nil
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='questions';")
	if err != nil {
		db.Close()
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		_, err = db.Exec(`
			CREATE TABLE questions (
			questionId TEXT PRIMARY KEY,
			title TEXT,
			titleSlug TEXT,
			difficulty TEXT,
			content TEXT,
			codeSnippets TEXT
			);
			`)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}

	return &Datastore{DB: db}, nil
}

func createDatabaseFile(dbPath string) (*sql.DB, error) {
	file, err := os.Create(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database file: %w", err)
	}
	defer file.Close()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec(stmt)
	if err != nil {
		return nil, err
	}
	return db, nil
}

const stmt string = ` 
	CREATE TABLE questions (
	questionId TEXT PRIMARY KEY,
	title TEXT,
	titleSlug TEXT,
	difficulty TEXT,
	content TEXT,
	codeSnippets TEXT
	);`
