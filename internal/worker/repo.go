package worker

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	git_config "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	git_http "github.com/go-git/go-git/v5/plumbing/transport/http"
	"golang.org/x/oauth2"
)

func cloneRepo(ctx context.Context, cloneURL string, sha string, token oauth2.Token) (dir string, err error) {
	dir, err = os.MkdirTemp("/tmp", "shark-ci-*")
	if err != nil {
		return "", err
	}

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		return "", err
	}

	_, err = repo.CreateRemote(&git_config.RemoteConfig{
		Name: "origin",
		URLs: []string{cloneURL},
	})
	if err != nil {
		return "", err
	}

	err = repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: "origin",
		Depth:      1,
		RefSpecs: []git_config.RefSpec{
			git_config.RefSpec(fmt.Sprintf("%s:refs/heads/test", sha)),
		},
		Auth: &git_http.BasicAuth{
			Username: "abc",
			Password: token.AccessToken,
		},
		Progress: log.Writer(),
	})
	if err != nil {
		return "", err
	}

	tree, err := repo.Worktree()
	if err != nil {
		return "", err
	}
	err = tree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(sha),
	})
	if err != nil {
		return "", err
	}

	return dir, err
}
