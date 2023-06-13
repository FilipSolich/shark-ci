package worker

import (
	"context"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	git_http "github.com/go-git/go-git/v5/plumbing/transport/http"
)

func CreateRepoDir(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}

func UpdateRepo(ctx context.Context, path string, cloneURL string, token string) (*git.Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return cloneRepo(ctx, path, cloneURL, token)
		}
		return nil, err
	}

	err = repo.FetchContext(ctx, &git.FetchOptions{
		Auth: &git_http.TokenAuth{
			Token: token,
		},
		Progress: log.Writer(),
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, err
	}

	return repo, nil
}

func cloneRepo(ctx context.Context, path string, cloneURL string, token string) (*git.Repository, error) {
	repo, err := git.PlainCloneContext(ctx, path, false, &git.CloneOptions{
		URL: cloneURL,
		Auth: &git_http.TokenAuth{
			Token: token,
		},
		Progress: log.Writer(),
	})
	return repo, err
}
