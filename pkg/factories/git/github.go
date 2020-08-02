package git

import (
	"context"
	"encoding/base64"
	googleGithub "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"net/http"
	"os"
)

type github struct {
	client *googleGithub.Client
}

func NewGithub(token string) *github {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.TODO(), ts)
	return &github{client: googleGithub.NewClient(tc)}
}

func (g *github) GetFileContents(owner, repo, branch, file string) (string, error) {
	getContentsOptions := &googleGithub.RepositoryContentGetOptions{Ref: branch}
	fileContent, _, _, err := g.client.Repositories.GetContents(
		context.TODO(), owner, repo, file, getContentsOptions)
	if err != nil {
		return "", err
	}
	decodedContentInBytes, errDecoding := base64.StdEncoding.DecodeString(*fileContent.Content)
	return string(decodedContentInBytes), errDecoding
}

func (g *github) ParseWebhook(req *http.Request) (interface{}, error) {
	payload, err := googleGithub.ValidatePayload(req, []byte(os.Getenv(EnvVarGithubWebhookSecret)))
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	return googleGithub.ParseWebHook(googleGithub.WebHookType(req), payload)
}

func (g *github) PushEventToPushEventMeta(pushEvent interface{}) *PushEventMeta {
	pE, ok := pushEvent.(*googleGithub.PushEvent)
	if !ok {
		return nil
	}

	pEMeta := &PushEventMeta{
		Ref:     *pE.Ref,
		Before:  *pE.Before,
		After:   *pE.After,
		Repo:    *pE.Repo.Name,
		Owner:   *pE.Repo.Owner.Name,
		Commits: make([]Commit, len(pE.Commits)),
	}
	for i := 0; i <= len(pE.Commits)-1; i++ {
		pEMeta.Commits[i].Added = pE.Commits[i].Added
		pEMeta.Commits[i].Removed = pE.Commits[i].Removed
		pEMeta.Commits[i].Modified = pE.Commits[i].Modified
		pEMeta.Commits[i].SHA = pE.Commits[i].GetSHA()
	}

	return pEMeta
}

func (g *github) GetDeployDir() string {
	return os.Getenv(EnvVarGithubReleaseFilesDir)
}
