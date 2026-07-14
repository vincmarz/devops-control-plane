package gitlab

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/app"
)

type fakeProviderClient struct {
	projectID    int
	branch       string
	ref          string
	filePath     string
	sourceBranch string
	targetBranch string
	mergeIID     int
	mergeSHA     string
	err          error
}

func (f *fakeProviderClient) CreateBranch(_ context.Context, projectID int, branch, ref string) (Branch, error) {
	f.projectID, f.branch, f.ref = projectID, branch, ref
	return Branch{Name: branch}, f.err
}
func (f *fakeProviderClient) CreateOrUpdateFile(_ context.Context, projectID int, branch, filePath, commitMessage, content string) (RepositoryFile, error) {
	f.projectID, f.branch, f.filePath = projectID, branch, filePath
	return RepositoryFile{FilePath: filePath, Branch: branch}, f.err
}
func (f *fakeProviderClient) OpenMergeRequest(_ context.Context, projectID int, sourceBranch, targetBranch, title, description string) (MergeRequest, error) {
	f.projectID, f.sourceBranch, f.targetBranch = projectID, sourceBranch, targetBranch
	return MergeRequest{IID: 7, WebURL: "https://gitlab.example/mr/7"}, f.err
}
func (f *fakeProviderClient) FindOpenMergeRequest(_ context.Context, projectID int, sourceBranch, targetBranch string) (MergeRequest, error) {
	f.projectID, f.sourceBranch, f.targetBranch = projectID, sourceBranch, targetBranch
	return MergeRequest{IID: 9, SHA: "abc123"}, f.err
}
func (f *fakeProviderClient) MergeMergeRequest(_ context.Context, projectID, mergeRequestIID int, sha, mergeCommitMessage string) (MergeRequest, error) {
	f.projectID, f.mergeIID, f.mergeSHA = projectID, mergeRequestIID, sha
	return MergeRequest{IID: mergeRequestIID, WebURL: "https://gitlab.example/mr/9", MergeCommitSHA: "def456"}, f.err
}

func gitLabTarget() app.GitRepositoryTarget {
	return app.GitRepositoryTarget{Provider: "gitlab", ProviderRef: "gitlab-lab", ProjectID: 42, ProjectPath: "group/app", RepositoryURL: "https://gitlab.example/group/app.git", DefaultBranch: "main"}
}

func TestProviderCreateBranchUsesTargetProject(t *testing.T) {
	client := &fakeProviderClient{}
	provider, err := NewProvider("gitlab-lab", client)
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.CreateBranch(context.Background(), gitLabTarget(), "change/CHG-1", "main"); err != nil {
		t.Fatal(err)
	}
	if client.projectID != 42 || client.branch != "change/CHG-1" || client.ref != "main" {
		t.Fatalf("unexpected call: %#v", client)
	}
}

func TestProviderCreateOrUpdateFileUsesTargetProject(t *testing.T) {
	client := &fakeProviderClient{}
	provider, _ := NewProvider("gitlab-lab", client)
	if err := provider.CreateOrUpdateFile(context.Background(), gitLabTarget(), "change/CHG-1", "manifests/change.yaml", "update", "content"); err != nil {
		t.Fatal(err)
	}
	if client.projectID != 42 || client.filePath != "manifests/change.yaml" {
		t.Fatalf("unexpected call: %#v", client)
	}
}

func TestProviderOpenMergeRequestReturnsNeutralResult(t *testing.T) {
	client := &fakeProviderClient{}
	provider, _ := NewProvider("gitlab-lab", client)
	iid, webURL, err := provider.OpenMergeRequest(context.Background(), gitLabTarget(), "change/CHG-1", "main", "title", "description")
	if err != nil {
		t.Fatal(err)
	}
	if iid != 7 || webURL != "https://gitlab.example/mr/7" || client.projectID != 42 {
		t.Fatalf("iid=%d url=%s client=%#v", iid, webURL, client)
	}
}

func TestProviderMergeRequestFindsAndMerges(t *testing.T) {
	client := &fakeProviderClient{}
	provider, _ := NewProvider("gitlab-lab", client)
	iid, webURL, sha, err := provider.MergeRequest(context.Background(), gitLabTarget(), "change/CHG-1", "main", "merge")
	if err != nil {
		t.Fatal(err)
	}
	if iid != 9 || webURL != "https://gitlab.example/mr/9" || sha != "def456" {
		t.Fatalf("iid=%d url=%s sha=%s", iid, webURL, sha)
	}
	if client.mergeIID != 9 || client.mergeSHA != "abc123" {
		t.Fatalf("unexpected merge call: %#v", client)
	}
}

func TestProviderRejectsInvalidTargets(t *testing.T) {
	provider, _ := NewProvider("gitlab-lab", &fakeProviderClient{})
	cases := []struct {
		name   string
		target app.GitRepositoryTarget
		want   string
	}{
		{"provider mismatch", app.GitRepositoryTarget{Provider: "github", ProviderRef: "gitlab-lab", ProjectID: 1}, "cannot handle provider"},
		{"providerRef mismatch", app.GitRepositoryTarget{Provider: "gitlab", ProviderRef: "other", ProjectID: 1}, "cannot handle target providerRef"},
		{"invalid projectID", app.GitRepositoryTarget{Provider: "gitlab", ProviderRef: "gitlab-lab"}, "projectID must be greater than zero"},
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
	if _, err := NewProvider("gitlab-lab", nil); err == nil {
		t.Fatal("expected client error")
	}
}

func TestProviderPropagatesClientError(t *testing.T) {
	provider, _ := NewProvider("gitlab-lab", &fakeProviderClient{err: errors.New("client failed")})
	err := provider.CreateBranch(context.Background(), gitLabTarget(), "branch", "main")
	if err == nil || !strings.Contains(err.Error(), "client failed") {
		t.Fatalf("error=%v", err)
	}
}
