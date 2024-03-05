package runTask

type RunTaskPayload struct {
	PayloadVersion int    `json:"payload_version"`
	Stage          string `json:"stage"`
	AccessToken    string `json:"access_token"`
	Capabilities   struct {
		Outcomes bool `json:"outcomes"`
	} `json:"capabilities"`
	ConfigurationVersionDownloadURL string `json:"configuration_version_download_url"`
	ConfigurationVersionID          string `json:"configuration_version_id"`
	IsSpeculative                   bool   `json:"is_speculative"`
	OrganizationName                string `json:"organization_name"`
	PlanJsonAPIURL                  string `json:"plan_json_api_url"`
	RunAppURL                       string `json:"run_app_url"`
	RunCreatedAt                    string `json:"run_created_at"`
	RunCreatedBy                    string `json:"run_created_by"`
	RunID                           string `json:"run_id"`
	RunMessage                      string `json:"run_message"`
	TaskResultCallbackURL           string `json:"task_result_callback_url"`
	TaskResultEnforcementLevel      string `json:"task_result_enforcement_level"`
	TaskResultID                    string `json:"task_result_id"`
	VCSBranch                       string `json:"vcs_branch"`
	VCSCommitURL                    string `json:"vcs_commit_url"`
	VCSPullRequestURL               string `json:"vcs_pull_request_url"`
	VCSRepoURL                      string `json:"vcs_repo_url"`
	WorkspaceAppURL                 string `json:"workspace_app_url"`
	WorkspaceID                     string `json:"workspace_id"`
	WorkspaceName                   string `json:"workspace_name"`
	WorkspaceWorkingDirectory       string `json:"workspace_working_directory"`
}

type RunTaskResult struct {
	Data struct {
		Type       string `json:"type"` // Must be "task-results".
		Attributes struct {
			Status  string `json:"status"` // "enum": ["running", "passed", "failed"]
			Message string `json:"message,omitempty"`
			URL     string `json:"url,omitempty"`
		} `json:"attributes"`
		Relationships *RunTaskResultRelationships `json:"relationships,omitempty"`
	} `json:"data"`
}
type RunTaskResultRelationships struct {
	Outcomes struct {
		Data []RunTaskOutcome `json:"data"`
	} `json:"outcomes,omitempty"`
}

type RunTaskOutcome struct {
	Type       string                           `json:"type"` // must be `task-result-outcomes`
	Attributes RunTaskResponseOutcomeAttributes `json:"attributes"`
}

type RunTaskResponseOutcomeAttributes struct {
	OutcomeID   string      `json:"outcome-id"`
	Description string      `json:"description"`
	Tags        RunTaskTags `json:"tags,omitempty"`
	Body        string      `json:"body,omitempty"`
	URL         string      `json:"url,omitempty"`
}

type RunTaskTags struct {
	Status []RunTaskRichLabel `json:"status,omitempty"`
}

type RunTaskRichLabel struct {
	Label string `json:"label"`
	Level string `json:"level,omitempty"` // "enum": ["none", "info", "warning", "error"]
}
