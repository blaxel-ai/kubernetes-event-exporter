package main

import (
	"os"
	"strings"
	"testing"
)

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func assertContains(t *testing.T, content, expected string) {
	t.Helper()

	if !strings.Contains(content, expected) {
		t.Fatalf("expected content to contain %q", expected)
	}
}

func assertNotContains(t *testing.T, content, unexpected string) {
	t.Helper()

	if strings.Contains(content, unexpected) {
		t.Fatalf("expected content not to contain %q", unexpected)
	}
}

func TestBranchDockerBuildWorkflowUsesPersistentRegistryCache(t *testing.T) {
	content := readFile(t, ".github/workflows/build.yml")

	assertContains(t, content, "cache-from: type=registry,ref=ghcr.io/${{ github.repository }}:buildcache-${{ github.ref_name }}")
	assertContains(t, content, "cache-to: type=registry,ref=ghcr.io/${{ github.repository }}:buildcache-${{ github.ref_name }},mode=max")
}

func TestReleaseDockerBuildWorkflowUsesStableCrossReleaseCache(t *testing.T) {
	content := readFile(t, ".github/workflows/release.yml")

	assertContains(t, content, "on:\n  release:\n    types: [published]")
	assertContains(t, content, "cache-from: |\n            type=registry,ref=ghcr.io/${{ github.repository }}:buildcache-release\n            type=registry,ref=ghcr.io/${{ github.repository }}:buildcache-main")
	assertContains(t, content, "cache-to: type=registry,ref=ghcr.io/${{ github.repository }}:buildcache-release,mode=max")
	assertNotContains(t, content, "buildcache-${{ github.ref_name }}")
}

func TestDockerfileKeepsVersionArgOutOfDependencyCacheLayers(t *testing.T) {
	content := readFile(t, "Dockerfile")

	assertContains(t, content, "COPY go.mod go.sum ./")
	assertContains(t, content, "RUN go mod download")
	assertContains(t, content, "COPY . .")
	assertContains(t, content, "ARG VERSION")
	assertContains(t, content, "--mount=type=cache,target=/root/.cache/go-build")
	assertNotContains(t, content, "--mount=type=cache,target=/go/pkg/mod")

	argVersionIndex := strings.Index(content, "ARG VERSION")
	goModDownloadIndex := strings.Index(content, "go mod download")
	copySourceIndex := strings.Index(content, "COPY . .")
	if argVersionIndex < 0 || goModDownloadIndex < 0 || copySourceIndex < 0 {
		t.Fatalf("Dockerfile is missing expected dependency/build stages")
	}
	if argVersionIndex < goModDownloadIndex || argVersionIndex < copySourceIndex {
		t.Fatalf("ARG VERSION should appear after dependency download and source copy so commit SHA changes only invalidate the final build layer")
	}
	if strings.Contains(content, "go build ") && strings.Contains(content, " -a ") {
		t.Fatalf("Dockerfile should not force rebuilding all packages with go build -a")
	}
}
