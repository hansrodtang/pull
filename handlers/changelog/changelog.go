package changelog

import (
	"strings"

	"github.com/google/go-github/github"
	"github.com/hansrodtang/pull"
)

func New(match string) pull.Middleware {
	return func(client *github.Client, event *github.PullRequest) {
		user, repo, number := pull.Data(event)

		commits, _, _ := client.PullRequests.ListCommits(user, repo, number, nil)
		files, _, _ := client.PullRequests.ListFiles(user, repo, number, nil)

		found := false
		for _, file := range files {
			found = strings.EqualFold(*file.Filename, match)
		}

		var message *github.RepoStatus
		url := "http://keepachangelog.com/"
		if found {
			message = pull.Success(url, "ChangeLog Kept", "changelog")
		} else {
			message = pull.Failure(url, "Keep a ChangeLog", "changelog")
		}

		for _, c := range commits {
			client.Repositories.CreateStatus(user, repo, *c.SHA, message)
		}

	}

}
