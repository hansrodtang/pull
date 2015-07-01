package cla

import (
	"github.com/google/go-github/github"
	"github.com/hansrodtang/pull"
)

type CheckerFunc func(username, email string) bool

func New(signed CheckerFunc) pull.Middleware {
	return func(client *github.Client, event *github.PullRequest) {
		user, repo, number := pull.Data(event)

		commits, _, _ := client.PullRequests.ListCommits(user, repo, number, nil)

		for _, c := range commits {
			var message *github.RepoStatus
			if signed(*c.Author.Login, *c.Commit.Author.Email) {
				message = pull.Success("", "CLA signed", "cla")
			} else {
				message = pull.Failure("", "CLA not signed", "cla")
			}
			client.Repositories.CreateStatus(user, repo, *c.SHA, message)
		}
	}
}
