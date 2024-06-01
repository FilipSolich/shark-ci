package worker

import (
	"context"
	"fmt"
	"testing"

	"golang.org/x/oauth2"
)

func TestCloneRepo(t *testing.T) {
	dir, cleanFunc, err := cloneRepo(context.TODO(), "https://github.com/FilipSolich/githubtest.git", "ac7898b622987634aa20b420c44c71efc088865e", oauth2.Token{AccessToken: "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"})
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	defer func() {
		if err := cleanFunc(); err != nil {
			t.Errorf("Error: %v", err)
		}
	}()
	fmt.Println(dir)
	if dir == "" {
		t.Errorf("Dir name is empty")
	}
}
