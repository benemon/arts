package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var ansibleHost string
var ansibleUser string
var ansiblePassword string

const (
	Passed  = "passed"
	Failed  = "failed"
	Running = "running"
)

const (
	TaskResults = "task-results"
)

const (
	TestToken = "test-token"
)

type RunTaskRequest struct {
	PayloadVersion                  int       `json:"payload_version,omitempty"`
	AccessToken                     string    `json:"access_token,omitempty"`
	Stage                           string    `json:"stage,omitempty"`
	IsSpeculative                   bool      `json:"is_speculative,omitempty"`
	TaskResultID                    string    `json:"task_result_id,omitempty"`
	TaskResultEnforcementLevel      string    `json:"task_result_enforcement_level,omitempty"`
	TaskResultCallbackURL           string    `json:"task_result_callback_url,omitempty"`
	RunAppURL                       string    `json:"run_app_url,omitempty"`
	RunID                           string    `json:"run_id,omitempty"`
	RunMessage                      string    `json:"run_message,omitempty"`
	RunCreatedAt                    time.Time `json:"run_created_at,omitempty"`
	RunCreatedBy                    string    `json:"run_created_by,omitempty"`
	WorkspaceID                     string    `json:"workspace_id,omitempty"`
	WorkspaceName                   string    `json:"workspace_name,omitempty"`
	WorkspaceAppURL                 string    `json:"workspace_app_url,omitempty"`
	OrganizationName                string    `json:"organization_name,omitempty"`
	PlanJSONAPIURL                  string    `json:"plan_json_api_url,omitempty"`
	VcsRepoURL                      string    `json:"vcs_repo_url,omitempty"`
	VcsBranch                       string    `json:"vcs_branch,omitempty"`
	VcsPullRequestURL               any       `json:"vcs_pull_request_url,omitempty"`
	VcsCommitURL                    string    `json:"vcs_commit_url,omitempty"`
	ConfigurationVersionID          string    `json:"configuration_version_id,omitempty"`
	ConfigurationVersionDownloadURL string    `json:"configuration_version_download_url,omitempty"`
	WorkspaceWorkingDirectory       string    `json:"workspace_working_directory,omitempty"`
}

type RunTaskResponse struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			Status  string `json:"status"`
			Message string `json:"message,omitempty"`
			URL     string `json:"url,omitempty"`
		} `json:"attributes"`
	} `json:"data"`
}

type HMACRequest struct {
	WorkspaceID string `json:"workspace_id"`
}

type HMACResponse struct {
	WorkspaceID string `json:"workspace_id"`
	HMAC        string `json:"hmac"`
}

type AnsibleJobTemplateRequest struct {
	Inventory string         `json:"inventory,omitempty"`
	ExtraVars map[string]any `json:"extra_vars,omitempty"`
}

type AnsibleWorkflowJobTemplateRequest struct {
	AskLimitOnLaunch     bool           `json:"ask_limit_on_launch,omitempty"`
	AskScmBranchOnLaunch bool           `json:"ask_scm_branch_on_launch,omitempty"`
	ExtraVars            map[string]any `json:"extra_vars,omitempty"`
	Inventory            int            `json:"inventory,omitempty"`
	Limit                string         `json:"limit,omitempty"`
	ScmBranch            string         `json:"scm_branch,omitempty"`
}

type AnsibleJobTemplateResponse struct {
	Job           int `json:"job,omitempty"`
	IgnoredFields struct {
	} `json:"ignored_fields,omitempty"`
	ID      int    `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	URL     string `json:"url,omitempty"`
	Related struct {
		CreatedBy          string `json:"created_by,omitempty"`
		ModifiedBy         string `json:"modified_by,omitempty"`
		Labels             string `json:"labels,omitempty"`
		Inventory          string `json:"inventory,omitempty"`
		Project            string `json:"project,omitempty"`
		Organization       string `json:"organization,omitempty"`
		Credentials        string `json:"credentials,omitempty"`
		UnifiedJobTemplate string `json:"unified_job_template,omitempty"`
		Stdout             string `json:"stdout,omitempty"`
		JobEvents          string `json:"job_events,omitempty"`
		JobHostSummaries   string `json:"job_host_summaries,omitempty"`
		ActivityStream     string `json:"activity_stream,omitempty"`
		Notifications      string `json:"notifications,omitempty"`
		CreateSchedule     string `json:"create_schedule,omitempty"`
		JobTemplate        string `json:"job_template,omitempty"`
		Cancel             string `json:"cancel,omitempty"`
		Relaunch           string `json:"relaunch,omitempty"`
	} `json:"related,omitempty"`
	SummaryFields struct {
		Organization struct {
			ID          int    `json:"id,omitempty"`
			Name        string `json:"name,omitempty"`
			Description string `json:"description,omitempty"`
		} `json:"organization,omitempty"`
		Inventory struct {
			ID                           int    `json:"id,omitempty"`
			Name                         string `json:"name,omitempty"`
			Description                  string `json:"description,omitempty"`
			HasActiveFailures            bool   `json:"has_active_failures,omitempty"`
			TotalHosts                   int    `json:"total_hosts,omitempty"`
			HostsWithActiveFailures      int    `json:"hosts_with_active_failures,omitempty"`
			TotalGroups                  int    `json:"total_groups,omitempty"`
			HasInventorySources          bool   `json:"has_inventory_sources,omitempty"`
			TotalInventorySources        int    `json:"total_inventory_sources,omitempty"`
			InventorySourcesWithFailures int    `json:"inventory_sources_with_failures,omitempty"`
			OrganizationID               int    `json:"organization_id,omitempty"`
			Kind                         string `json:"kind,omitempty"`
		} `json:"inventory,omitempty"`
		Project struct {
			ID            int    `json:"id,omitempty"`
			Name          string `json:"name,omitempty"`
			Description   string `json:"description,omitempty"`
			Status        string `json:"status,omitempty"`
			ScmType       string `json:"scm_type,omitempty"`
			AllowOverride bool   `json:"allow_override,omitempty"`
		} `json:"project,omitempty"`
		JobTemplate struct {
			ID          int    `json:"id,omitempty"`
			Name        string `json:"name,omitempty"`
			Description string `json:"description,omitempty"`
		} `json:"job_template,omitempty"`
		UnifiedJobTemplate struct {
			ID             int    `json:"id,omitempty"`
			Name           string `json:"name,omitempty"`
			Description    string `json:"description,omitempty"`
			UnifiedJobType string `json:"unified_job_type,omitempty"`
		} `json:"unified_job_template,omitempty"`
		CreatedBy struct {
			ID        int    `json:"id,omitempty"`
			Username  string `json:"username,omitempty"`
			FirstName string `json:"first_name,omitempty"`
			LastName  string `json:"last_name,omitempty"`
		} `json:"created_by,omitempty"`
		ModifiedBy struct {
			ID        int    `json:"id,omitempty"`
			Username  string `json:"username,omitempty"`
			FirstName string `json:"first_name,omitempty"`
			LastName  string `json:"last_name,omitempty"`
		} `json:"modified_by,omitempty"`
		UserCapabilities struct {
			Delete bool `json:"delete,omitempty"`
			Start  bool `json:"start,omitempty"`
		} `json:"user_capabilities,omitempty"`
		Labels struct {
			Count   int   `json:"count,omitempty"`
			Results []any `json:"results,omitempty"`
		} `json:"labels,omitempty"`
		Credentials []struct {
			ID          int    `json:"id,omitempty"`
			Name        string `json:"name,omitempty"`
			Description string `json:"description,omitempty"`
			Kind        string `json:"kind,omitempty"`
			Cloud       bool   `json:"cloud,omitempty"`
		} `json:"credentials,omitempty"`
	} `json:"summary_fields,omitempty"`
	Created              time.Time `json:"created,omitempty"`
	Modified             time.Time `json:"modified,omitempty"`
	Name                 string    `json:"name,omitempty"`
	Description          string    `json:"description,omitempty"`
	JobType              string    `json:"job_type,omitempty"`
	Inventory            int       `json:"inventory,omitempty"`
	Project              int       `json:"project,omitempty"`
	Playbook             string    `json:"playbook,omitempty"`
	ScmBranch            string    `json:"scm_branch,omitempty"`
	Forks                int       `json:"forks,omitempty"`
	Limit                string    `json:"limit,omitempty"`
	Verbosity            int       `json:"verbosity,omitempty"`
	ExtraVars            string    `json:"extra_vars,omitempty"`
	JobTags              string    `json:"job_tags,omitempty"`
	ForceHandlers        bool      `json:"force_handlers,omitempty"`
	SkipTags             string    `json:"skip_tags,omitempty"`
	StartAtTask          string    `json:"start_at_task,omitempty"`
	Timeout              int       `json:"timeout,omitempty"`
	UseFactCache         bool      `json:"use_fact_cache,omitempty"`
	Organization         int       `json:"organization,omitempty"`
	UnifiedJobTemplate   int       `json:"unified_job_template,omitempty"`
	LaunchType           string    `json:"launch_type,omitempty"`
	Status               string    `json:"status,omitempty"`
	ExecutionEnvironment any       `json:"execution_environment,omitempty"`
	Failed               bool      `json:"failed,omitempty"`
	Started              any       `json:"started,omitempty"`
	Finished             any       `json:"finished,omitempty"`
	CanceledOn           any       `json:"canceled_on,omitempty"`
	Elapsed              float64   `json:"elapsed,omitempty"`
	JobArgs              string    `json:"job_args,omitempty"`
	JobCwd               string    `json:"job_cwd,omitempty"`
	JobEnv               struct {
	} `json:"job_env,omitempty"`
	JobExplanation          string `json:"job_explanation,omitempty"`
	ExecutionNode           string `json:"execution_node,omitempty"`
	ControllerNode          string `json:"controller_node,omitempty"`
	ResultTraceback         string `json:"result_traceback,omitempty"`
	EventProcessingFinished bool   `json:"event_processing_finished,omitempty"`
	LaunchedBy              struct {
		ID   int    `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Type string `json:"type,omitempty"`
		URL  string `json:"url,omitempty"`
	} `json:"launched_by,omitempty"`
	WorkUnitID             any   `json:"work_unit_id,omitempty"`
	JobTemplate            int   `json:"job_template,omitempty"`
	PasswordsNeededToStart []any `json:"passwords_needed_to_start,omitempty"`
	AllowSimultaneous      bool  `json:"allow_simultaneous,omitempty"`
	Artifacts              struct {
	} `json:"artifacts,omitempty"`
	ScmRevision       string `json:"scm_revision,omitempty"`
	InstanceGroup     any    `json:"instance_group,omitempty"`
	DiffMode          bool   `json:"diff_mode,omitempty"`
	JobSliceNumber    int    `json:"job_slice_number,omitempty"`
	JobSliceCount     int    `json:"job_slice_count,omitempty"`
	WebhookService    string `json:"webhook_service,omitempty"`
	WebhookCredential any    `json:"webhook_credential,omitempty"`
	WebhookGUID       string `json:"webhook_guid,omitempty"`
}

type AnsibleWorkflowJobTemplateResponse struct {
	WorkflowJob   int `json:"workflow_job,omitempty"`
	IgnoredFields struct {
	} `json:"ignored_fields,omitempty"`
	ID      int    `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	URL     string `json:"url,omitempty"`
	Related struct {
		CreatedBy           string `json:"created_by,omitempty"`
		ModifiedBy          string `json:"modified_by,omitempty"`
		UnifiedJobTemplate  string `json:"unified_job_template,omitempty"`
		WorkflowJobTemplate string `json:"workflow_job_template,omitempty"`
		Notifications       string `json:"notifications,omitempty"`
		WorkflowNodes       string `json:"workflow_nodes,omitempty"`
		Labels              string `json:"labels,omitempty"`
		ActivityStream      string `json:"activity_stream,omitempty"`
		Relaunch            string `json:"relaunch,omitempty"`
		Cancel              string `json:"cancel,omitempty"`
	} `json:"related,omitempty"`
	SummaryFields struct {
		Organization struct {
			ID          int    `json:"id,omitempty"`
			Name        string `json:"name,omitempty"`
			Description string `json:"description,omitempty"`
		} `json:"organization,omitempty"`
		Inventory struct {
			ID                           int    `json:"id,omitempty"`
			Name                         string `json:"name,omitempty"`
			Description                  string `json:"description,omitempty"`
			HasActiveFailures            bool   `json:"has_active_failures,omitempty"`
			TotalHosts                   int    `json:"total_hosts,omitempty"`
			HostsWithActiveFailures      int    `json:"hosts_with_active_failures,omitempty"`
			TotalGroups                  int    `json:"total_groups,omitempty"`
			HasInventorySources          bool   `json:"has_inventory_sources,omitempty"`
			TotalInventorySources        int    `json:"total_inventory_sources,omitempty"`
			InventorySourcesWithFailures int    `json:"inventory_sources_with_failures,omitempty"`
			OrganizationID               int    `json:"organization_id,omitempty"`
			Kind                         string `json:"kind,omitempty"`
		} `json:"inventory,omitempty"`
		WorkflowJobTemplate struct {
			ID          int    `json:"id,omitempty"`
			Name        string `json:"name,omitempty"`
			Description string `json:"description,omitempty"`
		} `json:"workflow_job_template,omitempty"`
		UnifiedJobTemplate struct {
			ID             int    `json:"id,omitempty"`
			Name           string `json:"name,omitempty"`
			Description    string `json:"description,omitempty"`
			UnifiedJobType string `json:"unified_job_type,omitempty"`
		} `json:"unified_job_template,omitempty"`
		CreatedBy struct {
			ID        int    `json:"id,omitempty"`
			Username  string `json:"username,omitempty"`
			FirstName string `json:"first_name,omitempty"`
			LastName  string `json:"last_name,omitempty"`
		} `json:"created_by,omitempty"`
		ModifiedBy struct {
			ID        int    `json:"id,omitempty"`
			Username  string `json:"username,omitempty"`
			FirstName string `json:"first_name,omitempty"`
			LastName  string `json:"last_name,omitempty"`
		} `json:"modified_by,omitempty"`
		UserCapabilities struct {
			Delete bool `json:"delete,omitempty"`
			Start  bool `json:"start,omitempty"`
		} `json:"user_capabilities,omitempty"`
		Labels struct {
			Count   int   `json:"count,omitempty"`
			Results []any `json:"results,omitempty"`
		} `json:"labels,omitempty"`
	} `json:"summary_fields,omitempty"`
	Created            time.Time `json:"created,omitempty"`
	Modified           time.Time `json:"modified,omitempty"`
	Name               string    `json:"name,omitempty"`
	Description        string    `json:"description,omitempty"`
	UnifiedJobTemplate int       `json:"unified_job_template,omitempty"`
	LaunchType         string    `json:"launch_type,omitempty"`
	Status             string    `json:"status,omitempty"`
	Failed             bool      `json:"failed,omitempty"`
	Started            any       `json:"started,omitempty"`
	Finished           any       `json:"finished,omitempty"`
	CanceledOn         any       `json:"canceled_on,omitempty"`
	Elapsed            float64   `json:"elapsed,omitempty"`
	JobArgs            string    `json:"job_args,omitempty"`
	JobCwd             string    `json:"job_cwd,omitempty"`
	JobEnv             struct {
	} `json:"job_env,omitempty"`
	JobExplanation  string `json:"job_explanation,omitempty"`
	ResultTraceback string `json:"result_traceback,omitempty"`
	LaunchedBy      struct {
		ID   int    `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Type string `json:"type,omitempty"`
		URL  string `json:"url,omitempty"`
	} `json:"launched_by,omitempty"`
	WorkUnitID          any    `json:"work_unit_id,omitempty"`
	WorkflowJobTemplate int    `json:"workflow_job_template,omitempty"`
	ExtraVars           string `json:"extra_vars,omitempty"`
	AllowSimultaneous   bool   `json:"allow_simultaneous,omitempty"`
	JobTemplate         any    `json:"job_template,omitempty"`
	IsSlicedJob         bool   `json:"is_sliced_job,omitempty"`
	Inventory           int    `json:"inventory,omitempty"`
	Limit               any    `json:"limit,omitempty"`
	ScmBranch           string `json:"scm_branch,omitempty"`
	WebhookService      string `json:"webhook_service,omitempty"`
	WebhookCredential   any    `json:"webhook_credential,omitempty"`
	WebhookGUID         string `json:"webhook_guid,omitempty"`
	SkipTags            any    `json:"skip_tags,omitempty"`
	JobTags             any    `json:"job_tags,omitempty"`
}

type AnsibleBasicResponse struct {
	All []string `json:"__all__,omitempty"`
}

type AnsibleInventoryRequest struct {
	HostFilter   string `json:"host_filter"`
	Kind         string `json:"kind"`
	Name         string `json:"name"`
	Organization int    `json:"organization"`
}

type AnsibleInventoryResponse struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	URL     string `json:"url"`
	Related struct {
		NamedURL               string `json:"named_url,omitempty"`
		CreatedBy              string `json:"created_by,omitempty"`
		ModifiedBy             string `json:"modified_by,omitempty"`
		Hosts                  string `json:"hosts,omitempty"`
		Groups                 string `json:"groups,omitempty"`
		RootGroups             string `json:"root_groups,omitempty"`
		VariableData           string `json:"variable_data,omitempty"`
		Script                 string `json:"script,omitempty"`
		Tree                   string `json:"tree,omitempty"`
		InventorySources       string `json:"inventory_sources,omitempty"`
		UpdateInventorySources string `json:"update_inventory_sources,omitempty"`
		ActivityStream         string `json:"activity_stream,omitempty"`
		JobTemplates           string `json:"job_templates,omitempty"`
		AdHocCommands          string `json:"ad_hoc_commands,omitempty"`
		AccessList             string `json:"access_list,omitempty"`
		ObjectRoles            string `json:"object_roles,omitempty"`
		InstanceGroups         string `json:"instance_groups,omitempty"`
		Copy                   string `json:"copy,omitempty"`
		Labels                 string `json:"labels,omitempty"`
		Organization           string `json:"organization,omitempty"`
	} `json:"related,omitempty"`
	SummaryFields struct {
		Organization struct {
			ID          int    `json:"id,omitempty"`
			Name        string `json:"name,omitempty"`
			Description string `json:"description,omitempty"`
		} `json:"organization,omitempty"`
		CreatedBy struct {
			ID        int    `json:"id,omitempty"`
			Username  string `json:"username,omitempty"`
			FirstName string `json:"first_name,omitempty"`
			LastName  string `json:"last_name,omitempty"`
		} `json:"created_by,omitempty"`
		ModifiedBy struct {
			ID        int    `json:"id,omitempty"`
			Username  string `json:"username,omitempty"`
			FirstName string `json:"first_name,omitempty"`
			LastName  string `json:"last_name,omitempty"`
		} `json:"modified_by,omitempty"`
		ObjectRoles struct {
			AdminRole struct {
				Description string `json:"description,omitempty"`
				Name        string `json:"name,omitempty"`
				ID          int    `json:"id,omitempty"`
			} `json:"admin_role,omitempty"`
			UpdateRole struct {
				Description string `json:"description,omitempty"`
				Name        string `json:"name,omitempty"`
				ID          int    `json:"id,omitempty"`
			} `json:"update_role,omitempty"`
			AdhocRole struct {
				Description string `json:"description,omitempty"`
				Name        string `json:"name,omitempty"`
				ID          int    `json:"id,omitempty"`
			} `json:"adhoc_role,omitempty"`
			UseRole struct {
				Description string `json:"description,omitempty"`
				Name        string `json:"name,omitempty"`
				ID          int    `json:"id,omitempty"`
			} `json:"use_role,omitempty"`
			ReadRole struct {
				Description string `json:"description,omitempty"`
				Name        string `json:"name,omitempty"`
				ID          int    `json:"id,omitempty"`
			} `json:"read_role,omitempty"`
		} `json:"object_roles,omitempty"`
		UserCapabilities struct {
			Edit   bool `json:"edit,omitempty"`
			Delete bool `json:"delete,omitempty"`
			Copy   bool `json:"copy,omitempty"`
			Adhoc  bool `json:"adhoc,omitempty"`
		} `json:"user_capabilities,omitempty"`
		Labels struct {
			Count   int   `json:"count,omitempty"`
			Results []any `json:"results,omitempty"`
		} `json:"labels,omitempty"`
	} `json:"summary_fields,omitempty"`
	Created                      time.Time `json:"created,omitempty"`
	Modified                     time.Time `json:"modified,omitempty"`
	Name                         string    `json:"name,omitempty"`
	Description                  string    `json:"description,omitempty"`
	Organization                 int       `json:"organization,omitempty"`
	Kind                         string    `json:"kind,omitempty"`
	HostFilter                   any       `json:"host_filter,omitempty"`
	Variables                    string    `json:"variables,omitempty"`
	HasActiveFailures            bool      `json:"has_active_failures,omitempty"`
	TotalHosts                   int       `json:"total_hosts,omitempty"`
	HostsWithActiveFailures      int       `json:"hosts_with_active_failures,omitempty"`
	TotalGroups                  int       `json:"total_groups,omitempty"`
	HasInventorySources          bool      `json:"has_inventory_sources,omitempty"`
	TotalInventorySources        int       `json:"total_inventory_sources,omitempty"`
	InventorySourcesWithFailures int       `json:"inventory_sources_with_failures,omitempty"`
	PendingDeletion              bool      `json:"pending_deletion,omitempty"`
	PreventInstanceGroupFallback bool      `json:"prevent_instance_group_fallback,omitempty"`
}

type AnsibleAuthResponse struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	URL     string `json:"url"`
	Related struct {
		User           string `json:"user"`
		ActivityStream string `json:"activity_stream"`
	} `json:"related"`
	SummaryFields struct {
		User struct {
			ID        int    `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		} `json:"user"`
	} `json:"summary_fields"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
	Description  string    `json:"description"`
	User         int       `json:"user"`
	Token        string    `json:"token"`
	RefreshToken any       `json:"refresh_token"`
	Application  any       `json:"application"`
	Expires      time.Time `json:"expires"`
	Scope        string    `json:"scope"`
}

type APIErrors struct {
	Errors []APIError `json:"errors"`
}

type APIError struct {
	Status string `json:"status"`
	Source struct {
		Pointer string `json:"pointer"`
	} `json:"source,omitempty"`
	Title  string `json:"title"`
	Detail string `json:"detail,omitempty"`
}

// handler for Run Task payload
func parseRunTaskPayload(c *gin.Context) RunTaskRequest {
	var request RunTaskRequest

	err := c.Bind(&request)
	if err != nil {
		log.Print(err.Error())
	}

	return request
}

func createRunTaskResponse(status string, message string, detailsUrl string) *RunTaskResponse {

	var response RunTaskResponse

	response.Data.Type = TaskResults
	response.Data.Attributes.Status = status
	if len(message) > 0 {
		response.Data.Attributes.Message = message
	}
	if len(detailsUrl) > 0 {
		response.Data.Attributes.URL = detailsUrl
	}

	// json, err := json.MarshalIndent(response, "", " ")
	// if err != nil {
	// 	log.Print(err)
	// }
	// log.Println((string(json)))

	return &response

}

func tfcRunTaskResponse(runTaskResponse *RunTaskResponse, uri string, token string) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	jsonResponse, jsonErr := json.Marshal(runTaskResponse)

	if jsonErr != nil {
		log.Print(jsonErr)
	}

	req, reqErr := http.NewRequest("PATCH", uri, bytes.NewBuffer(jsonResponse))

	if reqErr != nil {
		log.Print(reqErr)
	}

	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	response, respErr := client.Do(req)

	if respErr != nil {
		log.Print(respErr.Error())
	}
	defer response.Body.Close()
}

func ansibleTokenRequest() (*AnsibleAuthResponse, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, reqErr := http.NewRequest("POST", fmt.Sprintf("%s/%s", ansibleHost, "/api/v2/tokens/"), nil)

	if reqErr != nil {
		return nil, fmt.Errorf(reqErr.Error())
	}

	req.SetBasicAuth(ansibleUser, ansiblePassword)

	response, respErr := client.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf(respErr.Error())
	}

	body, bodyErr := ioutil.ReadAll(response.Body)
	if bodyErr != nil {
		return nil, fmt.Errorf(bodyErr.Error())
	}
	defer response.Body.Close()

	var authResponse AnsibleAuthResponse
	bindErr := json.Unmarshal(body, &authResponse)
	if bindErr != nil {
		return nil, fmt.Errorf(bindErr.Error())
	}

	log.Printf("Sucessfully requested Ansible Token %d", authResponse.ID)

	return &authResponse, nil
}

func ansibleTokenRevoke(authResponse *AnsibleAuthResponse) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, reqErr := http.NewRequest("DELETE", fmt.Sprintf("%s/%s/%d/", ansibleHost, "/api/v2/tokens/", authResponse.ID), nil)

	if reqErr != nil {
		return fmt.Errorf(reqErr.Error())
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authResponse.Token))
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	defer response.Body.Close()

	log.Printf("Sucessfully revoked Ansible Token %d", authResponse.ID)

	return nil
}

func ansibleCreateInventoryRequest(request RunTaskRequest, organisation int, ansibleAuth *AnsibleAuthResponse) (*AnsibleInventoryResponse, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	var inventoryReq AnsibleInventoryRequest
	inventoryReq.Kind = ""
	inventoryReq.Name = request.WorkspaceName
	inventoryReq.Organization = organisation
	inventoryReq.HostFilter = ""

	jsonResponse, jsonErr := json.Marshal(inventoryReq)

	if jsonErr != nil {
		return nil, fmt.Errorf(jsonErr.Error())
	}

	req, reqErr := http.NewRequest("POST", fmt.Sprintf("%s/%s", ansibleHost, "/api/v2/inventories/"), bytes.NewBuffer(jsonResponse))

	if reqErr != nil {
		return nil, fmt.Errorf(reqErr.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ansibleAuth.Token))

	response, respErr := client.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf(respErr.Error())
	}

	body, bodyErr := ioutil.ReadAll(response.Body)
	if reqErr != nil {
		return nil, fmt.Errorf(bodyErr.Error())
	}
	defer response.Body.Close()

	// Not sure which one of these we will get back
	var invResponse AnsibleInventoryResponse
	var basicResponse AnsibleBasicResponse

	bindErr := json.Unmarshal(body, &invResponse)
	bindErrBasic := json.Unmarshal(body, &basicResponse)
	if bindErr != nil {
		return nil, fmt.Errorf(bindErr.Error())
	}
	if bindErrBasic != nil {
		return nil, fmt.Errorf(bindErrBasic.Error())
	}

	if invResponse.ID == 0 && basicResponse.All != nil {
		var sb strings.Builder
		for _, str := range basicResponse.All {
			sb.WriteString(str)
		}
		return nil, fmt.Errorf(sb.String())
	}

	return &invResponse, nil
}

func ansibleJobTemplateRequest(request RunTaskRequest, jobTemplateId string, ansibleAuth *AnsibleAuthResponse) (*AnsibleJobTemplateResponse, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	var jtReq AnsibleJobTemplateRequest
	jsonResponse, jsonErr := json.Marshal(jtReq)

	if jsonErr != nil {
		return nil, fmt.Errorf(jsonErr.Error())
	}

	req, reqErr := http.NewRequest("POST", fmt.Sprintf("%s/%s/%s/%s/", ansibleHost, "/api/v2/job_templates/", jobTemplateId, "launch"), bytes.NewBuffer(jsonResponse))

	if reqErr != nil {
		return nil, fmt.Errorf(reqErr.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ansibleAuth.Token))

	response, respErr := client.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf(respErr.Error())
	}

	body, bodyErr := ioutil.ReadAll(response.Body)
	if bodyErr != nil {
		return nil, fmt.Errorf(bodyErr.Error())
	}

	var jtResponse AnsibleJobTemplateResponse
	bindErr := json.Unmarshal(body, &jtResponse)
	if bindErr != nil {
		return nil, fmt.Errorf(bindErr.Error())
	}

	defer response.Body.Close()

	return &jtResponse, nil
}

func ansibleWorkflowJobTemplateRequest(request RunTaskRequest, workflowTemplateId string, ansibleAuth *AnsibleAuthResponse) (*AnsibleWorkflowJobTemplateResponse, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	var wftjtReq AnsibleWorkflowJobTemplateRequest
	jsonResponse, jsonErr := json.Marshal(wftjtReq)

	if jsonErr != nil {
		return nil, fmt.Errorf(jsonErr.Error())
	}

	req, reqErr := http.NewRequest("POST", fmt.Sprintf("%s/%s/%s/%s/", ansibleHost, "/api/v2/workflow_job_templates/", workflowTemplateId, "launch"), bytes.NewBuffer(jsonResponse))

	if reqErr != nil {
		return nil, fmt.Errorf(reqErr.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ansibleAuth.Token))

	response, respErr := client.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf(respErr.Error())
	}

	body, bodyErr := ioutil.ReadAll(response.Body)
	if bodyErr != nil {
		return nil, fmt.Errorf(bodyErr.Error())
	}

	var wfjtResponse AnsibleWorkflowJobTemplateResponse
	bindErr := json.Unmarshal(body, &wfjtResponse)
	if bindErr != nil {
		return nil, fmt.Errorf(bindErr.Error())
	}

	defer response.Body.Close()

	return &wfjtResponse, nil
}

func handleJobTemplateRunTask(c *gin.Context) {
	var runTask = parseRunTaskPayload(c)
	jobTemplateId := c.Param("jobTemplateId")

	log.Printf("Run Task event received for Job Template ID %s", jobTemplateId)

	c.Status(http.StatusOK)
	// if this isn't a test, send the ackowledgement that we've had the request
	if runTask.AccessToken != TestToken {
		var ansibleAuthResponse, tokErr = ansibleTokenRequest()
		if tokErr != nil {
			errResponse := createRunTaskResponse(Failed, tokErr.Error(), "")
			tfcRunTaskResponse(errResponse, runTask.TaskResultCallbackURL, runTask.AccessToken)
		}

		var jobTemplateResponse, jtErr = ansibleJobTemplateRequest(runTask, jobTemplateId, ansibleAuthResponse)
		if jtErr != nil {
			errResponse := createRunTaskResponse(Failed, jtErr.Error(), "")
			tfcRunTaskResponse(errResponse, runTask.TaskResultCallbackURL, runTask.AccessToken)
		} else {
			response := createRunTaskResponse(Passed, fmt.Sprintf("Succesfully triggered Ansible Job Template, %s", jobTemplateResponse.Name), fmt.Sprintf("%s/#/jobs/playbook/%d/output", ansibleHost, jobTemplateResponse.ID))
			tfcRunTaskResponse(response, runTask.TaskResultCallbackURL, runTask.AccessToken)
		}
		ansibleTokenRevoke(ansibleAuthResponse)
	}

}

func handleWorkflowJobTemplateRunTask(c *gin.Context) {
	var runTask = parseRunTaskPayload(c)
	workflowTemplateId := c.Param("workflowTemplateId")

	log.Printf("Run Task event received for Workflow Template ID %s", workflowTemplateId)

	c.Status(http.StatusOK)
	// if this isn't a test, send the ackowledgement that we've had the request
	if runTask.AccessToken != TestToken {
		var ansibleAuthResponse, tokErr = ansibleTokenRequest()
		if tokErr != nil {
			errResponse := createRunTaskResponse(Failed, tokErr.Error(), "")
			tfcRunTaskResponse(errResponse, runTask.TaskResultCallbackURL, runTask.AccessToken)
		}

		var workflowJobTemplateResponse, wfjtErr = ansibleWorkflowJobTemplateRequest(runTask, workflowTemplateId, ansibleAuthResponse)
		if wfjtErr != nil {
			errResponse := createRunTaskResponse(Failed, wfjtErr.Error(), "")
			tfcRunTaskResponse(errResponse, runTask.TaskResultCallbackURL, runTask.AccessToken)
		} else {
			response := createRunTaskResponse(Passed, fmt.Sprintf("Succesfully triggered Ansible Workflow Job Template, %s", workflowJobTemplateResponse.Name), fmt.Sprintf("%s/#/jobs/workflow/%d/output", ansibleHost, workflowJobTemplateResponse.ID))
			tfcRunTaskResponse(response, runTask.TaskResultCallbackURL, runTask.AccessToken)
		}
		ansibleTokenRevoke(ansibleAuthResponse)
	}

}

func handleInventoryRunTask(c *gin.Context) {
	var runTask = parseRunTaskPayload(c)
	orgIdStr := c.Param("organisationId")
	organisationId, err := strconv.Atoi(orgIdStr)
	if err != nil {
		log.Print(err.Error())
	}

	log.Printf("Inventory Run Task event received for Organisation ID %s", orgIdStr)

	c.Status(http.StatusOK)
	// if this isn't a test, send the ackowledgement that we've had the request
	if runTask.AccessToken != TestToken {
		var ansibleAuthResponse, tokErr = ansibleTokenRequest()
		if tokErr != nil {
			errResponse := createRunTaskResponse(Failed, tokErr.Error(), "")
			tfcRunTaskResponse(errResponse, runTask.TaskResultCallbackURL, runTask.AccessToken)
		}

		var ansibleInvResponse, invErr = ansibleCreateInventoryRequest(runTask, organisationId, ansibleAuthResponse)
		if invErr != nil {
			errResponse := createRunTaskResponse(Failed, invErr.Error(), "")
			tfcRunTaskResponse(errResponse, runTask.TaskResultCallbackURL, runTask.AccessToken)
		} else {
			response := createRunTaskResponse(Passed, fmt.Sprintf("Successfully created Ansible Inventory %s", ansibleInvResponse.Name), fmt.Sprintf("%s/#/inventories/inventory/%d/details", ansibleHost, ansibleInvResponse.ID))
			tfcRunTaskResponse(response, runTask.TaskResultCallbackURL, runTask.AccessToken)
		}
		ansibleTokenRevoke(ansibleAuthResponse)
	}

}

func init() {
	ansibleHost = os.Getenv("ARTS_ANSIBLE_HOST")
	ansibleUser = os.Getenv("ARTS_ANSIBLE_USER")
	ansiblePassword = os.Getenv("ARTS_ANSIBLE_PASSWORD")
}

func main() {
	iface := flag.String("interface", "0.0.0.0", "the default interface on which to listen for requests")
	port := flag.String("port", "9090", "the default port on which to listen for requests")
	log.Println(os.Hostname())
	flag.Parse()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/public/job/:jobTemplateId", handleJobTemplateRunTask)
	router.POST("/public/workflow/:workflowTemplateId", handleWorkflowJobTemplateRunTask)
	router.POST("/public/inventory/:organisationId", handleInventoryRunTask)
	router.Run(fmt.Sprintf("%s:%s", *iface, *port))
}
