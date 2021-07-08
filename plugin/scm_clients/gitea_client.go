package scm_clients

import (
	"context"
	"errors"

	"code.gitea.io/sdk/gitea"
	"github.com/drone/drone-go/drone"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type GiteaClient struct {
	delegate *gitea.Client
	repo     drone.Repo
}

func NewGiteaClient(ctx context.Context, uuid uuid.UUID, server string, token string, repo drone.Repo) (ScmClient, error) {
	var client *gitea.Client
	var err error

	if server == "" {
		server = "https://gitea.com"
	}
	client, err = gitea.NewClient(server, gitea.SetToken(token), gitea.SetContext(ctx))

	if err != nil {
		logrus.Errorf("%s Unable to connect to Github: '%v'", uuid, err)
		return nil, err
	}
	return GiteaClient{
		delegate: client,
		repo:     repo,
	}, nil
}

func (s GiteaClient) ChangedFilesInPullRequest(ctx context.Context, pullRequestID int) ([]string, error) {

	// TODO: Will Change the implementation in case gitea-sdk is support GetPullRequestCommits directly.
	// See: https://gitea.com/api/swagger#/repository/repoGetPullRequestCommits
	var changedFiles []string
	pullRequest, _, err := s.delegate.GetPullRequest(s.repo.Namespace, s.repo.Name, int64(pullRequestID))
	if err != nil {
		return nil, err
	}

	cm, _, err := s.delegate.GetSingleCommit(s.repo.Namespace, s.repo.Name, *pullRequest.MergedCommitID)
	if err != nil {
		return nil, err
	}

	for _, file := range cm.Files {
		changedFiles = append(changedFiles, file.Filename)
	}

	return changedFiles, nil
}

func (s GiteaClient) ChangedFilesInDiff(ctx context.Context, base string, head string) ([]string, error) {

	// TODO: Gitea now is not support get diffs between branches via Web API.
	var changedFiles []string

	return changedFiles, errors.New("getting changed files in diff via Web API is not supported by Gitea yet")
}

func (s GiteaClient) GetFileContents(ctx context.Context, path string, commitRef string) (fileContent string, err error) {

	file, _, err := s.delegate.GetContents(s.repo.Namespace, s.repo.Name, commitRef, path)
	if err != nil {
		return "", err
	}
	var content string
	if file.Content != nil {
		content = *file.Content
	}

	return content, err
}

func (s GiteaClient) GetFileListing(ctx context.Context, path string, commitRef string) (fileListing []FileListingEntry, err error) {

	dir, _, err := s.delegate.ListContents(s.repo.Namespace, s.repo.Name, commitRef, path)
	var result []FileListingEntry
	if err != nil {
		return result, err
	}
	for _, file := range dir {
		fileListingEntry := FileListingEntry{
			Path: file.Path,
			Name: file.Name,
			Type: file.Type,
		}
		result = append(result, fileListingEntry)
	}

	return result, err
}
