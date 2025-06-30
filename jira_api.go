package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// --- Jira API Configuration ---
const (
	jiraBaseURL = "https://vus-edtech.atlassian.net"
	jiraAPIPath = "/rest/api/2/search"
)

// --- Jira Data Structures ---

type JiraSearchResponse struct {
	Issues []Issue `json:"issues"`
}

type Issue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields Fields `json:"fields"`
}

type Fields struct {
	Summary     string      `json:"summary"`
	Status      Status      `json:"status"`
	IssueType   IssueType   `json:"issuetype"`
	Assignee    *User       `json:"assignee"`
	Reporter    *User       `json:"reporter"`
	Priority    *Priority   `json:"priority"`
	Description interface{} `json:"description"` // Jira description can be string or object
	Created     CustomTime  `json:"created"`
	Updated     CustomTime  `json:"updated"`
	Comments    *Comments   `json:"comment"`
}

type Comments struct {
	Comments   []Comment `json:"comments"`
	MaxResults int       `json:"maxResults"`
	Total      int       `json:"total"`
	StartAt    int       `json:"startAt"`
}

type Comment struct {
	Author  *User      `json:"author"`
	Body    string     `json:"body"`
	Created CustomTime `json:"created"`
	Updated CustomTime `json:"updated"`
}

type Status struct {
	Name string `json:"name"`
}

type IssueType struct {
	Name string `json:"name"`
}

type User struct {
	DisplayName string `json:"displayName"`
}

type Priority struct {
	Name string `json:"name"`
}

type Board struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Sprint struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

type CustomTime struct {
	time.Time
}

const jiraTimeLayout = "2006-01-02T15:04:05.999-0700"

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	s = s[1 : len(s)-1] // Remove quotes
	ct.Time, err = time.Parse(jiraTimeLayout, s)
	return
}

// --- Helper Functions for Jira API ---

// FetchJiraStatuses fetches all available statuses from Jira
func FetchJiraStatuses(apiToken string) ([]Status, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	url := fmt.Sprintf("%s/rest/api/2/status", jiraBaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.SetBasicAuth("huannguyenh@vus-etsc.edu.vn", apiToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to Jira: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API returned non-OK status: %s Response: %s", resp.Status, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var statuses []Status
	if err := json.Unmarshal(bodyBytes, &statuses); err != nil {
		return nil, fmt.Errorf("error unmarshalling statuses: %w", err)
	}

	return statuses, nil
}

// FetchJiraUsers fetches all active users from Jira
func FetchJiraUsers(apiToken string) ([]User, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	// Jira Cloud API for user search requires a query parameter, e.g., 'query=.' for all users
	url := fmt.Sprintf("%s/rest/api/2/user/search?query=.", jiraBaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.SetBasicAuth("huannguyenh@vus-etsc.edu.vn", apiToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to Jira: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API returned non-OK status: %s Response: %s", resp.Status, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var users []User
	if err := json.Unmarshal(bodyBytes, &users); err != nil {
		return nil, fmt.Errorf("error unmarshalling users: %w", err)
	}

	return users, nil
}

// FetchJiraBoards fetches all boards from Jira
func FetchJiraBoards(apiToken string) ([]Board, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	url := fmt.Sprintf("%s/rest/agile/1.0/board", jiraBaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.SetBasicAuth("huannguyenh@vus-etsc.edu.vn", apiToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to Jira: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API returned non-OK status: %s Response: %s", resp.Status, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var boardResponse struct {
		Values []Board `json:"values"`
	}
	if err := json.Unmarshal(bodyBytes, &boardResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling boards: %w", err)
	}

	return boardResponse.Values, nil
}

// FetchJiraSprints fetches all sprints for a given board from Jira
func FetchJiraSprints(apiToken string, boardID int) ([]Sprint, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	url := fmt.Sprintf("%s/rest/agile/1.0/board/%d/sprint", jiraBaseURL, boardID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.SetBasicAuth("huannguyenh@vus-etsc.edu.vn", apiToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to Jira: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API returned non-OK status: %s Response: %s", resp.Status, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var sprintResponse struct {
		Values []Sprint `json:"values"`
	}
	if err := json.Unmarshal(bodyBytes, &sprintResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling sprints: %w", err)
	}

	return sprintResponse.Values, nil
}

/* <<<<<<<<<<<<<<  ✨ Windsurf Command ⭐ >>>>>>>>>>>>>>>> */
// FetchJiraIssues fetches issues from Jira matching the given JQL query, and returns them in a slice of Issue objects.
//
// The function takes in the email address associated with the Jira account, the API token for that account, and the JQL query to use for the search.
//
// The function first constructs the URL for the Jira API call by appending the JQL query to the base Jira URL, and then makes the request using the provided API token.
//
// If the request is successful, the function reads the response body into a byte slice and unmarshals the JSON into a JiraSearchResponse struct.
//
// If unmarshalling is successful, the function returns the slice of Issues from the JiraSearchResponse struct. If either the request or the unmarshalling fails, the function returns an error.
/* <<<<<<<<<<  22c8d570-0811-4420-838e-2264f3f593a4  >>>>>>>>>>> */
func FetchJiraIssues(email, apiToken string, jql string) ([]Issue, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	jiraURL, err := url.Parse(fmt.Sprintf("%s%s", jiraBaseURL, jiraAPIPath))
	if err != nil {
		return nil, fmt.Errorf("error parsing Jira URL: %w", err)
	}

	params := url.Values{}
	params.Add("jql", jql)
	params.Add("maxResults", "100")
	// Requesting specific fields
	params.Add("fields", "summary,status,issuetype,assignee,reporter,priority,description,created,updated,comment")
	params.Add("expand", "renderedFields")
	jiraURL.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", jiraURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.SetBasicAuth(email, apiToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to Jira: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API returned non-OK status: %s Response: %s", resp.Status, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var jiraResponse JiraSearchResponse
	if err := json.Unmarshal(bodyBytes, &jiraResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	return jiraResponse.Issues, nil
}
