package git

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/6de1ay/auto-semver-tag/pkg/semver"
	"github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
)

const (
	IncrementTypeMajorLabel = "Major"
	IncrementTypeMinorLabel = "Minor"
)

type GitRepository struct {
	name          string
	owner         string
	releaseBranch string
	version       semver.SemVer
}

type GitClient struct {
	token  string
	repo   GitRepository
	client *github.Client
}

func New(token string, repository string, releaseBranch string) *GitClient {
	ctx := context.Background()

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	client := github.NewClient(oauth2.NewClient(ctx, tokenSource))

	parts := strings.Split(repository, "/")
	owner := parts[0]
	repoName := parts[1]

	version := getLatestTag(client, owner, repoName)

	repo := GitRepository{
		repoName,
		owner,
		releaseBranch,
		version,
	}

	return &GitClient{
		token,
		repo,
		client,
	}
}

func (gitClient *GitClient) PerformAction(commitSha string, eventDataFilePath string) {
	event := parseEventDataFile(eventDataFilePath)

	if event.Action == nil || *event.Action != "closed" {
		log.Fatal("pull request is not closed")
	}

	if event.PullRequest.Merged == nil || !*event.PullRequest.Merged {
		log.Fatal("pull request is not merged")
	}

	if event.PullRequest.Base == nil || event.PullRequest.Base.Ref == nil {
		log.Fatal("could not determine pull request base branch")
	}

	if *event.PullRequest.Base.Ref != gitClient.repo.releaseBranch {
		log.Fatal("pull request is merged not into the release branch")
	}

	incrementType := getIncrementTypeFromLabels(event.PullRequest)
	newVersion := gitClient.repo.version.IncrementVersion(incrementType)

	gitClient.createTag(newVersion.String(), commitSha)
}

func (gitClient *GitClient) createTag(version string, commitSha string) error {
	ctx := context.Background()
	ref := &github.Reference{
		Ref: github.String(fmt.Sprintf("refs/tags/%s", version)),
		Object: &github.GitObject{
			SHA: &commitSha,
		},
	}

	_, _, err := gitClient.client.Git.CreateRef(ctx, gitClient.repo.owner, gitClient.repo.name, ref)

	return err
}

func getIncrementTypeFromLabels(pr *github.PullRequest) string {
	isValidLabelFound := false
	incrementType := semver.DefaultIncrementType

	for _, label := range pr.Labels {
		if label.Name == nil {
			continue
		}

		if *label.Name == IncrementTypeMajorLabel {
			if isValidLabelFound {
				panic("several valid labels found")
			}

			isValidLabelFound = true
			incrementType = semver.IncrementTypeMajor
			continue
		}

		if *label.Name == IncrementTypeMinorLabel {
			if isValidLabelFound {
				panic("several valid labels found")
			}

			isValidLabelFound = true
			incrementType = semver.IncrementTypeMinor
			continue
		}

	}

	return incrementType
}

func parseEventDataFile(filePath string) *github.PullRequestEvent {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("%s. Filepath: %s", err, filePath)
	}
	defer file.Close()

	event, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("%s. Filepath: %s", err, filePath)
	}

	eventData, err := github.ParseWebHook("pull_request", stripOrg(event))
	if err != nil {
		log.Fatalf("%s. Filepath: %s", err, filePath)
	}

	res, ok := eventData.(*github.PullRequestEvent)
	if !ok {
		log.Fatal("could not parse GitHub event into a PullRequestEvent")
	}

	return res
}

func getLatestTag(client *github.Client, owner string, repo string) semver.SemVer {
	res := semver.SemVer{}
	ctx := context.Background()

	refs, response, err := client.Git.ListMatchingRefs(ctx, owner, repo, &github.ReferenceListOptions{
		Ref: "tags",
	})
	if err != nil {
		log.Fatal(err)
	}

	if response != nil && response.StatusCode == http.StatusNotFound {
		return res
	}

	for _, ref := range refs {
		version, err := semver.New(strings.Replace(*ref.Ref, "refs/tags/", "", 1))
		if err != nil {
			continue
		}

		if version.IsGreaterThan(res) {
			res = version
		}
	}

	return res
}

func stripOrg(byteString []byte) []byte {
	// workaround for https://github.com/google/go-github/issues/131
	var o map[string]interface{}
	_ = json.Unmarshal(byteString, &o)
	if o != nil {
		repo := o["repository"]
		if repo != nil {
			if repo, ok := repo.(map[string]interface{}); ok {
				delete(repo, "organization")
			}
		}
	}
	b, _ := json.MarshalIndent(o, "", "  ")
	return b
}
