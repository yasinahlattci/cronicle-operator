package cronicle_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	CreateEventEndpoint   = "/api/app/create_event/v1"
	UpdateEventEndpoint   = "/api/app/update_event/v1"
	DeleteEventEndpoint   = "/api/app/delete_event/v1"
	getActiveJobsEndpoint = "/api/app/get_active_jobs/v1"
)

type CreateEventResponse struct {
	ID          string `json:"id"`
	Code        int    `json:"code"`
	Description string `json:"description,omitempty"`
}

type StandardResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description,omitempty"`
}

type Job struct {
	Event string `json:"event"`
}

type JobsData struct {
	Jobs map[string]Job `json:"jobs"`
}

// +k8s:deepcopy-gen=true
type CronicleTiming struct {
	Hours    []int `json:"hours,omitempty"`
	Days     []int `json:"days,omitempty"`
	Months   []int `json:"months,omitempty"`
	Weekdays []int `json:"weekdays,omitempty"`
	Years    []int `json:"years,omitempty"`
	Minutes  []int `json:"minutes,omitempty"`
}

// +k8s:deepcopy-gen=true
type CronicleParams struct {
	Script   string `json:"script,omitempty"`
	Annotate int    `json:"annotate,omitempty"`
	Json     int    `json:"json,omitempty"`
}

type CreateEventRequest struct {
	CatchUp       int            `json:"catch_up"`
	Category      string         `json:"category"`
	CpuLimit      int            `json:"cpu_limit"`
	CpuSustain    int            `json:"cpu_sustain"`
	Detached      int            `json:"detached"`
	Enabled       int            `json:"enabled"`
	LogMaxSize    int            `json:"log_max_size"`
	MaxChildren   int            `json:"max_children"`
	MemoryLimit   int            `json:"memory_limit"`
	MemorySustain int            `json:"memory_sustain"`
	Multiplex     int            `json:"multiplex"`
	Notes         string         `json:"notes"`
	NotifyFail    string         `json:"notify_fail"`
	NotifySuccess string         `json:"notify_success"`
	Params        CronicleParams `json:"params"`
	Plugin        string         `json:"plugin"`
	Retries       int            `json:"retries"`
	RetryDelay    int            `json:"retry_delay"`
	Target        string         `json:"target"`
	Timeout       int            `json:"timeout"`
	Timezone      string         `json:"timezone"`
	Timing        CronicleTiming `json:"timing"`
	Title         string         `json:"title"`
	WebHook       string         `json:"web_hook"`
	Algorithm     string         `json:"algorithm"`
}

type UpdateEventRequest struct {
	Id            string         `json:"id"`
	CatchUp       int            `json:"catch_up"`
	Category      string         `json:"category"`
	CpuLimit      int            `json:"cpu_limit"`
	CpuSustain    int            `json:"cpu_sustain"`
	Detached      int            `json:"detached"`
	Enabled       int            `json:"enabled"`
	LogMaxSize    int            `json:"log_max_size"`
	MaxChildren   int            `json:"max_children"`
	MemoryLimit   int            `json:"memory_limit"`
	MemorySustain int            `json:"memory_sustain"`
	Multiplex     int            `json:"multiplex"`
	Notes         string         `json:"notes"`
	NotifyFail    string         `json:"notify_fail"`
	NotifySuccess string         `json:"notify_success"`
	Params        CronicleParams `json:"params"`
	Plugin        string         `json:"plugin"`
	Retries       int            `json:"retries"`
	RetryDelay    int            `json:"retry_delay"`
	Target        string         `json:"target"`
	Timeout       int            `json:"timeout"`
	Timezone      string         `json:"timezone"`
	Timing        CronicleTiming `json:"timing"`
	Title         string         `json:"title"`
	WebHook       string         `json:"web_hook"`
	Algorithm     string         `json:"algorithm"`
}

// CreateEvent is a method that sends a request to the CreateEventEndpoint
func (c *Client) CreateEvent(request CreateEventRequest) (string, error) {
	url := fmt.Sprintf("%s%s", c.config.BaseUrl, CreateEventEndpoint)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	var response CreateEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}
	if response.Code != 0 {
		return "", fmt.Errorf("Error when creating event: %s", response.Description)
	}
	return response.ID, nil
}

func (c *Client) CheckRunningJobs(eventID string) (bool, error) {
	url := fmt.Sprintf("%s%s", c.config.BaseUrl, getActiveJobsEndpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}

	var response JobsData
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, err
	}

	for _, job := range response.Jobs {
		if job.Event == eventID {
			return true, nil
		}
	}
	return false, nil
}

func (c *Client) DeleteEvent(eventID string) error {
	url := fmt.Sprintf("%s%s", c.config.BaseUrl, DeleteEventEndpoint)

	jsonData, err := json.Marshal(map[string]string{"id": eventID})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response code when deleting event: %d", resp.StatusCode)
	}
	var response StandardResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	if response.Code != 0 {
		return fmt.Errorf("Error when deleting event: %s", response.Description)
	}
	return nil
}

func (c *Client) DisableEvent(eventID string) error {
	url := fmt.Sprintf("%s%s", c.config.BaseUrl, UpdateEventEndpoint)

	jsonData, err := json.Marshal(map[string]interface{}{"id": eventID, "enabled": 0})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	var response StandardResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	if response.Code != 0 {
		return fmt.Errorf("Error when disabling event: %s", response.Description)
	}
	return nil
}

func (c *Client) UpdateEvent(request UpdateEventRequest) error {
	url := fmt.Sprintf("%s%s", c.config.BaseUrl, UpdateEventEndpoint)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	var response StandardResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	if response.Code != 0 {
		return fmt.Errorf("Error when updating event: %s", response.Description)
	}
	return nil
}
