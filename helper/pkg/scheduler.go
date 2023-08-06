package pkg

import (
	"autopilot-helper/helper/pkg/identity"
	"autopilot-helper/helper/pkg/model"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type CreateScheduledRequestParams struct {
	THJobId             int64
	THJobName           string
	THJobRunMode        string
	THJobCreatedByEmail string
	THJobCallingCode    string

	JobBranch        string
	RunBranch        string
	Shards           string
	JenkinsToken     string
	RequesterSlackId *string
	Env              string

	CaseNames []string
	Crontab   string

	DisabledJobEnv string
}

/**
* This is for scheduling tasks.
* At the beginning, The scheduler:
* 1. Create a new sqlite3 database for each run (don't re-use old db files because we don't need to)
* 2. Initialize a task itself to sync jobs from test-hub db to it's sqlite3 database in every minute
* 3. Schedule and run scheduled jobs
 */
type TaskScheduler struct {
	jkBuildToken string
	scheduler    *cron.Cron
	lock         sync.Mutex
	jobManager   *model.Manager
	thManager    *model.THManager
	jobBranch    string
	Notificator  *model.Notificator
}

func NewTaskScheduler(jobBranch string, thManager *model.THManager, notificator *model.Notificator) *TaskScheduler {
	jkToken := strings.TrimSpace(os.Getenv("JK_JOBBUILD_TOKEN"))
	return &TaskScheduler{
		jkBuildToken: jkToken,
		scheduler:    cron.New(),
		jobManager:   model.NewJobManager(),
		thManager:    thManager,
		jobBranch:    jobBranch,
	}
}

// Start scheduler and start a sync job to sync with testhub db
func (s *TaskScheduler) Start() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.scheduler == nil {
		s.scheduler = cron.New()
	}
	s.scheduler.Start()
	// add default tasks
	s.scheduler.AddFunc("* * * * *", s.SyncTaskWithDB)
}

func (s *TaskScheduler) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.scheduler != nil {
		s.scheduler.Stop()
	}
}

// Add a new job/task
func (s *TaskScheduler) AddTask(cron string, task func()) (cron.EntryID, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.scheduler != nil {
		id, err := s.scheduler.AddFunc(cron, task)
		if err != nil {
			fmt.Println("Cannot add new function:", err)
		}
		return id, err
	}
	return 0, errors.New("nil scheduler")
}

func (s *TaskScheduler) RemoveTask(taskID cron.EntryID) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.scheduler != nil {
		s.scheduler.Remove(taskID)
		return nil
	}
	return errors.New("nil scheduler")
}

// SyncTaskWithDB runs every minute to sync with testhub
// 1. Get all enabled jobs from testhub with `multiple-time` mode and enabled = true
// 2. Lock the process
// 3. Reset `enabled` field of all jobs in it's sqlite3 db to false (0)
// 4. For each testhub jobs:
// 4.1. Check if the job has already scheduled?
//
//	4.1.1. If true --> update enable = true (1) on sqlite3 db
//	4.1.2. If false --> create new scheduled job, schedule it and add to sqlite3 db
//
// 5. Get all remaining jobs from sqlite3
// 6. Remove them
func (s *TaskScheduler) SyncTaskWithDB() {
	fmt.Println("[Scheduling] Running Sync Task With DB at: ", time.Now().Format(time.RFC3339))
	allTHJobs, err := s.thManager.GetAllEnabledJobs()
	if err != nil {
		fmt.Println("Cannot get all jobs from testhub:", err)
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.jobManager.ResetEnabledForAllJobs(); err != nil {
		fmt.Println("Cannot reset enabled jobs:", err)
		return
	}

	defaultBranch := "master"
	defaultCallingCode := ""
	disabledEnv := s.getDisabledEnv()
	for _, thJob := range allTHJobs {
		//fmt.Println("Checking job: ", thJob)
		if thJob.Branch == nil {
			thJob.Branch = &defaultBranch
		}

		if thJob.CallingCode == nil {
			thJob.CallingCode = &defaultCallingCode
		}

		existed, ji, err := s.jobManager.ExistsJobByTH(thJob.ID)
		if err == nil {
			if !existed {
				// if not existed -> add
				if cases, err := s.thManager.GetCasesByJobID(thJob.ID); err == nil {
					params := &CreateScheduledRequestParams{
						THJobId:             thJob.ID,
						THJobName:           thJob.Name,
						THJobRunMode:        thJob.Mode,
						THJobCreatedByEmail: *thJob.CreatedByEmail,
						THJobCallingCode:    *thJob.CallingCode,

						JobBranch:        s.jobBranch,
						RunBranch:        *thJob.Branch,
						Shards:           "1",
						JenkinsToken:     s.jkBuildToken,
						RequesterSlackId: thJob.CreatedBySlackID,
						Env:              *thJob.Env,
						CaseNames:        cases,
						Crontab:          thJob.Crontab,
						DisabledJobEnv:   disabledEnv,
					}

					internalID, err := s.createScheduledRequest(params)
					if err == nil {
						if internalID > 0 {
							s.jobManager.AddNewJob(&model.JobInfo{
								THJobID:       int(thJob.ID),
								InternalID:    int(internalID),
								Enabled:       1,
								ServiceID:     identity.ServiceID,
								LastUpdatedAt: thJob.UpdatedAt.Unix(),
							})
						} // Not have id ~> skip
					} else {
						fmt.Println("Cannot create scheduled request: ", err)
					}
				} else {
					fmt.Println("Cannot schedule a job (cannot get cases): ", err)
				}
			} else {
				if thJob.UpdatedAt.Unix() > ji.LastUpdatedAt {
					fmt.Printf("Job %v is updated.\n", thJob.ID)
					// if the job is updated
					s.scheduler.Remove(cron.EntryID(ji.ID))
					if cases, err := s.thManager.GetCasesByJobID(thJob.ID); err == nil {
						params := &CreateScheduledRequestParams{
							THJobId:             thJob.ID,
							THJobName:           thJob.Name,
							THJobRunMode:        thJob.Mode,
							THJobCreatedByEmail: *thJob.CreatedByEmail,
							THJobCallingCode:    *thJob.CallingCode,

							JobBranch:        s.jobBranch,
							RunBranch:        *thJob.Branch,
							Shards:           "1",
							JenkinsToken:     s.jkBuildToken,
							RequesterSlackId: thJob.CreatedBySlackID,
							Env:              *thJob.Env,
							CaseNames:        cases,
							Crontab:          thJob.Crontab,
							DisabledJobEnv:   disabledEnv,
						}

						internalID, err := s.createScheduledRequest(params)
						if err == nil {
							if internalID > 0 {
								s.jobManager.AddNewJob(&model.JobInfo{
									THJobID:       int(thJob.ID),
									InternalID:    int(internalID),
									Enabled:       1,
									ServiceID:     identity.ServiceID,
									LastUpdatedAt: thJob.UpdatedAt.Unix(),
								})
							} // Not have id ~> skip
						}
					} else {
						fmt.Println("Cannot schedule a job (cannot get cases to update): ", err)
					}
				} else {
					if err = s.jobManager.UpdateEnabled(thJob.ID, 1); err != nil {
						fmt.Println("Cannot enable job: ", thJob.ID, err)
					}
				}
			}
		}
	}
	jobs, err := s.jobManager.GetAllDisabledJobs()
	if err != nil {
		fmt.Println("Cannot get all jobs from internal db:", err)
		return
	}
	for _, job := range jobs {
		s.scheduler.Remove(cron.EntryID(job))
	}
	if err = s.jobManager.RemoveJobByEnabled(0); err != nil {
		fmt.Println("Cannot remove jobs from internal db: ", err)
		return
	}
}

func (s *TaskScheduler) createScheduledRequest(params *CreateScheduledRequestParams) (cron.EntryID, error) {
	if params.Env == params.DisabledJobEnv && !strings.HasPrefix(params.THJobName, "scheduled_job_") {
		fmt.Printf("Skip job %v (%v) due to disable config on env %v\n", params.THJobName, params.THJobId, params.DisabledJobEnv)
		return 0, nil
	}

	jkUrl := os.Getenv("JK_BUILD_URL")
	if jkUrl == "" {
		panic("Cannot load jenkins build params url")
	}
	jkUrl = strings.Replace(jkUrl, "BRANCH", params.JobBranch, 1)

	data := url.Values{}
	jkParams := &JkParams{}

	param := "TEST_FILE_OR_FOLDER"
	jkParams.Parameters = append(jkParams.Parameters, JkParamKV{Name: param, Value: "./tests"})

	param = "SHARDS"
	jkParams.Parameters = append(jkParams.Parameters, JkParamKV{Name: param, Value: params.Shards})

	param = "BRANCH"
	jkParams.Parameters = append(jkParams.Parameters, JkParamKV{Name: param, Value: params.RunBranch})

	param = "SUITE_ID_OR_CASE_ID"
	if v := params.CaseNames; len(v) > 0 {
		var ids []string
		for _, id := range v {
			if !strings.HasPrefix(id, "@TC_") && !strings.HasPrefix(id, "@SB_") &&
				!strings.HasPrefix(id, "@DP_") && !strings.HasPrefix(id, "@PLB_") &&
				!strings.HasPrefix(id, "TC_") && !strings.HasPrefix(id, "SB_") &&
				!strings.HasPrefix(id, "DP_") && !strings.HasPrefix(id, "PLB_") {
				id = "@TC_" + strings.Replace(id, "@", "", 1)
			}
			ids = append(ids, id)
		}
		jkParams.Parameters = append(jkParams.Parameters, JkParamKV{Name: param, Value: "(" + strings.Join(ids, "|") + ")"})
	}

	param = "CI_ENV"
	jkParams.Parameters = append(jkParams.Parameters, JkParamKV{Name: param, Value: params.Env})

	cronID, err := s.scheduler.AddFunc(params.Crontab, func() {
		forRequestJkParams := &JkParams{}
		// create run group
		runGroupId, err := s.thManager.CreateRunGroup(params.THJobId, "", 0)
		if err == nil && runGroupId > 0 {
			param = "TH_RUN_GROUP_ID"
			forRequestJkParams.Parameters = append(jkParams.Parameters, JkParamKV{Name: param, Value: fmt.Sprintf("%v", runGroupId)})
		} else {
			fmt.Println("Cannot create run group: ", err)
		}

		b, _ := json.Marshal(forRequestJkParams)
		data.Set("json", string(b))
		data.Set("token", params.JenkinsToken)
		jkReq, err := http.NewRequest(http.MethodPost, jkUrl, strings.NewReader(data.Encode()))
		if err != nil {
			fmt.Println("Cannot create new request: ", err)
			return
		}
		auth := os.Getenv("JK_AUTH")

		b, _ = httputil.DumpRequest(jkReq, true)
		fmt.Println("Created request for: ", params.THJobId, "with run group id: ", runGroupId, ": ", string(b))

		jkReq.Header.Add("Authorization", fmt.Sprintf("Basic %v", auth))
		jkReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		fmt.Printf("[Scheduled] - Running job with Job ID (%v), Run Group Id (%v) at: %v\n", params.THJobId, runGroupId, time.Now().Format("2006-01-02 15:04:05"))
		// request to jenkins
		resp, err := NewClient().Do(jkReq)
		if err != nil {
			fmt.Println("Cannot do request", err)
			return
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			b, _ := httputil.DumpResponse(resp, true)
			fmt.Println("Wrong return code from jenkins: ", string(b))
			notifyMessage := fmt.Sprintf("url: %v, response: %v", jkUrl, string(b))
			_ = s.Notificator.NotifySimple("automation-system-monitor", notifyMessage)
			return
		}
	})
	if err != nil {
		fmt.Println("Cannot create job", err)
	} else {
		fmt.Println("Job is created: ", cronID)
	}
	return cronID, err
}

func (s *TaskScheduler) getDisabledEnv() string {
	disableEnvConfig, err := s.thManager.GetSystemDisabledJobEnvConfig()
	if err != nil {
		fmt.Printf("Error when get system disabled job env config: %v", err.Error())
		return ""
	}

	disabledEnv := disableEnvConfig.Value

	if disabledEnv == "empty" {
		return ""
	}

	return disabledEnv
}
