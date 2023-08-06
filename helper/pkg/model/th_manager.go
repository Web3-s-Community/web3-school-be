package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type THManager struct {
	ConnectionString string `json:"connection_string"`
	db               *sqlx.DB
}

func NewTHManagerWDB(db *sqlx.DB) *THManager {
	return &THManager{
		db: db,
	}
}

func NewTHManager(connStr string) *THManager {
	// db, err := sqlx.Connect("mysql", "test:test@(localhost:3306)/test")
	db, err := sqlx.Connect("mysql", connStr)
	if err != nil {
		log.Fatalln(err)
	}
	return &THManager{
		db: db,
	}
}

func (m *THManager) ExistsTHCaseByID(thID int64) (bool, error) {
	var id int
	err := m.db.Get(&id, "SELECT COUNT(id) FROM th_job WHERE id = ? and enabled = true", thID)
	if err != nil {
		fmt.Println("Cannot get from testhub: ", err)
		return false, err
	}
	return id > 0, nil
}

type NullTime struct {
	time.Time
	Valid bool
}

func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}

	if v, ok := value.([]uint8); ok {
		t, err := time.Parse("2006-01-02 15:04:05", string(v))
		if err != nil {
			return err
		}
		nt.Time, nt.Valid = t, true
		return nil
	}

	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type *NullTime", value)
}

func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

type THJob struct {
	ID               int64     `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	Crontab          string    `json:"crontab" db:"crontab"`
	Enabled          bool      `json:"enabled" db:"enabled"`
	Mode             string    `json:"mode" db:"mode"`
	Branch           *string   `json:"branch" db:"branch"`
	Env              *string   `json:"env" db:"env"`
	CreatedBy        *int64    `json:"created_by_id" db:"created_by_id"`
	CallingCode      *string   `json:"calling_code" db:"calling_code"`
	CreatedBySlackID *string   `json:"created_by_slack_id" db:"created_by_slack_id"`
	CreatedByEmail   *string   `json:"created_by_email" db:"created_by_email"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

func (t *THJob) String() string {
	b, _ := json.Marshal(t)
	return string(b)
}

func (m *THManager) GetAllEnabledJobs() ([]*THJob, error) {
	var thJobs []*THJob
	err := m.db.Select(&thJobs, "SELECT id, name, crontab, enabled, mode, branch, env, created_by_id, created_by_slack_id, created_by_email, updated_at, calling_code FROM th_job WHERE enabled = true and mode = 'multiple-time';")
	// err := m.db.Select(&thJobs, "SELECT id, name, crontab, enabled, mode, branch, env, created_by_id, created_by_slack_id, created_by_email, updated_at FROM th_job WHERE name like '%_RC_Jun_%';")
	if err != nil {
		fmt.Println("Cannot get enabled jobs from testhub: ", err)
		return nil, err
	}
	return thJobs, nil
}

type ThSystemConfig struct {
	ID    int64  `json:"id" db:"id"`
	Key   string `json:"key" db:"key"`
	Value string `json:"value" db:"value"`
}

func (m *THManager) GetSystemDisabledJobEnvConfig() (*ThSystemConfig, error) {
	key := "disable_scheduler_env"
	config := &ThSystemConfig{}
	err := m.db.Get(config, "SELECT id, th_systemconfig.key, th_systemconfig.value FROM th_systemconfig WHERE th_systemconfig.key = ?", key)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Not found system config w key: " + key)
			return nil, nil
		}
		fmt.Println("Cannot get system config from testhub: ", err)
		return nil, err
	}
	return config, nil
}

type UserInfo struct {
	ID       int64   `json:"id" db:"id"`
	UserName *string `json:"username" db:"username"`
	Email    *string `json:"email" db:"email"`
	TeamId   *int64  `json:"team_id" db:"team_id"`
	IsActive *int64  `json:"is_active" db:"is_active"`
	SlackId  *string `json:"slack_id" db:"slack_id"`
}

func (m *THManager) GetUserByEmail(email string) (*UserInfo, error) {
	user := &UserInfo{}
	err := m.db.Get(user, "SELECT id, username, email, team_id, is_active, slack_id FROM auth_user WHERE username = ? LIMIT 1", email)
	if err != nil {
		fmt.Println("Cannot get user by email from testhub: ", err)
		return nil, err
	}
	return user, nil
}

func (m *THManager) GetUserById(id int64) (*UserInfo, error) {
	user := &UserInfo{}
	err := m.db.Get(user, "SELECT id, username, email, team_id, is_active, slack_id FROM auth_user WHERE id = ? LIMIT 1", id)
	if err != nil {
		fmt.Println("Cannot get user by email from testhub: ", err)
		return nil, err
	}
	return user, nil
}

func (m *THManager) GetRunGroupInfo(rgId int64) (*RunGroupResult, error) {
	rg := &RunGroupResult{}
	err := m.db.Get(rg, "SELECT * FROM auth_user WHERE id = ? LIMIT 1", rgId)
	if err != nil {
		fmt.Println("Cannot get run group from testhub: ", err)
		return nil, err
	}
	return rg, nil
}

func (m *THManager) GetRunGroupJob(rgId int64) (*RunGroupJob, error) {
	rg := &RunGroupJob{}
	err := m.db.Get(rg, "SELECT rg.id AS rungroup_id, tj.id AS job_id, tj.name as job_name, tj.calling_code AS job_calling_code FROM th_rungroup rg JOIN th_job tj ON rg.job_id = tj.id WHERE rg.id = ?", rgId)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}
	return rg, nil
}

func (m *THManager) GetJobInfo(jobId int64) (*THJob, error) {
	job := &THJob{}
	err := m.db.Get(job, "SELECT id, name, crontab, enabled, mode, branch, env, created_by_id, created_by_slack_id, created_by_email, updated_at FROM th_job WHERE id = ? LIMIT 1", jobId)
	if err != nil {
		fmt.Println("Cannot get run group from testhub: ", err)
		return nil, err
	}
	return job, nil
}

func (m *THManager) GetCasesByJobID(thJobID int64) ([]string, error) {
	var caseNames []string
	err := m.db.Select(&caseNames, "SELECT c.code FROM th_case c INNER JOIN th_case_jobs tj ON c.id = tj.case_id WHERE tj.job_id = ?", thJobID)
	if err != nil {
		fmt.Println("Cannot get cases from testhub: ", err)
		return nil, err
	}
	return caseNames, nil
}

type RunGroupMetaData struct {
	SlackThread  *string `db:"slack_thread" json:"slack_thread,omitempty"`
	Env          *string `db:"env" json:"env,omitempty"`
	UserId       *int64  `db:"user_id" json:"user_id,omitempty"`
	SlackId      *string `db:"slack_id" json:"slack_id,omitempty"`
	TeamId       *int64  `db:"team_id" json:"team_id,omitempty"`
	SlackChannel *string `db:"slack_channel" json:"slack_channel,omitempty"`
	LeadId       *int64  `db:"lead_id" json:"lead_id,omitempty"`
	JobId        *int64  `db:"job_id" json:"job_id,omitempty"`
	JobName      *string `db:"job_name" json:"job_name,omitempty"`
	RunMode      *string `db:"run_mode" json:"run_mode,omitempty"`
}

func (m *THManager) GetRunGroupMetaDataFromRunGroupId(runGroupId int64) (*RunGroupMetaData, error) {
	userTeam := &RunGroupMetaData{}
	err := m.db.Get(userTeam, `
  SELECT
		rg.slack_thread as slack_thread,
		rg.env AS env,
		rg.user_id AS user_id,
		au.slack_id AS slack_id,
		t.id AS team_id,
		t.slack_chann AS slack_channel,
		t.lead_id AS lead_id,
		j.name AS job_name,
		j.mode AS run_mode,
        j.id as job_id
	FROM
		th_rungroup rg
		LEFT JOIN auth_user au ON au.id = rg.user_id
		LEFT JOIN th_team t ON t.id = au.team_id
		LEFT JOIN th_job j ON j.id = rg.job_id
	WHERE
		rg.id = ?`, runGroupId)
	if err != nil {
		fmt.Println("Cannot get user and team from run group: ", runGroupId, err)
		return nil, err
	}
	return userTeam, nil
}

/*
 * Serve for both insert & update using upsert
 */
func (m *THManager) CreateRunResult(r *RunResult) (int64, error) {
	if r.CodeName == "" {
		return -1, errors.New("empty case code name")
	}
	var caseId int64
	err := m.db.Get(&caseId, "SELECT id FROM th_case WHERE code = ?", r.CodeName)
	if err != nil {
		return -1, errors.New("wrong case code name: " + r.CodeName)
	}
	if r.Id == 0 { // just that
		if res, err := m.db.Exec(fmt.Sprintf("INSERT INTO %v (case_id, type, env, report_url, started_at, finished_at, result, run_group_id, test_result, run_url, test_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", "th_run"), caseId, r.Type, r.Env, r.ReportUrl, r.StartedAt, r.FinishedAt, r.Result, r.RunGroupId, r.TestResult, r.BuildUrl, r.TestId); err == nil {
			// insert into test case
			if runId, err2 := res.LastInsertId(); err2 == nil && runId > 0 && r.Env != "local" {
				_, err4 := m.db.Exec(fmt.Sprintf("UPDATE %v SET report_url = ? WHERE id = ?", "th_run"), fmt.Sprintf("https://report-autopilot.shopbase.dev/show?run_id=%v#?testId=%v", runId, r.TestId), runId)
				if err4 != nil {
					fmt.Printf("Update report url for run id %v error\n", runId)
				}
				if r.Result == "pass" {
					if _, err3 := m.db.Exec(fmt.Sprintf("UPDATE %v SET last_run_%v_id = ?, last_run_success_%v_id = ?, is_automated_%v = ?, need_automation = ? WHERE id = ?", "th_case", r.Env, r.Env, r.Env), runId, runId, true, true, caseId); err3 == nil {
						return res.LastInsertId()
					} else {
						return 0, err3
					}
				} else {
					if _, err3 := m.db.Exec(fmt.Sprintf("UPDATE %v SET last_run_%v_id = ?, is_automated_%v = ? WHERE id = ?", "th_case", r.Env, r.Env), runId, false, caseId); err3 == nil {
						return res.LastInsertId()
					} else {
						return 0, err3
					}
				}

			} else {
				return 0, err2
			}
		} else {
			return 0, err
		}
	} else {
		if res, err := m.db.Exec(fmt.Sprintf("UPDATE %v SET case_id = ?, type = ?, env = ?, report_url = ?, started_at = ?, finished_at = ?, result = ?, run_group_id = ?, test_result = ?, run_url = ?, test_id = ?, need_automation = ? WHERE id = ?", "th_run"), caseId, r.Type, r.Env, r.ReportUrl, r.StartedAt, r.FinishedAt, r.Result, r.RunGroupId, r.TestResult, r.BuildUrl, r.TestId, true, r.Id); err == nil {
			return res.RowsAffected()
		} else {
			return 0, err
		}
	}
}

func (m *THManager) CreateRunResults(runs []*RunResult) (int64, int64, int64, error) {
	success := int64(0)
	fail := int64(0)
	var lastError error
	for _, run := range runs {
		if _, err := m.CreateRunResult(run); err == nil {
			success++
		} else {
			lastError = err
			fmt.Println("Cannot create new run result: ", lastError)
			fail++
		}
	}
	return int64(len(runs)), success, fail, lastError
}

func (m *THManager) GetFailRunInfo(runGroupId int64) ([]*ThRunInfo, error) {
	var runs []*ThRunInfo
	err := m.db.Select(&runs, "SELECT tr.id AS run_id, tr.case_id, tr.report_url, tr.env, tr.result, tc.code AS case_code FROM th_run tr JOIN th_case tc ON tr.case_id = tc.id WHERE tr.run_group_id = ? AND result = 'fail';", runGroupId)
	if err != nil {
		fmt.Println("Cannot get fail runs info by rungroup id from testhub: ", err)
		return nil, err
	}

	return runs, nil
}

/*
 * Serve for both insert & update using upsert
 */
func (m *THManager) CreateRunGroupResult(r *RunGroupResultRequest) (int64, error) {
	if r.Id == 0 { // just that,
		// in-case no job was found
		if res, err := m.db.Exec(fmt.Sprintf("INSERT INTO %v (name, env, started_at, finished_at, result, test_result) VALUES (?, ?, ?, ?, ?, ?)", "th_rungroup"), generateRunGroupName(&r.Env, "", r.RunningUserEmail, r.Name), r.Env, r.StartedAt, r.FinishedAt, r.Result, r.TestResult); err == nil {
			runGroupId, runGrErr := res.LastInsertId()
			if r.RunningUserEmail != "" && runGrErr == nil && runGroupId > 0 {
				if user, err := m.GetUserByEmail(r.RunningUserEmail); err == nil {
					if _, err := m.db.Exec(fmt.Sprintf("UPDATE %v SET user_id = ? WHERE id = ?", "th_rungroup"), user.ID, runGroupId); err != nil {
						fmt.Println("Cannot update user_id in rungroup", err)
					}
				} else {
					fmt.Println("Cannot get user by email or cannot create run group: ", runGroupId, runGrErr, err)
				}
			}
			return runGroupId, runGrErr
		} else {
			return 0, err
		}
	} else {
		if res, err := m.db.Exec(fmt.Sprintf("UPDATE %v SET started_at = ?, finished_at = ?, result = ?, test_result = ? WHERE id = ?", "th_rungroup"), r.StartedAt, r.FinishedAt, r.Result, r.TestResult, r.Id); err == nil {
			return res.RowsAffected()
		} else {
			return 0, err
		}
	}
}

func (m *THManager) UpdateSlackThreadForRunGroup(slackThreadURL string, runGroupId int64) (bool, error) {
	res, err := m.db.Exec("UPDATE th_rungroup SET slack_thread = ? WHERE id = ?", slackThreadURL, runGroupId)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	return rows > 0, err
}

func (m *THManager) UpdateRunGroupJobIdToNull(jobId int64) (int64, error) {
	res, err := m.db.Exec("UPDATE th_rungroup set job_id = null where job_id = ?", jobId)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	return rows, err
}

func (m *THManager) CreateRunGroup(thJobID int64, buildUserEmail string, buildUserId int64) (int64, error) {
	thJob := &THJob{}
	err := m.db.Get(thJob, "SELECT * FROM th_job WHERE id = ?", thJobID)
	switch err {
	case sql.ErrNoRows:
		res, err := m.db.Exec("INSERT INTO th_rungroup (env, started_at, name, user_id) VALUES (?, ?, ?, ?)", thJob.Env, time.Now(), generateRunGroupName(thJob.Env, thJob.Mode, buildUserEmail, thJob.Name), buildUserId)
		if err != nil {
			fmt.Println("Cannot insert new rungroup into testhub: ", err)
			return -1, err
		}
		if id, err := res.LastInsertId(); err != nil {
			fmt.Println("Cannot get newly inserted rungroup into testhub: ", err)
			return -1, err
		} else {
			return id, nil
		}
	case nil:
		res, err := m.db.Exec("INSERT INTO th_rungroup (env, started_at, name, job_id, user_id) VALUES (?, ?, ?, ?, ?)", thJob.Env, time.Now(), generateRunGroupName(thJob.Env, thJob.Mode, *thJob.CreatedByEmail, thJob.Name), thJobID, thJob.CreatedBy)
		if err != nil {
			fmt.Println("Cannot insert new rungroup into testhub: ", err)
			return -1, err
		}
		if id, err := res.LastInsertId(); err != nil {
			fmt.Println("Cannot get newly inserted rungroup into testhub: ", err)
			return -1, err
		} else {
			return id, nil
		}
	default:
		return -1, err
	}

}

/*
 * Serve for scheduler
 */

type THCase struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type THCaseData struct {
	ID        int64  `json:"id" db:"id"`
	Env       string `json:"env" db:"env"`
	CaseCode  string `json:"case_code" db:"case_code"`
	Data      string `json:"data" db:"data"`
	CaseId    int64  `json:"case_id" db:"case_id"`
	Branch    string `json:"branch" db:"branch"`
	CreatedAt string `json:"created_at" db:"created_at"`
}

type ThRunInfo struct {
	RunID     int64   `json:"run_id" db:"run_id"`
	CaseID    int64   `json:"case_id" db:"case_id"`
	ReportURL *string `json:"report_url" db:"report_url"`
	Env       string  `json:"env" db:"env"`
	Result    string  `json:"result" db:"result"`
	CaseCode  string  `json:"case_code" db:"case_code"`
}

func (m *THManager) GetCaseByCode(code string) (*THCase, error) {
	thCase := &THCase{}
	err := m.db.Get(thCase, "SELECT id, name FROM th_case WHERE code = ?", code)
	if err != nil {
		fmt.Println("Cannot get case from testhub: ", err)
		return nil, err
	}
	return thCase, nil
}

func (m *THManager) UpdateOutdatedAutomationCase(env string) (int64, error) {
	if res, err := m.db.Exec(fmt.Sprintf("UPDATE th_case SET is_automated_%v = false, is_automated = false WHERE id IN ( SELECT id FROM (SELECT tc.id FROM th_case tc JOIN th_run tr ON tc.last_run_%v_id = tr.id WHERE tc.is_automated_%v = true AND tr.finished_at <= DATE_SUB(NOW(), INTERVAL 7 DAY) AND tr.type = 'automation') as subquery); ", env, env, env)); err == nil {
		return res.RowsAffected()
	} else {
		return 0, err
	}
}

func (m *THManager) GetCaseData(caseId int64, branch string, env string) (*THCaseData, error) {
	thCaseData := &THCaseData{}
	err := m.db.Get(thCaseData, "SELECT id, data FROM th_casedata WHERE case_id = ? and branch = ? and env = ?", caseId, branch, env)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			return nil, nil
		}
		fmt.Println("Cannot get case data from testhub: ", err)
		return nil, err
	}
	return thCaseData, nil
}

func (m *THManager) UpdateCaseData(caseDataReq *UpdateCaseDataRequest) (int64, error) {
	if res, err := m.db.Exec(fmt.Sprintf("UPDATE th_casedata SET data = ? where id = ?"), caseDataReq.Data, caseDataReq.Id); err == nil {
		return res.RowsAffected()
	} else {
		return 0, err
	}

}

func (m *THManager) InsertCaseData(caseDataReq *InsertCaseDataRequest) (int64, error) {
	res, err := m.db.Exec(fmt.Sprintf("INSERT INTO th_casedata (env, case_code, data, created_at, case_id, branch) VALUES (?, ?, ?, ?, ?, ?)"),
		caseDataReq.Env,
		caseDataReq.CaseCode,
		caseDataReq.Data,
		time.Now().Format("2006-01-02 15:04:05"),
		caseDataReq.CaseId,
		caseDataReq.Branch,
	)

	if err != nil {
		return 0, err
	}

	if id, err := res.LastInsertId(); err != nil {
		fmt.Println("Cannot get newly inserted case data into testhub: ", err)
		return -1, err
	} else {
		return id, nil
	}
}

func (m *THManager) GetJob(jobName string, branch string, env string) (*THJob, error) {
	thJob := &THJob{}
	err := m.db.Get(thJob, "SELECT id, name, crontab, branch, env FROM th_job WHERE name = ? and branch = ? and env = ?", jobName, branch, env)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			return nil, nil
		}

		fmt.Println("Cannot get job from testhub: ", err)
		return nil, err
	}
	return thJob, nil
}

func (m *THManager) UpdateJob(updateJobReq *UpdateJobRequest) (int64, error) {
	if res, err := m.db.Exec(fmt.Sprintf("UPDATE th_job SET crontab = ?, updated_at = NOW() where id = ?"), updateJobReq.Crontab, updateJobReq.Id); err == nil {
		return res.RowsAffected()
	} else {
		return 0, err
	}
}

func (m *THManager) InsertJob(insertJobReq *InsertJobRequest) (int64, error) {
	res, err := m.db.Exec(fmt.Sprintf("INSERT INTO th_job (name, crontab, enabled, mode, env, created_by_id, created_by_slack_id, created_by_email, updated_at, branch) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"),
		insertJobReq.Name,
		insertJobReq.Crontab,
		insertJobReq.Enabled,
		insertJobReq.Mode,
		insertJobReq.Env,
		insertJobReq.CreatedById,
		insertJobReq.CreatedBySlackId,
		insertJobReq.CreatedByEmail,
		time.Now().Format("2006-01-02 15:04:05"),
		insertJobReq.Branch,
	)

	if err != nil {
		return 0, err
	}

	if id, err := res.LastInsertId(); err != nil {
		fmt.Println("Cannot get newly inserted case data into testhub: ", err)
		return -1, err
	} else {
		return id, nil
	}
}

func (m *THManager) RemoveJob(removeReq *RemoveJobRequest) (int64, error) {
	if res, err := m.db.Exec(fmt.Sprintf("DELETE FROM th_job  where id = ?"), removeReq.Id); err == nil {
		return res.RowsAffected()
	} else {
		return 0, err
	}
}

func (m *THManager) InsertCaseJob(insertReq *InsertCaseJobRequest) (int64, error) {
	res, err := m.db.Exec(fmt.Sprintf("INSERT INTO th_case_jobs (case_id, job_id) VALUES (?, ?)"),
		insertReq.CaseId,
		insertReq.JobId,
	)

	if err != nil {
		return 0, err
	}

	if id, err := res.LastInsertId(); err != nil {
		fmt.Println("Cannot get newly inserted case data into testhub: ", err)
		return -1, err
	} else {
		return id, nil
	}
}

func (m *THManager) RemoveCaseJob(removeReq *RemoveCaseJobRequest) (int64, error) {
	if res, err := m.db.Exec(fmt.Sprintf("DELETE FROM th_case_jobs  where case_id = ? and job_id = ?"),
		removeReq.CaseId,
		removeReq.JobId); err == nil {
		return res.RowsAffected()
	} else {
		return 0, err
	}
}

func (m *THManager) RemoveCaseData(removeReq *RemoveCaseDataRequest) (int64, error) {
	if res, err := m.db.Exec(fmt.Sprintf("DELETE FROM th_casedata where case_id = ? and branch = ? and env = ?"),
		removeReq.CaseId,
		removeReq.Branch,
		removeReq.Env,
	); err == nil {
		return res.RowsAffected()
	} else {
		return 0, err
	}
}

var unknownEnv = "unknown"

func generateRunGroupName(env *string, runningMode, email, name string) string {
	mode := "mul"
	if runningMode != "multiple-time" {
		mode = "one"
	}
	if email == "" {
		email = "unknown_user"
	}
	if name == "" {
		name = RandString(6)
	}
	if env == nil {

		env = &unknownEnv
	}
	return fmt.Sprintf("auto-%v-%v-%v-%v-%v", *env, mode, email, name, RandString(10))
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
