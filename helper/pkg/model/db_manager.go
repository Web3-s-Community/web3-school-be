package model

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
)

type DbManager struct {
	ConnectionString string `json:"connection_string"`
	db               *sqlx.DB
}

func NewDbManagerWDB(db *sqlx.DB) *DbManager {
	return &DbManager{
		db: db,
	}
}

func NewDbManager(connStr string) *DbManager {
	// db, err := sqlx.Connect("mysql", "test:test@(localhost:3306)/test")
	db, err := sqlx.Connect("mysql", connStr)
	if err != nil {
		log.Fatalln(err)
	}
	return &DbManager{
		db: db,
	}
}

func (m *DbManager) GetAllChallenges() ([]*Challenge, error) {
	var challenges []*Challenge
	query := "SELECT id, slug, language, title, difficulty, points, free, tags, prompt, videos, starter, tasks, hints, code, solution, test FROM challenge"
	rows, err := m.db.Query(query)
	if err != nil {
		fmt.Println("Cannot get challenges: ", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Challenge
		err := rows.Scan(
			&c.ID,
			&c.Slug,
			&c.Language,
			&c.Title,
			&c.Difficulty,
			&c.Points,
			&c.Free,
			&c.Tags,
			&c.Prompt,
			&c.Videos,
			&c.Starter,
			&c.Tasks,
			&c.Hints,
			&c.Code,
			&c.Solution,
			&c.Test,
		)
		if err != nil {
			fmt.Println("Error scanning challenge row: ", err)
			return nil, err
		}
		challenges = append(challenges, &c)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Error iterating through challenge rows: ", err)
		return nil, err
	}

	return challenges, nil
}

func (m *DbManager) GetChallengeByID(challengeID string) (*Challenge, error) {
	var c Challenge
	query := "SELECT id, slug, language, title, difficulty, points, free, tags, prompt, videos, starter, tasks, hints, code, solution, test FROM challenge WHERE id = ?"
	err := m.db.QueryRow(query, challengeID).Scan(
		&c.ID,
		&c.Slug,
		&c.Language,
		&c.Title,
		&c.Difficulty,
		&c.Points,
		&c.Free,
		&c.Tags,
		&c.Prompt,
		&c.Videos,
		&c.Starter,
		&c.Tasks,
		&c.Hints,
		&c.Code,
		&c.Solution,
		&c.Test,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Challenge not found.")
			return nil, nil // Return nil and nil error for indicating that the challenge is not found.
		}
		fmt.Println("Error retrieving challenge by ID: ", err)
		return nil, err
	}
	return &c, nil
}
