package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/JudgmentLabs/judgeval-go/internal/api/models"
)

type Client struct {
	baseURL        string
	apiKey         string
	organizationID string
	httpClient     *http.Client
}

func NewClient(baseURL, apiKey, organizationID string) *Client {
	return &Client{
		baseURL:        baseURL,
		apiKey:         apiKey,
		organizationID: organizationID,
		httpClient:     &http.Client{},
	}
}

func (c *Client) buildURL(path string, queryParams map[string]string) string {
	u, _ := url.Parse(c.baseURL + path)
	if len(queryParams) > 0 {
		q := u.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}
	return u.String()
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Organization-Id", c.organizationID)
}

func (c *Client) GetBaseURL() string {
	return c.baseURL
}

func (c *Client) GetAPIKey() string {
	return c.apiKey
}

func (c *Client) GetOrganizationID() string {
	return c.organizationID
}

func (c *Client) PostOtelV1Traces() (*interface{}, error) {
	path := "/otel/v1/traces"
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(struct{}{})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostOtelTriggerRootSpanRules(payload *models.TriggerRootSpanRulesRequest) (*models.TriggerRootSpanRulesResponse, error) {
	path := "/otel/trigger_root_span_rules"
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.TriggerRootSpanRulesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsResolve(payload *models.ResolveProjectRequest) (*models.ResolveProjectResponse, error) {
	path := "/v1/projects/resolve/"
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.ResolveProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjects(payload *models.AddProjectRequest) (*models.AddProjectResponse, error) {
	path := "/v1/projects"
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.AddProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteProjects(projectId string) (*models.DeleteProjectResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s", projectId)
	url := c.buildURL(path, nil)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.DeleteProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsDatasets(projectId string, payload *models.CreateDatasetRequest) (*models.CreateDatasetResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/datasets", projectId)
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.CreateDatasetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProjectsDatasets(projectId string) (*models.PullAllDatasetsResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/datasets", projectId)
	url := c.buildURL(path, nil)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.PullAllDatasetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsDatasetsByDatasetNameExamples(projectId string, datasetName string, payload *models.InsertExamplesRequest) (*models.InsertExamplesResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/datasets/%s/examples", projectId, datasetName)
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.InsertExamplesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProjectsDatasetsByDatasetName(projectId string, datasetName string) (*models.PullDatasetResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/datasets/%s", projectId, datasetName)
	url := c.buildURL(path, nil)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.PullDatasetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsEvaluateExamples(projectId string, payload *models.ExampleEvaluationRun) (*interface{}, error) {
	path := fmt.Sprintf("/v1/projects/%s/evaluate/examples", projectId)
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsEvaluateTraces(projectId string, payload *models.TraceEvaluationRun) (*interface{}, error) {
	path := fmt.Sprintf("/v1/projects/%s/evaluate/traces", projectId)
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsEvalResults(projectId string, payload *models.LogEvalResultsRequest) (*models.LogEvalResultsResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/eval-results", projectId)
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.LogEvalResultsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProjectsExperimentsByRunId(projectId string, runId string) (*models.FetchExperimentRunResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/experiments/%s", projectId, runId)
	url := c.buildURL(path, nil)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.FetchExperimentRunResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsEvalQueueExamples(projectId string, payload *models.ExampleEvaluationRun) (*models.AddToRunEvalQueueExamplesResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/eval-queue/examples", projectId)
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.AddToRunEvalQueueExamplesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsEvalQueueTraces(projectId string, payload *models.TraceEvaluationRun) (*models.AddToRunEvalQueueTracesResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/eval-queue/traces", projectId)
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.AddToRunEvalQueueTracesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProjectsPromptsByName(projectId string, name string, commit_id *string, tag *string) (*models.FetchPromptResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/prompts/%s", projectId, name)
	queryParams := make(map[string]string)
	if commit_id != nil {
		queryParams["commit_id"] = *commit_id
	}
	if tag != nil {
		queryParams["tag"] = *tag
	}
	url := c.buildURL(path, queryParams)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.FetchPromptResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsPrompts(projectId string, payload *models.InsertPromptRequest) (*models.InsertPromptResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/prompts", projectId)
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.InsertPromptResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsPromptsByNameTags(projectId string, name string, payload *models.TagPromptRequest) (*models.TagPromptResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/prompts/%s/tags", projectId, name)
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.TagPromptResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteProjectsPromptsByNameTags(projectId string, name string, payload *models.UntagPromptRequest) (*models.UntagPromptResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/prompts/%s/tags", projectId, name)
	url := c.buildURL(path, nil)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.UntagPromptResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProjectsPromptsByNameVersions(projectId string, name string) (*models.GetPromptVersionsResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/prompts/%s/versions", projectId, name)
	url := c.buildURL(path, nil)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.GetPromptVersionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProjectsScorers(projectId string, names *string, is_trace *string) (*models.FetchPromptScorersResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/scorers", projectId)
	queryParams := make(map[string]string)
	if names != nil {
		queryParams["names"] = *names
	}
	if is_trace != nil {
		queryParams["is_trace"] = *is_trace
	}
	url := c.buildURL(path, queryParams)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.FetchPromptScorersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsScorers(projectId string, payload *models.SavePromptScorerRequest) (*models.SavePromptScorerResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/scorers", projectId)
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.SavePromptScorerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProjectsScorersByNameExists(projectId string, name string) (*models.ScorerExistsResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/scorers/%s/exists", projectId, name)
	url := c.buildURL(path, nil)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.ScorerExistsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsScorersCustom(projectId string, payload *models.UploadCustomScorerRequest) (*models.UploadCustomScorerResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/scorers/custom", projectId)
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.UploadCustomScorerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProjectsScorersCustomByNameExists(projectId string, name string) (*models.CustomScorerExistsResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/scorers/custom/%s/exists", projectId, name)
	url := c.buildURL(path, nil)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.CustomScorerExistsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostProjectsTracesByTraceIdTags(projectId string, traceId string, payload *models.AddTraceTagsRequest) (*models.AddTraceTagsResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/traces/%s/tags", projectId, traceId)
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.AddTraceTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostE2eFetchTrace(payload *models.E2EFetchTraceRequest) (*models.E2EFetchTraceResponse, error) {
	path := "/v1/e2e_fetch_trace/"
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.E2EFetchTraceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PostE2eFetchSpanScore(payload *models.E2EFetchSpanScoreRequest) (*models.E2EFetchSpanScoreResponse, error) {
	path := "/v1/e2e_fetch_span_score/"
	url := c.buildURL(path, nil)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))
	}

	var result models.E2EFetchSpanScoreResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
