package todoist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testTitleLabel struct {
	title      string
	titleLabel string
}

func TestTodoist(t *testing.T) {
	id := "12345"
	name := "Tests"
	proj := []Project{{
		Id:   &id,
		Name: &name,
	}}

	assertError := func(t testing.TB, got error, want error) {
		t.Helper()
		if got == nil {
			t.Fatal("didn't get an error but wanted one")
		}

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	assertNoError := func(t testing.TB, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("got an error but didn't want one: %q", err)
		}
	}

	assertEqualString := func(t testing.TB, got, want string) {
		t.Helper()
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}

	assertEmptyTodo := func(t testing.TB, todo Task) {
		t.Helper()
		if todo.ProjectId != nil {
			t.Fatalf("projectId should be null but equal: %v", todo.ProjectId)
		}
		if todo.Content != nil {
			t.Fatalf("content should be null but equal: %v", todo.Content)
		}
		if todo.Description != nil {
			t.Fatalf("description should be null but equal: %v", todo.Description)
		}
		if todo.Labels != nil {
			t.Fatalf("labels should be null but equal: %v", todo.Labels)
		}
	}

	t.Run("It should init all field on init call", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			data, err := json.Marshal(proj)
			if err != nil {
				t.Fatal("can't marshal json for testserver answer")
			}
			rw.Write(data)
		}))

		defer server.Close()

		t.Setenv(API_KEY, "XXX")
		todoist := Todoist{Client: server.Client(), baseUrl: server.URL}

		err := todoist.Init(name)

		assertNoError(t, err)
		assertEqualString(t, todoist.apiKey, "XXX")
		assertEqualString(t, todoist.projectId, id)
		assertEqualString(t, todoist.ProjectName, name)
	})

	t.Run("It should retrieve project id", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			data, err := json.Marshal(proj)
			if err != nil {
				t.Fatal("can't marshal json for testserver answer")
			}
			rw.Write(data)
		}))

		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL}
		projId, err := todoist.getProjectId(name)

		assertNoError(t, err)
		assertEqualString(t, projId, id)
	})

	t.Run("It should return error when project not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			data, err := json.Marshal(proj)
			if err != nil {
				t.Fatal("can't marshal json for testserver answer")
			}
			rw.Write(data)
		}))

		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL}
		_, err := todoist.getProjectId("unknown")

		assertError(t, err, ErrProjectNotFound)
	})

	t.Run("It should handle error when http call failed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		}))
		server.URL = "" // to simulate client.Do error

		defer server.Close()

		request, _ := http.NewRequest(http.MethodGet, server.URL, nil)

		_, err := doHttpRequest(request, server.Client(), "XXX")
		if err == nil {
			t.Fatal("didn't get an error but wanted one")
		}
	})

	t.Run("It should return error when get unauthorized status code on projects route", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}))

		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL}
		_, err := todoist.getProjectId(name)

		assertError(t, err, ErrHttpRequestUnauthorized)
	})

	t.Run("It should return error when status code is neither OK not unauthorized on projects route", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "oups", http.StatusInternalServerError)
		}))

		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL}
		_, err := todoist.getProjectId(name)

		assertEqualString(t, err.Error(), fmt.Sprintf("%s: %s", ErrHttpRequestDefault.Error(), "oups\n"))
	})

	t.Run("It should return error when get unauthorized status code on tasks route", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}))

		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL, apiKey: "XXX"}
		err := todoist.CreateTodo("test", "test")

		assertError(t, err, ErrHttpRequestUnauthorized)
	})

	t.Run("It should return error when status code is neither OK not unauthorized on tasks route", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "oups", http.StatusInternalServerError)
		}))

		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL, apiKey: "XXX"}
		err := todoist.CreateTodo("test", "test")

		assertEqualString(t, err.Error(), fmt.Sprintf("%s: %s", ErrHttpRequestDefault.Error(), "oups\n"))
	})

	t.Run("It should return error when apiKey is not set", func(t *testing.T) {
		title := "foobar"
		todoist := Todoist{}
		err := todoist.CreateTodo(title, title)
		assertError(t, err, ErrNotInitialized)
	})

	t.Run("It should return error when todo already exist", func(t *testing.T) {
		title := "foobar"
		todos := []Task{
			{
				ProjectId:   &id,
				Labels:      []string{title},
				Content:     &title,
				Description: &title,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			data, err := json.Marshal(todos)
			if err != nil {
				t.Fatal("can't marshal json for testserver answer")
			}
			rw.Write(data)
		}))
		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL, apiKey: "XXX"}
		err := todoist.CreateTodo(title, title)
		assertError(t, err, ErrAlreadyExist)
	})

	t.Run("It should retrieve a due date for today when today todos are less than MAX_TODO_PER_DAY", func(t *testing.T) {
		dateFormated := "1970-01-01"
		todos := []Task{
			{}, {},
		}
		date, err := time.Parse("2006-01-02", dateFormated)
		assertNoError(t, err)

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			data, err := json.Marshal(todos)
			if err != nil {
				t.Fatal("can't marshal json for testserver answer")
			}
			rw.Write(data)
		}))
		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL, apiKey: "XXX"}
		dueDate, err := todoist.defineDueDate(date)

		assertNoError(t, err)
		assertEqualString(t, dueDate, dateFormated)
	})

	t.Run("It should retrieve a due date for next day when today todos have already MAX_TODO_PER_DAY", func(t *testing.T) {
		dateFormated := "1970-01-01"
		todos := []Task{{}, {}, {}, {}, {}}
		date, err := time.Parse("2006-01-02", dateFormated)
		assertNoError(t, err)

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			data, err := json.Marshal([]Task{})
			if strings.Contains(req.URL.RawQuery, dateFormated) {
				data, err = json.Marshal(todos)
			}
			if err != nil {
				t.Fatal("can't marshal json for testserver answer")
			}
			rw.Write(data)
		}))
		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL, apiKey: "XXX"}
		dueDate, err := todoist.defineDueDate(date)

		assertNoError(t, err)
		assertEqualString(t, dueDate, date.AddDate(0, 0, 1).Format("2006-01-02"))
	})

	t.Run("It should retrieve a due date for a maximum of MAX_DAYS_TO_LOOK_UP", func(t *testing.T) {
		dateFormated := "1970-01-01"
		todos := []Task{{}, {}, {}, {}, {}}
		date, err := time.Parse("2006-01-02", dateFormated)
		assertNoError(t, err)

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			data, err := json.Marshal(todos)
			if err != nil {
				t.Fatal("can't marshal json for testserver answer")
			}
			rw.Write(data)
		}))
		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL, apiKey: "XXX"}
		dueDate, err := todoist.defineDueDate(date)

		assertNoError(t, err)
		assertEqualString(t, dueDate, date.AddDate(0, 0, MAX_DAYS_TO_LOOK_UP).Format("2006-01-02"))
	})

	t.Run("It should create DTO with all mandatory fields", func(t *testing.T) {
		dueDateFormated := time.Now().Format("2006-01-02")
		title := "foobar"
		description := "foobar todo"

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// On DTO creation, todoist is called to retrieve proper due date, returning empty to have today date
			data, err := json.Marshal([]Task{})
			if err != nil {
				t.Fatal("can't marshal json for testserver answer")
			}
			rw.Write(data)
		}))
		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL, apiKey: "XXX", projectId: id}
		got, err := todoist.createTodoDTO(title, description)

		assertNoError(t, err)
		assertEqualString(t, *got.ProjectId, id)
		assertEqualString(t, *got.Content, title)
		assertEqualString(t, *got.Description, description)
		if len(got.Labels) != 2 {
			t.Fatalf("todo should have 2 labels but got %d", len(got.Labels))
		}
		assertEqualString(t, got.Labels[0], title)
		assertEqualString(t, got.Labels[1], dueDateFormated)
	})

	t.Run("It should replace space by - for title label and trim it", func(t *testing.T) {
		expected := []testTitleLabel{
			{title: "foobar", titleLabel: "foobar"},
			{title: "foo bar", titleLabel: "foo-bar"},
			{title: "   foo bar title   ", titleLabel: "foo-bar-title"},
			{title: "   foobar", titleLabel: "foobar"},
			{title: "foobar     ", titleLabel: "foobar"},
		}
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// On DTO creation, todoist is called to retrieve proper due date, returning empty to have today date
			data, err := json.Marshal([]Task{})
			if err != nil {
				t.Fatal("can't marshal json for testserver answer")
			}
			rw.Write(data)
		}))
		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL, apiKey: "XXX", projectId: id}

		for _, current := range expected {

			got, err := todoist.createTodoDTO(current.title, current.title)

			assertNoError(t, err)
			if len(got.Labels) != 2 {
				t.Fatalf("todo should have 2 labels")
			}
			assertEqualString(t, got.Labels[0], current.titleLabel)
		}
	})

	t.Run("It should return empty DTO on todoist error", func(t *testing.T) {
		title := "foo"
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "oups", http.StatusInternalServerError)
		}))
		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL, apiKey: "XXX", projectId: id}
		got, err := todoist.createTodoDTO(title, title)

		if err == nil {
			t.Fatal("didn't get any error but wanted one")
		}

		assertEqualString(t, err.Error(), fmt.Sprintf("%s: oups\n", ErrHttpRequestDefault))
		assertEmptyTodo(t, got)
	})

	t.Run("It should create todo", func(t *testing.T) {
		var todoInReq *Task
		title := "foobar"
		description := "foobar todo"
		dueDateFormated := time.Now().Format("2006-01-02")

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.RawQuery == "" {
				responseData, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatal("can't read request body for testserver")
				}

				err = json.Unmarshal(responseData, &todoInReq)
				if err != nil {
					t.Fatal("can't unmarshal json for testserver request")
				}
				rw.Write([]byte("OK"))
			} else {
				data, err := json.Marshal([]Task{})
				if err != nil {
					t.Fatal("can't marshal json for testserver answer")
				}
				rw.Write(data)
			}
		}))
		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL, apiKey: "XXX", projectId: id}
		err := todoist.CreateTodo(title, description)

		assertNoError(t, err)
		assertEqualString(t, *todoInReq.ProjectId, id)
		assertEqualString(t, *todoInReq.Content, title)
		assertEqualString(t, *todoInReq.Description, description)
		if len(todoInReq.Labels) != 2 {
			t.Fatalf("todo should have 2 labels but todo %d", len(todoInReq.Labels))
		}
		assertEqualString(t, todoInReq.Labels[0], title)
		assertEqualString(t, todoInReq.Labels[1], dueDateFormated)
	})

	t.Run("It should return error when an error happens in createTodoDTO", func(t *testing.T) {
		title := "foobar"
		dueDateFormated := time.Now().Format("2006-01-02")

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if strings.Contains(req.URL.RawQuery, dueDateFormated) {
				http.Error(rw, "oups", http.StatusInternalServerError)
			} else {
				data, err := json.Marshal([]Task{})
				if err != nil {
					t.Fatal("can't marshal json for testserver answer")
				}
				rw.Write(data)
			}
		}))
		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL, apiKey: "XXX", projectId: id}
		err := todoist.CreateTodo(title, title)
		if err == nil {
			t.Fatal("didn't get any error but wanted one")
		}

		assertEqualString(t, err.Error(), fmt.Sprintf("%s: oups\n", ErrHttpRequestDefault))
	})

	t.Run("It should return error when an error on POST task", func(t *testing.T) {
		title := "foobar"

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.RawQuery == "" {
				http.Error(rw, "oups", http.StatusInternalServerError)
			} else {
				data, err := json.Marshal([]Task{})
				if err != nil {
					t.Fatal("can't marshal json for testserver answer")
				}
				rw.Write(data)
			}
		}))
		defer server.Close()

		todoist := Todoist{Client: server.Client(), baseUrl: server.URL, apiKey: "XXX", projectId: id}
		err := todoist.CreateTodo(title, title)
		if err == nil {
			t.Fatal("didn't get any error but wanted one")
		}

		assertEqualString(t, err.Error(), fmt.Sprintf("%s: oups\n", ErrHttpRequestDefault))
	})
}
