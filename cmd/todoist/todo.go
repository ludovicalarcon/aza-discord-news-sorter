package todoist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	BASE_URL = "https://api.todoist.com/rest/v2"
	API_KEY  = "API_KEY"
)

var (
	ErrProjectNotFound         = errors.New("todoist project not found")
	ErrNotInitialized          = errors.New("todoist object not initialized, call init method")
	ErrHttpRequestDefault      = errors.New("error on todoist api call")
	ErrHttpRequestUnauthorized = errors.New("unauthorized todoist access")
)

type Todoist struct {
	baseUrl     string
	Client      *http.Client
	ProjectName string

	apiKey    string
	projectId string
}

func (t *Todoist) Init(projectName string) (err error) {
	key := fmt.Sprintf("$%s", API_KEY)
	t.apiKey = os.ExpandEnv(key)

	t.ProjectName = projectName
	if t.baseUrl == "" {
		t.baseUrl = BASE_URL
	}
	t.projectId, err = t.getProjectId(projectName)
	return
}

func verifyErrorInAnswer(response *http.Response) (err error) {
	if response.StatusCode == http.StatusOK {
		return nil
	}
	if response.StatusCode == http.StatusUnauthorized {
		return ErrHttpRequestUnauthorized
	}
	responseData, _ := io.ReadAll(response.Body)
	return fmt.Errorf("%w: %s", ErrHttpRequestDefault, string(responseData))
}

func doHttpRequest(request *http.Request, client *http.Client, apiKey string) (response *http.Response, err error) {
	bearer := fmt.Sprintf("Bearer %s", apiKey)
	request.Header.Set("Authorization", bearer)
	request.Header.Set("Content-Type", "application/json")

	response, err = client.Do(request)
	return
}

func (t *Todoist) getProjectId(projectName string) (projectId string, err error) {
	projects, err := t.getProjects()
	if err != nil {
		return
	}

	for _, project := range projects {
		if *project.Name == projectName {
			projectId = *project.Id
			break
		}
	}

	if projectId == "" {
		err = ErrProjectNotFound
	}

	return
}

func (t *Todoist) getProjects() (projects []Project, err error) {
	url := fmt.Sprintf("%s/projects", t.baseUrl)
	request, _ := http.NewRequest(http.MethodGet, url, nil)

	response, err := doHttpRequest(request, t.Client, t.apiKey)
	if err != nil {
		return
	}

	defer response.Body.Close()
	err = verifyErrorInAnswer(response)
	if err != nil {
		return
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(responseData, &projects)

	return
}

func (t *Todoist) CreateTodo(title, description string) (err error) {
	labels := []string{title}
	todo := Task{
		ProjectId:   &t.projectId,
		Content:     &title,
		Description: &description,
		Labels:      labels,
	}

	if t.apiKey == "" {
		return ErrNotInitialized
	}

	url := fmt.Sprintf("%s/tasks", t.baseUrl)
	data, err := json.Marshal(todo)
	if err != nil {
		return err
	}
	request, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))

	response, err := doHttpRequest(request, t.Client, t.apiKey)
	if err != nil {
		return
	}

	defer response.Body.Close()
	err = verifyErrorInAnswer(response)

	return
}
