package github

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/app"
)

type fakeProviderClient struct {
	projectPath  string
	branch       string
	ref          string
	filePath     string
	sourceBranch string
	targetBranch string
	pullNumber   int
	commit       string
	err          error
}

func (f *fakeProviderClient) CreateBranch(_ context.Context, projectPath, branch, ref string) (Reference, error) {
	f.projectPath, f.branch, f.ref = projectPath, branch, ref
	return Reference{Ref: "refs/heads/" + branch}, f.err
}

func (f *fakeProviderClient) CreateOrUpdateFile(_ context.Context, projectPath, branch, filePath, commitMessage, content string) (RepositoryContent, error) {
	f.projectPath, f.branch, f.filePath = projectPath, branch, filePath
	return RepositoryContent{SHA: "filesha"}, f.err
}

func (f *fakeProviderClient) OpenPullRequest(_ context.Context, projectPath, sourceBranch, targetBranch, title, description string) (PullRequest, error) {
	f.projectPath, f.sourceBranch, f.targetBranch = projectPath, sourceBranch, targetBranch
	return PullRequest{Number: 7, HTMLURL: "https://github.example/org/repo/pull/7"}, f.err
}

func (f *fakeProviderClient) FindOpenPullRequest(_ context.Context, projectPath, sourceBranch, targetBranch string) (PullRequest, error) {
	f.projectPath, f.sourceBranch, f.targetBranch = projectPath, sourceBranch, targetBranch
	return PullRequest{Number: 9, HTMLURL: "https://github.example/org/repo/pull/9"}, f.err
}

func (f *fakeProviderClient) MergePullRequest(_ context.Context, projectPath string, number int, commitMessage string) (PullRequestMerge, error) {
	f.projectPath, f.pullNumber, f.commit = projectPath, number, commitMessage
	return PullRequestMerge{SHA: "def456", Merged: true}, f.err
}

func gitHubTarget() app.GitRepositoryTarget {
	return app.GitRepositoryTarget{Provider: "github", ProviderRef: "github-public", ProjectPath: "org/repo", RepositoryURL: "https://github.com/org/repo.git", DefaultBranch: "main"}
}

func TestProviderCreateBranchUsesTargetProjectPath(t *testing.T) {
	client := &fakeProviderClient{}
	provider, err := NewProvider("github-public", client)
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.CreateBranch(context.Background(), gitHubTarget(), "change/CHG-1", "main"); err != nil {
		t.Fatal(err)
	}
	if client.projectPath != "org/repo" || client.branch != "change/CHG-1" || client.ref != "main" {
		t.Fatalf("unexpected call: %#v", client)
	}
}

func TestProviderCreateOrUpdateFileUsesTargetProjectPath(t *testing.T) {
	client := &fakeProviderClient{}
	provider, _ := NewProvider("github-public", client)
	if err := provider.CreateOrUpdateFile(context.Background(), gitHubTarget(), "change/CHG-1", "manifests/change.yaml", "update", "content"); err != nil {
		t.Fatal(err)
	}
	if client.projectPath != "org/repo" || client.filePath != "manifests/change.yaml" {
		t.Fatalf("unexpected call: %#v", client)
	}
}

func TestProviderOpenMergeRequestMapsPullRequest(t *testing.T) {
	client := &fakeProviderClient{}
	provider, _ := NewProvider("github-public", client)
	number, webURL, err := provider.OpenMergeRequest(context.Background(), gitHubTarget(), "change/CHG-1", "main", "title", "description")
	if err != nil {
		t.Fatal(err)
	}
	if number != 7 || webURL != "https://github.example/org/repo/pull/7" || client.projectPath != "org/repo" {
		t.Fatalf("number=%d url=%s client=%#v", number, webURL, client)
	}
}

func TestProviderMergeRequestFindsAndMergesPullRequest(t *testing.T) {
	client := &fakeProviderClient{}
	provider, _ := NewProvider("github-public", client)
	number, webURL, sha, err := provider.MergeRequest(context.Background(), gitHubTarget(), "change/CHG-1", "main", "merge")
	if err != nil {
		t.Fatal(err)
	}
	if number != 9 || webURL != "https://github.example/org/repo/pull/9" || sha != "def456" {
		t.Fatalf("number=%d url=%s sha=%s", number, webURL, sha)
	}
	if client.pullNumber != 9 || client.commit != "merge" || client.projectPath != "org/repo" {
		t.Fatalf("unexpected merge call: %#v", client)
	}
}

func TestProviderRejectsInvalidTargets(t *testing.T) {
	provider, _ := NewProvider("github-public", &fakeProviderClient{})
	cases := []struct {
		name   string
		target app.GitRepositoryTarget
		want   string
	}{
		{"provider mismatch", app.GitRepositoryTarget{Provider: "gitlab", ProviderRef: "github-public", ProjectPath: "org/repo"}, "cannot handle provider"},
		{"providerRef mismatch", app.GitRepositoryTarget{Provider: "github", ProviderRef: "other", ProjectPath: "org/repo"}, "cannot handle target providerRef"},
		{"invalid projectPath", app.GitRepositoryTarget{Provider: "github", ProviderRef: "github-public", ProjectPath: "org"}, "owner/repository format"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := provider.CreateBranch(context.Background(), tc.target, "branch", "main")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v want=%q", err, tc.want)
			}
		})
	}
}

func TestNewProviderValidation(t *testing.T) {
	if _, err := NewProvider("", &fakeProviderClient{}); err == nil {
		t.Fatal("expected providerRef error")
	}
	if _, err := NewProvider("github-public", nil); err == nil {
		t.Fatal("expected client error")
	}
}

func TestProviderPropagatesClientError(t *testing.T) {
	provider, _ := NewProvider("github-public", &fakeProviderClient{err: errors.New("client failed")})
	err := provider.CreateBranch(context.Background(), gitHubTarget(), "branch", "main")
	if err == nil || !strings.Contains(err.Error(), "client failed") {
		t.Fatalf("error=%v", err)
	}
}
