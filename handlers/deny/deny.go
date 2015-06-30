package deny

import (
	"github.com/google/go-github/github"
	"github.com/hansrodtang/pull"
)

func New(message string) pull.Middleware {
	return func(client *github.Client, event *github.PullRequest) {
		user, repo, number := pull.Data(event)

		config := &github.IssueRequest{
			State: github.String("closed"),
		}

		comment := &github.IssueComment{Body: github.String(message)}

		client.Issues.CreateComment(user, repo, number, comment)
		client.Issues.Edit(user, repo, number, config)

	}
}
