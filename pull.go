package pull

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/github"
)

type Middleware func(*github.Client, *github.PullRequest)

type Handler struct {
	Client *github.Client
	*Configuration
}

type Configuration struct {
	Client       *http.Client
	Repositories []Repository
	Middlewares  []Middleware
	Secret       string
}

type Repository struct {
	Username string
	Name     string
}

func New(c Configuration) *Handler {
	return &Handler{github.NewClient(c.Client), &c}
}

func (l *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	w := io.MultiWriter(resp, logger{})

	t := req.Header.Get("X-GitHub-Event")
	if t == "ping" {
		resp.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Received ping")
		return
	}

	if t != "pull_request" {
		resp.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintf(w, "Unsupported event")
		return
	}

	body, _ := ioutil.ReadAll(req.Body)

	if l.Secret != "" {
		s := req.Header.Get("X-Hub-Signature")

		if s == "" {
			resp.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "X-Hub-Signature required for HMAC verification")
			return
		}

		hash := hmac.New(sha1.New, []byte(l.Secret))
		hash.Write(body)
		expected := "sha1=" + hex.EncodeToString(hash.Sum(nil))

		if !hmac.Equal([]byte(expected), []byte(s)) {
			resp.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "HMAC verification failed")
			return
		}
	}

	var event github.PullRequestEvent
	err := json.Unmarshal(body, &event)
	//err := decoder.Decode(&event)

	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad JSON")
		return
	}
	fmt.Printf("%s %s:%d\n", *event.Action, *event.Repo.FullName, *event.Number)

	pending(l.Client, event.PullRequest)
	go l.runner(event.PullRequest)
}

func (l *Handler) runner(pr *github.PullRequest) {
	for _, m := range l.Middlewares {
		m(l.Client, pr)
	}
}

func pending(client *github.Client, event *github.PullRequest) {
	user, repo, number := Data(event)

	commits, _, _ := client.PullRequests.ListCommits(user, repo, number, nil)
	for _, c := range commits {
		status := &github.RepoStatus{
			State:       github.String("pending"),
			Description: github.String("Running tests."),
		}

		client.Repositories.CreateStatus(user, repo, *c.SHA, status)
	}
}
