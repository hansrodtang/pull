package pull

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

// Simple test to see if middlewares are run.
func TestRunner(t *testing.T) {

	expected := "test"

	mw := func(client *github.Client, event *github.PullRequest) {
		client.UserAgent = expected
	}

	handler := New(
		Configuration{
			Middlewares: Middlewares{
				"opened": []Middleware{mw}}})

	event := github.PullRequestEvent{
		Action: github.String("opened"),
	}

	runner(handler, event)

	actual := handler.Client.UserAgent

	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestRepoStatus(t *testing.T) {

	expected := &github.RepoStatus{
		State:       github.String("failure"),
		TargetURL:   github.String("example.com"),
		Description: github.String("test"),
		Context:     github.String("pull/failure"),
	}

	actual := Failure("example.com", "test", "failure")

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %+v, got %+v", expected, actual)
	}

	expected = &github.RepoStatus{
		State:       github.String("success"),
		TargetURL:   github.String("example.com"),
		Description: github.String("test"),
		Context:     github.String("pull/failure"),
	}

	actual = Success("example.com", "test", "failure")

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %+v, got %+v", expected, actual)
	}
}

func TestPending(t *testing.T) {
	user := "test"
	repo := "repo"
	number := 1

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	client := github.NewClient(nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
	client.UploadURL = url

	status := &github.RepoStatus{
		State:       github.String("pending"),
		Description: github.String("Running tests."),
	}

	statusendpoint := fmt.Sprintf("/repos/%s/%s/statuses/", user, repo)
	mux.HandleFunc(statusendpoint, func(w http.ResponseWriter, r *http.Request) {
		body := new(github.RepoStatus)
		json.NewDecoder(r.Body).Decode(body)

		if !reflect.DeepEqual(body, status) {
			t.Errorf("Expected %+v, got %+v", status, body)
		}
		fmt.Fprint(w, `{"id":1}`)
	})

	commitendpoint := fmt.Sprintf("/repos/%s/%s/pulls/%d/commits", user, repo, number)
	mux.HandleFunc(commitendpoint, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"sha": "3", "parents": [{"sha": "2"}]},
			            {"sha": "2","parents":  [{"sha": "1"}]}]`)
	})

	req := &github.PullRequest{
		Number: &number,
		Base: &github.PullRequestBranch{
			User: &github.User{
				Login: &user,
			},
			Repo: &github.Repository{
				Name: &repo,
			},
		},
	}

	pending(client, req)
}
