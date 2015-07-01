package pull

import "github.com/google/go-github/github"

type runnerFunc func(*Handler, github.PullRequestEvent)

func Success(url string, desc string, c string) *github.RepoStatus {
	return &github.RepoStatus{
		State:       github.String("success"),
		TargetURL:   github.String(url),
		Description: github.String(desc),
		Context:     github.String("pull/" + c),
	}
}

func Failure(url string, desc string, c string) *github.RepoStatus {
	return &github.RepoStatus{
		State:       github.String("failure"),
		TargetURL:   github.String(url),
		Description: github.String(desc),
		Context:     github.String("pull/" + c),
	}
}

func Data(req *github.PullRequest) (user string, repository string, number int) {
	user = *req.Base.User.Login
	repository = *req.Base.Repo.Name
	number = *req.Number
	return
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

func runner(handler *Handler, event github.PullRequestEvent) {
	for _, m := range handler.Middlewares[*event.Action] {
		m(handler.Client, event.PullRequest)
	}
}
