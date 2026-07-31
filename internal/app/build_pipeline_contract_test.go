package app

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestBuildPipelineManifestsAreIdenticalAndExposeResults(t *testing.T) {
	pipeline, err := os.ReadFile("../../pipelines/go-build-and-push.yaml")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := os.ReadFile("../../manifests/tekton/go-build-and-push-pipeline.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pipeline, manifest) {
		t.Fatal("build pipeline copies differ")
	}

	text := string(pipeline)
	required := []string{
		"name: go-build-and-push",
		"devops-control-plane/workflow-type: build",
		"name: GIT_URL",
		"name: GIT_REVISION",
		"name: IMAGE",
		"name: shared-workspace",
		"name: dockerconfig",
		"name: git-clone",
		"name: buildah",
		"name: SOURCE_COMMIT",
		"value: $(tasks.clone-repository.results.COMMIT)",
		"value: $(tasks.build-and-push.results.IMAGE_URL)",
		"value: $(tasks.build-and-push.results.IMAGE_DIGEST)",
	}
	for _, value := range required {
		if !strings.Contains(text, value) {
			t.Fatalf("build pipeline does not contain %q", value)
		}
	}

	forbidden := []string{"PRIVATE-TOKEN", "GITHUB_TOKEN", "GITLAB_TOKEN", "client_secret", "password:"}
	for _, value := range forbidden {
		if strings.Contains(text, value) {
			t.Fatalf("build pipeline contains sensitive field %q", value)
		}
	}
}
