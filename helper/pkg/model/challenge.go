package model

type Challenge struct {
	ID         string  `db:"id" json:"id"`
	Slug       string  `db:"slug" json:"slug"`
	Language   string  `db:"language" json:"language"`
	Title      string  `db:"title" json:"title"`
	Difficulty string  `db:"difficulty" json:"difficulty"`
	Points     int     `db:"points" json:"points"`
	Free       bool    `db:"free" json:"free"`
	Tags       string  `db:"tags" json:"tags"`
	Prompt     string  `db:"prompt" json:"prompt"`
	Videos     string  `db:"videos" json:"videos"`
	Starter    string  `db:"starter" json:"starter"`
	Tasks      string  `db:"tasks" json:"tasks"`
	Hints      string  `db:"hints" json:"hints"`
	Code       *string `db:"code" json:"code"`
	Solution   string  `db:"solution" json:"solution"`
	Test       string  `db:"test" json:"test"`
}
