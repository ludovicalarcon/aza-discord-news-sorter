package todoist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	BASE_URL            = "https://api.todoist.com/rest/v2"
	API_KEY             = "API_KEY"
	MAX_TODO_PER_DAY    = 5
	MAX_DAYS_TO_LOOK_UP = 30
)

var (
	ErrProjectNotFound         = errors.New("todoist project not found")
	ErrNotInitialized          = errors.New("todoist object not initialized, call init method")
	ErrHttpRequestDefault      = errors.New("error on todoist api call")
	ErrHttpRequestUnauthorized = errors.New("unauthorized todoist access")
	ErrAlreadyExist            = errors.New("todo already exist")
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
	if err != nil {
		return
	}

	err = verifyErrorInAnswer(response)
	if err != nil {
		return
	}

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

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(responseData, &projects)

	return
}

func (t *Todoist) getTodosByLabel(label string) (todos []Task, err error) {
	url := fmt.Sprintf("%s/tasks?project_id=%s&label=%s", t.baseUrl, t.projectId, label)
	request, _ := http.NewRequest(http.MethodGet, url, nil)

	response, err := doHttpRequest(request, t.Client, t.apiKey)
	if err != nil {
		return
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(responseData, &todos)
	return
}

func (t *Todoist) defineDueDate() (dueDate string, err error) {
	currentDate := time.Now().Format("2006-01-02")

	for i := 1; i < MAX_DAYS_TO_LOOK_UP; i++ {
		todos, err := t.getTodosByLabel(currentDate)
		if err != nil {
			return "", err
		}
		if len(todos) < MAX_TODO_PER_DAY {
			return currentDate, err
		}
		currentDate = time.Now().AddDate(0, 0, i).Format("2006-01-02")
	}

	dueDate = currentDate
	return
}

func (t *Todoist) createTodoDTO(title, titleLabel, description string) (todo Task, err error) {
	dueDate, err := t.defineDueDate()
	if err != nil {
		return
	}

	labels := []string{titleLabel, dueDate}
	todo = Task{
		ProjectId:   &t.projectId,
		Content:     &title,
		Description: &description,
		Labels:      labels,
		DueDate:     &dueDate,
	}
	return
}

func ensureTodoNotAlreadyExist(title string, todoist *Todoist) (err error) {
	titleLabel := strings.ReplaceAll(strings.Trim(title, " "), " ", "-")
	todos, err := todoist.getTodosByLabel(titleLabel)
	if err != nil {
		return
	}
	if len(todos) > 0 {
		log.Printf("a todo for %s already exist, skip", title)
		return ErrAlreadyExist
	}
	return nil
}

func (t *Todoist) CreateTodo(title, description string) (err error) {
	if t.apiKey == "" {
		return ErrNotInitialized
	}

	err = ensureTodoNotAlreadyExist(title, t)
	if err != nil {
		return
	}

	titleLabel := strings.ReplaceAll(strings.Trim(title, " "), " ", "-")
	todo, err := t.createTodoDTO(title, titleLabel, description)
	if err != nil {
		return
	}

	url := fmt.Sprintf("%s/tasks", t.baseUrl)
	data, err := json.Marshal(todo)
	if err != nil {
		return
	}

	request, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	response, err := doHttpRequest(request, t.Client, t.apiKey)
	if err != nil {
		return
	}
	defer response.Body.Close()

	return
}
