package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/MichaelMure/git-bug/bridge/core"
	"github.com/MichaelMure/git-bug/bug"
	"github.com/MichaelMure/git-bug/cache"
	"github.com/MichaelMure/git-bug/repository"
	"github.com/MichaelMure/git-bug/util/interrupt"
)

const (
	testRepoBaseName = "git-bug-test-github-exporter"
)

// testCases creates bugs in repo cache
func testCases(repo *cache.RepoCache) (map[string]*cache.BugCache, error) {
	author, err := repo.NewIdentity("test identity", "hello@testidentity.org")
	if err != nil {
		return nil, err
	}

	bugs := make(map[string]*cache.BugCache)

	// simple bug
	simpleBug, err := repo.NewBugRaw(author, time.Now().Unix(), "simple bug", "new bug", nil, nil)
	if err != nil {
		return nil, err
	}
	bugs["simple bug"] = simpleBug

	// bug with comments
	bugWithComments, err := repo.NewBugRaw(author, time.Now().Unix(), "bug with comments", "new bug", nil, nil)
	if err != nil {
		return nil, err
	}

	_, err = bugWithComments.AddCommentRaw(author, time.Now().Unix(), "new comment", nil, nil)
	if err != nil {
		return nil, err
	}
	bugs["bug with comments"] = bugWithComments

	// bug with label changes
	bugLabelChange, err := repo.NewBugRaw(author, time.Now().Unix(), "bug label change", "new bug", nil, nil)
	if err != nil {
		return nil, err
	}

	_, _, err = bugLabelChange.ChangeLabelsRaw(author, time.Now().Unix(), []string{"bug", "core"}, nil, nil)
	if err != nil {
		return nil, err
	}

	_, _, err = bugLabelChange.ChangeLabelsRaw(author, time.Now().Unix(), nil, []string{"bug"}, nil)
	if err != nil {
		return nil, err
	}
	bugs["bug change label"] = bugWithComments

	return nil, err
}

func TestExporter(t *testing.T) {
	user := os.Getenv("TEST_USER")
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		t.Skip("Env var GITHUB_TOKEN_PRIVATE missing")
	}

	repo := repository.CreateTestRepo(false)
	defer repository.CleanupTestRepos(t, repo)

	backend, err := cache.NewRepoCache(repo)
	require.NoError(t, err)

	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	bugs, err := testCases(backend)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		operations []bug.Operation
	}{
		{
			name:       "simple bug",
			operations: bugs["simple bug"].Snapshot().Operations,
		},
		{
			name:       "bug with comments",
			operations: bugs["bug with comments"].Snapshot().Operations,
		},
		{
			name:       "bug label change",
			operations: bugs["bug change label"].Snapshot().Operations,
		},
		/*
			{
				name: "bug status edition",
			},
			{
				name: "bug title edition",
			},
			{
				name: "bug comments edition",
			},
			{
				name: "complex bug",
			},
		*/
	}
	// generate project name
	projectName := generateRepoName()
	if err := createRepository(projectName, token); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := deleteRepository(projectName, user, token); err != nil {
			t.Fatal(err)
		}
	}()

	exporter := &githubExporter{}
	err = exporter.Init(core.Configuration{
		keyOwner:   user,
		keyProject: projectName,
		keyToken:   token,
	})
	require.NoError(t, err)

	start := time.Now()

	err = exporter.ExportAll(backend, time.Time{})
	require.NoError(t, err)

	fmt.Printf("test repository exported in %f seconds\n", time.Since(start).Seconds())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

		})
	}
}

func generateRepoName() string {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return fmt.Sprintf("%s-%s", testRepoBaseName, string(b))
}

func createRepository(project, token string) error {
	url := fmt.Sprintf("%s/user/repos", githubV3Url)

	params := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Private     bool   `json:"private"`
		HasIssues   bool   `json:"has_issues"`
	}{
		Name:        project,
		Description: "git-bug exporter temporary test repository",
		Private:     true,
		HasIssues:   true,
	}

	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	// need the token for private repositories
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	client := &http.Client{
		Timeout: defaultTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	return resp.Body.Close()
}

func deleteRepository(project, owner, token string) error {
	url := fmt.Sprintf("%s/repos/%s/%s", githubV3Url, owner, project)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	// need the token for private repositories
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	client := &http.Client{
		Timeout: defaultTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	return resp.Body.Close()
}
