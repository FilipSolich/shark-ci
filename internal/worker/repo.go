package worker

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	git_http "github.com/go-git/go-git/v5/plumbing/transport/http"
)

func CreateTmpDir() (string, error) {
	return os.MkdirTemp("/tmp", "shark-ci-*")
}

func updateRepo(ctx context.Context, path string, cloneURL string, token string) (*git.Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return cloneRepo(ctx, path, cloneURL, token)
		}
		return nil, err
	}

	err = repo.FetchContext(ctx, &git.FetchOptions{
		Auth: &git_http.BasicAuth{
			Username: "abc",
			Password: token,
		},
		Progress: log.Writer(),
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil, err
	}

	return repo, nil
}

func cloneRepo(ctx context.Context, path string, cloneURL string, token string) (*git.Repository, error) {
	repo, err := git.PlainCloneContext(ctx, path, false, &git.CloneOptions{
		URL: cloneURL,
		Auth: &git_http.BasicAuth{
			Username: "abc",
			Password: token,
		},
		Progress: log.Writer(),
	})
	return repo, err
}

func archiveRepo(repoPath string, dir string, repoName string, repoSha string) (string, error) {
	target := path.Join(dir, strings.ReplaceAll(repoName, "/", "_")+repoSha+".tar.gz")
	_, err := os.Stat(target)
	if err == nil {
		return target, nil
	} else {
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}

	file, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	err = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return target, nil
}
