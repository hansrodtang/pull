package pull

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/github"
)

type Middleware func(*github.Client, *github.PullRequest)
type Middlewares map[string][]Middleware

type Handler struct {
	Client *github.Client
	*Configuration
}

type Configuration struct {
	Client      *http.Client
	Middlewares Middlewares
	Secret      string
}

func New(c Configuration) *Handler {
	return &Handler{github.NewClient(c.Client), &c}
}

func (l *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pingHandler(implementsHandler(secretHandler(l.Secret,
		http.HandlerFunc(
			func(w http.ResponseWriter, req *http.Request) {

				decoder := json.NewDecoder(req.Body)
				var event github.PullRequestEvent
				err := decoder.Decode(&event)

				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintf(w, "Bad JSON")
					return
				}
				fmt.Printf("%s %s:%d\n", *event.Action, *event.Repo.FullName, *event.Number)

				//pending(l.Client, event.PullRequest)
				go func() {
					for _, m := range l.Middlewares[*event.Action] {
						m(l.Client, event.PullRequest)
					}
				}()
			})))).ServeHTTP(w, req)
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

func pingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t := req.Header.Get("X-GitHub-Event")
		if t == "ping" {
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprintf(w, "Received ping")
			return
		}
		next.ServeHTTP(w, req)
	})
}

func implementsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t := req.Header.Get("X-GitHub-Event")
		if t != "pull_request" {
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprintf(w, "Unsupported event")
			return
		}
		next.ServeHTTP(w, req)
	})
}

func secretHandler(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if secret != "" {
			s := req.Header.Get("X-Hub-Signature")

			if s == "" {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, "X-Hub-Signature required for HMAC verification")
				return
			}

			body, _ := ioutil.ReadAll(req.Body)

			hash := hmac.New(sha1.New, []byte(secret))
			hash.Write(body)
			expected := "sha1=" + hex.EncodeToString(hash.Sum(nil))

			if !hmac.Equal([]byte(expected), []byte(s)) {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, "HMAC verification failed")
				return
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}

		next.ServeHTTP(w, req)
	})
}
