package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type SQLModel struct {
	db  *sql.DB
	rnd *rand.Rand
}

type Drop struct {
	ID          int
	Content     string
	TimeCreated time.Time
}

func NewSQLModel(db *sql.DB) (*SQLModel, error) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	model := &SQLModel{db, rnd}
	_, err := model.db.Exec(`
		CREATE TABLE IF NOT EXISTS drops (
			id INTEGER NOT NULL PRIMARY KEY,
			content VARCHAR(1000) NOT NULL,
			uuid VARCHAR(500) NOT NULL,
			time_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  time_deleted TIMESTAMP
		);
    `)
	return model, err
}

func (m *SQLModel) Add(content string) (string, error) {
	// Generate time here because SQLite's CURRENT_TIMESTAMP only returns seconds.
	timeCreated := time.Now().In(time.UTC).Format(time.RFC3339Nano)
	uuid := uuid.NewString()
	_, err := m.db.Exec(`INSERT INTO drops (content, uuid, time_created) VALUES (?, ?, ?)`, content, uuid, timeCreated)
	if err != nil {
		return "", err
	}
	return uuid, nil
}

func (m *SQLModel) Get(uuid string) (string, error) {
	rows, err := m.db.Query("select content from drops WHERE uuid = ?", uuid)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var content string
	c := 0
	for rows.Next() {
		if c > 0 {
			return "", fmt.Errorf("getting more than one row for uuid: %s", uuid)
		}
		err = rows.Scan(&content)
		if err != nil {
			return "", err
		}
		c += 1
	}

	return content, nil
}

func (m *SQLModel) Delete(uuid string) error {
	_, err := m.db.Exec("DELETE FROM drops WHERE uuid = ?", uuid)
	if err != nil {
		return err
	}

	return nil
}
