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
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var badSecret string = "some_awful_secret_thing_I_need_to_source_from_elsewhere"
var tfcTaskSignatureHeader string = "X-TFC-Task-Signature"
var ansibleHost string
var ansibleUser string
var ansiblePassword string

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

func checkBasicAuth(workspaceID string, hmac string) bool {
	generated := generateHMAC(workspaceID, badSecret)
	return generated == hmac
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

	runTaskInitialResponse(&request, c)
}

func runTaskInitialResponse(req *RunTaskRequest, c *gin.Context) {
	authHeader := c.GetHeader(tfcTaskSignatureHeader)

	if len(authHeader) > 0 {
		if checkBasicAuth(req.WorkspaceID, authHeader) {
			c.Status(http.StatusOK)
		} else {
			unauthorisedResponse(c)
		}

	} else {
		unauthorisedResponse(c)

	}
}

func unauthorisedResponse(c *gin.Context) {
	var apiError APIError
	apiError.Status = strconv.Itoa(http.StatusUnauthorized)
	apiError.Title = "Unauthorised Request"

	var apiErrors APIErrors
	apiErrors.Errors = append(apiErrors.Errors, apiError)
	c.JSON(http.StatusUnauthorized, apiErrors)
}

func hmacRequest(c *gin.Context) {
	var request HMACRequest

	err := c.Bind(&request)
	if err != nil {
		log.Fatalf(err.Error())
	}

	hmacResponse(c, &request)
}

func hmacResponse(c *gin.Context, req *HMACRequest) {
	var hmacResponse HMACResponse

	hmacResponse.WorkspaceID = req.WorkspaceID
	hmacResponse.HMAC = generateHMAC(req.WorkspaceID, badSecret)

	c.JSON(http.StatusOK, hmacResponse)
}

func generateHMAC(data string, secret string) string {
	hmac := hmac.New(sha256.New, []byte(secret))
	hmac.Write([]byte(data))
	sha := hex.EncodeToString(hmac.Sum(nil))
	log.Println("====================================")
	log.Println("Workspace ID:", data)
	log.Println("HMAC:", sha)
	log.Println("====================================")
	return sha
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
	router.POST("/public/job/:jobTemplateId", runTaskRequest)
	router.POST("/public/workflow/:workflowTemplateId", runTaskRequest)
	router.POST("/public/inventory", runTaskRequest)
	router.POST("/private/hmac", hmacRequest)
	router.Run(fmt.Sprintf("%s:%s", *iface, *port))

}
