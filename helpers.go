package pull

import (
	"fmt"
	"os"

	"github.com/google/go-github/github"
)

func Success(url string, desc string) *github.RepoStatus {
	return &github.RepoStatus{
		State:       github.String("success"),
		TargetURL:   github.String(url),
		Description: github.String(desc),
	}
}

func Failure(url string, desc string) *github.RepoStatus {
	return &github.RepoStatus{
		State:       github.String("failure"),
		TargetURL:   github.String(url),
		Description: github.String(desc),
	}
}

func Data(req *github.PullRequest) (user string, repository string, number int) {
	user = *req.Base.User.Login
	repository = *req.Base.Repo.Name
	number = *req.Number
	return
}
