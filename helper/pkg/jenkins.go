package pkg

type JobParams struct {
	TestFileOrFolder string   `json:"test_file_or_folder"`
	Shards           string   `json:"shards"`
	Ids              []string `json:"ids"`
	CiEnv            string   `json:"ci_env"`
	Branch           string   `json:"branch"`
	Mode             string   `json:"mode"`
	RunUnit          string   `json:"run_unit"`
	RunValue         int64    `json:"run_value"`
	Crontab          string   `json:"crontab"`
	JobID            int64    `json:"job_id"`
	RunGroupID       string   `json:"run_group_id"`
	SlackID          string   `json:"slack_id"`
	JobName          string   `json:"job_name"`
	Email            string   `json:"email"`
}

// these structs are for working with jenkins
type JkParams struct {
	Parameters []JkParamKV `json:"parameter"`
}

type JkParamKV struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
