package todoist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

	t.Run("It should return not itialized error API_KEY not set", func(t *testing.T) {
		todoist := Todoist{}
		err := todoist.CreateTodo("test", "test")

		assertError(t, err, ErrNotInitialized)
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
}
