package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	_ "github.com/mattn/go-sqlite3"
)

func GetDbPath() string {
	return filepath.Join(xdg.DataHome, "kata", "kata.db")
}

func EnsureDB(dbPath string) (*sql.DB, error) {
	if dbErr := os.MkdirAll(filepath.Dir(dbPath), os.ModePerm); dbErr != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", dbErr)
	}

	if _, err := os.Stat(dbPath); errors.Is(err, os.ErrNotExist) {
		db, err := createDatabaseFile(dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file: %w", err)
		}
		return db, nil
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
			questionId INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			titleSlug TEXT UNIQUE NOT NULL,
			difficulty TEXT CHECK (difficulty IN ('Easy', 'Medium', 'Hard')) NOT NULL,
			functionName TEXT NOT NULL,
			content TEXT NOT NULL,
			codeSnippets TEXT NOT NULL
			);

			CREATE TABLE status (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			questionId INTEGER NOT NULL,
			langSlug TEXT NOT NULL,
			solved INTEGER CHECK (solved IN (0, 1)) NOT NULL,
			FOREIGN KEY (questionId) REFERENCES questions(questionId) ON DELETE CASCADE
			);
			`)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}

	return db, nil
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
	questionId INTEGER PRIMARY KEY,
	title TEXT NOT NULL,
	titleSlug TEXT UNIQUE NOT NULL,
	difficulty TEXT CHECK (difficulty IN ('Easy', 'Medium', 'Hard')) NOT NULL,
	functionName TEXT NOT NULL,
	content TEXT NOT NULL,
	codeSnippets TEXT NOT NULL
	);

	CREATE TABLE status (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	questionId INTEGER NOT NULL,
	langSlug TEXT NOT NULL,
	solved INTEGER CHECK (solved IN (0, 1)) NOT NULL,
	FOREIGN KEY (questionId) REFERENCES questions(questionId) ON DELETE CASCADE
	);`
