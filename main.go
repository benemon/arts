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
			Message string `json:"message"`
			URL     string `json:"url"`
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

type AnsibleInventoryRequest struct {
	HostFilter   string `json:"host_filter"`
	Kind         string `json:"kind"`
	Name         string `json:"name"`
	Organization int    `json:"organization"`
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

	jsonOutput, err := json.MarshalIndent(request, "", " ")
	if err != nil {
		log.Print(err.Error())
	}
	log.Printf("%s", string(jsonOutput))

	return request
}

func createRunTaskResponse(status string, message string) RunTaskResponse {

	var response RunTaskResponse

	response.Data.Type = TaskResults
	response.Data.Attributes.Status = status
	response.Data.Attributes.Message = message
	response.Data.Attributes.URL = "https://arts-arts.apps.hoth.onmi.cloud"

	return response

}

func sendTFCResponse(runTaskResponse *RunTaskResponse, uri string, token string) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	jsonResponse, jsonErr := json.Marshal(runTaskResponse)

	if jsonErr != nil {
		log.Print(jsonErr)
	}

	log.Println("OBJECT")
	log.Println(string(jsonResponse))

	req, reqErr := http.NewRequest("PATCH", uri, bytes.NewBuffer(jsonResponse))

	if reqErr != nil {
		log.Print(reqErr)
	}

	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	log.Println("REQUEST")
	log.Println(req)

	response, err := client.Do(req)

	if err != nil {
		log.Print(err.Error())
	}

	log.Println("RESPONSE")
	log.Print(response)

	defer response.Body.Close()

}

func ansibleAuthRequest() AnsibleAuthResponse {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, reqErr := http.NewRequest("POST", fmt.Sprintf("%s/%s", ansibleHost, "/api/v2/tokens/"), nil)

	if reqErr != nil {
		log.Print(reqErr)
	}

	req.SetBasicAuth(ansibleUser, ansiblePassword)

	response, err := client.Do(req)

	if err != nil {
		log.Print(err.Error())
	}

	body, nil := ioutil.ReadAll(response.Body)
	var authResponse AnsibleAuthResponse
	bindErr := json.Unmarshal(body, &authResponse)
	if bindErr != nil {
		log.Print(bindErr.Error())
	}

	defer response.Body.Close()

	return authResponse
}

func ansibleCreateInventoryRequest(request RunTaskRequest, organisation int, token string) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	var inventoryReq AnsibleInventoryRequest
	inventoryReq.Kind = ""
	inventoryReq.Name = request.WorkspaceID
	inventoryReq.Organization = organisation
	inventoryReq.HostFilter = ""

	jsonResponse, jsonErr := json.Marshal(inventoryReq)

	if jsonErr != nil {
		log.Print(jsonErr)
	}

	req, reqErr := http.NewRequest("POST", fmt.Sprintf("%s/%s", ansibleHost, "/api/v2/inventories/"), bytes.NewBuffer(jsonResponse))

	if reqErr != nil {
		log.Print(reqErr)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	response, err := client.Do(req)

	if err != nil {
		log.Print(err.Error())
	}

	defer response.Body.Close()

}

func handleJobTemplateRunTask(c *gin.Context) {
	var runTask = parseRunTaskPayload(c)
	jobTemplateId := c.Param("jobTemplateId")

	log.Printf("Run Task event received for Job Template ID %s", jobTemplateId)

	c.Status(http.StatusOK)
	// if this isn't a test, send the ackowledgement that we've had the request
	if runTask.AccessToken != TestToken {
		// we'll just send an immediate response because we're not doing anything yet
		response := createRunTaskResponse(Passed, "Request Completed Succesfully")
		sendTFCResponse(&response, runTask.TaskResultCallbackURL, runTask.AccessToken)
	}

}

func handleWorkflowTemplateRunTask(c *gin.Context) {
	var runTask = parseRunTaskPayload(c)
	workflowTemplateId := c.Param("workflowTemplateId")

	log.Printf("Run Task event received for Workflow Template ID %s", workflowTemplateId)

	c.Status(http.StatusOK)
	// if this isn't a test, send the ackowledgement that we've had the request
	if runTask.AccessToken != TestToken {
		// we'll just send an immediate response because we're not doing anything yet
		response := createRunTaskResponse(Passed, "Request Completed Succesfully")
		sendTFCResponse(&response, runTask.TaskResultCallbackURL, runTask.AccessToken)
	}

}

func handleInventoryRunTask(c *gin.Context) {
	var runTask = parseRunTaskPayload(c)
	var ansibleAuthRequest = ansibleAuthRequest()
	orgIdStr := c.Param("organisationId")
	organisationId, err := strconv.Atoi(orgIdStr)

	c.Status(http.StatusOK)

	log.Printf("ACCESS TOKEN: %s", runTask.AccessToken)
	// if this isn't a test, send the ackowledgement that we've had the request
	if runTask.AccessToken != TestToken {
		log.Printf("Inventory Run Task event received for Organisation Template ID %s", orgIdStr)
		if err != nil {
			log.Print(err.Error())
		}
		ansibleCreateInventoryRequest(runTask, organisationId, ansibleAuthRequest.Token)
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

	router := gin.Default()
	router.POST("/public/job/:jobTemplateId", handleJobTemplateRunTask)
	router.POST("/public/workflow/:workflowTemplateId", handleWorkflowTemplateRunTask)
	router.POST("/public/inventory/:organisationId", handleInventoryRunTask)
	router.Run(fmt.Sprintf("%s:%s", *iface, *port))
}
