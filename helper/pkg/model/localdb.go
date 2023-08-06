package model

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"autopilot-helper/helper/pkg/identity"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var schema = `
CREATE TABLE jobs (
	id    INTEGER PRIMARY KEY,
  th_job_id INTEGER  DEFAULT 0,
  internal_id INTEGER DEFAULT 0,
	service_id INTEGER DEFAULT -1,
	enabled INTEGER DEFAULT 0,
	last_updated_at INTEGER DEFAULT 0
);
`

type JobInfo struct {
	ID            int   `db:"id"`
	THJobID       int   `db:"th_job_id"`
	InternalID    int   `db:"internal_id"`
	ServiceID     int   `db:"service_id"`
	Enabled       int   `db:"enabled"`
	LastUpdatedAt int64 `db:"last_updated_at"`
}

func (t *JobInfo) String() string {
	b, _ := json.Marshal(t)
	return string(b)
}

type Manager struct {
	db *sqlx.DB
}

func NewJobManager() *Manager {
	db, err := sqlx.Connect("sqlite3", fmt.Sprintf("./data/helper_%v_%v.db", time.Now().Format("20060102T15_04"), identity.ServiceID))
	if err != nil {
		log.Fatalln(err)
	}
	db.MustExec(schema)
	return &Manager{db: db}
}

func (m *Manager) AddNewJob(job *JobInfo) error {
	res := m.db.MustExec("INSERT INTO jobs (th_job_id, internal_id, enabled, service_id, last_updated_at) VALUES ($1, $2, $3, $4, $5)",
		job.THJobID, job.InternalID, job.Enabled, identity.ServiceID, job.LastUpdatedAt)
	n, err := res.RowsAffected()
	if n <= 0 || err != nil {
		fmt.Println("AddNewJob - Cannot insert into db", err)
	}
	return err
}

func (m *Manager) ResetEnabledForAllJobs() error {
	res := m.db.MustExec("UPDATE jobs SET enabled = 0 WHERE service_id = $1", identity.ServiceID)
	n, err := res.RowsAffected()
	if err != nil {
		fmt.Println("ResetEnabledForAllJobs - Cannot reset enabled for all jobs", err)
	}
	if n <= 0 {
		fmt.Println("ResetEnabledForAllJobs - We may have no enabled jobs: ", n)
	}
	return err
}

func (m *Manager) UpdateEnabled(thJobID int64, enabled int64) error {
	res := m.db.MustExec("UPDATE jobs SET enabled = $1 WHERE th_job_id = $2 AND service_id = $3", enabled, thJobID, identity.ServiceID)
	n, err := res.RowsAffected()
	if err != nil {
		fmt.Println("UpdateEnabled - Cannot update enabled jobs: ", err)
	}
	if n <= 0 {
		fmt.Println("UpdateEnabled - We may have no enabled jobs: ", n)
	}
	return err
}

func (m *Manager) RemoveJobByEnabled(enabled int64) error {
	res := m.db.MustExec("DELETE FROM jobs WHERE enabled = $1 AND service_id = $2", enabled, identity.ServiceID)
	n, err := res.RowsAffected()
	if err != nil {
		fmt.Println("RemoveJobByEnabled - Cannot remove all jobs by enabled: ", enabled, err)
	}
	if n <= 0 {
		fmt.Println("RemoveJobByEnabled - We may have no enabled jobs: ", n)
	}
	return err
}

func (m *Manager) GetAllDisabledJobs() ([]int, error) {
	var cronIds []int
	err := m.db.Select(&cronIds, "SELECT internal_id FROM jobs WHERE enabled = $1 AND service_id = $2", 0, identity.ServiceID)
	if err != nil {
		fmt.Println("GetAllDisabledJobs - Cannot get all disabled jobs: ", err)
		return nil, err
	}
	return cronIds, err
}

func (m *Manager) RemoveJob(thJobID int64) error {
	res := m.db.MustExec("DELETE FROM jobs WHERE th_job_id = $1 AND service_id = $2", thJobID, identity.ServiceID)
	n, err := res.RowsAffected()
	if err != nil {
		fmt.Println("Cannot remove job from db: ", thJobID, err)
	}
	if n <= 0 {
		fmt.Println("RemoveJob - We may have no enabled jobs: ", n)
	}
	return err
}

func (m *Manager) ExistsJobByTH(thJobID int64) (bool, *JobInfo, error) {
	var ji JobInfo
	err := m.db.Get(&ji, "SELECT * FROM jobs WHERE th_job_id = $1 AND service_id = $2 LIMIT 1", thJobID, identity.ServiceID)
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			return false, nil, nil
		}
		fmt.Println("ExistsJobByTH - Cannot count from db: ", err)
		return false, nil, err
	}
	return ji.ID != 0, &ji, nil
}

func (m *Manager) GetAllJobs() ([]*JobInfo, error) {
	var jobs []*JobInfo
	err := m.db.Get(&jobs, "SELECT * FROM jobs WHERE service_id = $1", identity.ServiceID)
	if err != nil {
		fmt.Println("GetAllJobs - Cannot count from db: ", err)
		return nil, err
	}
	return jobs, nil
}

func (m *Manager) ExistsJobByInternalID(internalID int64) (bool, error) {
	var n int
	err := m.db.Get(&n, "SELECT COUNT(id) FROM jobs WHERE internal_id = $1 AND service_id = $2 LIMIT 1", internalID, identity.ServiceID)
	if err != nil {
		fmt.Println("ExistsJobByInternalID - Cannot count from db: ", err)
		return false, err
	}
	return n > 0, nil
}
