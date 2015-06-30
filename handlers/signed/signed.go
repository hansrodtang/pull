package signed

import (
	"fmt"
	"regexp"

	"github.com/google/go-github/github"
	"github.com/hansrodtang/pull"
)

func Signed(client *github.Client, event *github.PullRequest) {
	user, repo, number := pull.Data(event)

	commits, _, _ := client.PullRequests.ListCommits(user, repo, number, nil)

	for _, c := range commits {
		r := fmt.Sprintf("Signed-off-by: .* <%s>", *c.Commit.Author.Email)
		match, _ := regexp.Match(r, []byte(*c.Commit.Message))

		if match {
			success := pull.Success("", "Commit is signed-off")
			client.Repositories.CreateStatus(user, repo, *c.SHA, success)
		} else {
			s := fmt.Sprintf("Commit not signed off: [%s](%s).", *c.SHA, *c.HTMLURL)
			comment := &github.IssueComment{Body: &s}

			cm, _, _ := client.Issues.CreateComment(user, repo, number, comment)
			failure := pull.Failure(*cm.HTMLURL, "Commit is not signed-off")

			client.Repositories.CreateStatus(user, repo, *c.SHA, failure)
		}
	}
}
