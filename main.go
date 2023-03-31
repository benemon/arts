package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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

func runTaskRequest(c *gin.Context) {
	var request RunTaskRequest
	jobTemplateId := c.Param("jobTemplateId")

	log.Printf("Run Task event received for Job Template ID %s", jobTemplateId)

	err := c.Bind(&request)
	if err != nil {
		log.Fatalf(err.Error())
	}

	jsonOutput, err := json.MarshalIndent(request, "", " ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf("%s", string(jsonOutput))

	runTaskResponse(c)
}

func runTaskResponse(c *gin.Context) {
	var response RunTaskResponse
	response.Data.Type = "task-results"
	response.Data.Attributes.Status = "passed"
	response.Data.Attributes.Message = "The task completed succesfully"
	response.Data.Attributes.URL = "http://localhost:9090/request"

	c.JSON(http.StatusOK, response)
}

func hmacRequest(c *gin.Context) {
	var request HMACRequest

	err := c.Bind(&request)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println(request.WorkspaceID)
	hmacResponse(c, &request)
}

func hmacResponse(c *gin.Context, req *HMACRequest) {
	var hmacResponse HMACResponse
	log.Println(req.WorkspaceID)

	hmac := hmac.New(sha256.New, []byte("would_use_vault_as_a_kms"))
	hmac.Write([]byte(req.WorkspaceID))
	sha := hex.EncodeToString(hmac.Sum(nil))

	hmacResponse.WorkspaceID = req.WorkspaceID
	hmacResponse.HMAC = sha

	c.JSON(http.StatusOK, hmacResponse)
}

func main() {
	iface := flag.String("interface", "0.0.0.0", "the default interface on which to listen for requests")
	port := flag.String("port", "9090", "the default port on which to listen for requests")
	log.Println(os.Hostname())
	flag.Parse()

	// ansibleHost := os.Getenv("ARTS_ANSIBLE_HOST")
	// ansibleUser := os.Getenv("ARTS_ANSIBLE_USER")
	// ansiblePassword := os.Getenv("ARTS_ANSIBLE_PASSWORD")
	//

	router := gin.Default()
	router.POST("/job/:jobTemplateId", runTaskRequest)
	router.POST("/hmac", hmacRequest)
	router.Run(fmt.Sprintf("%s:%s", *iface, *port))

}
