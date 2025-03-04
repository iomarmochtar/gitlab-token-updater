// Package gitlab abstracting Gitlab API execution
package gitlab

import (
	"time"

	gl "github.com/xanzy/go-gitlab"
)

//go:generate mockgen -destination ../../test/mocks/gitlab/gitlab.go -source=gitlab.go

type GitlabTargetType int

const (
	GitlabTargetTypeRepo GitlabTargetType = iota
	GitlabTargetTypeGroup
	GitlabTargetTypePersonal
)

var (
	includeRevoked = false
	fetchPerPage   = 20
)

// GitlabCICDVar CICD variable
type GitlabCICDVar struct {
	Key   string
	Value string
	Type  GitlabTargetType
}

// GitlabAccessToken instance in joining repo and group access token
type GitlabAccessToken struct {
	ID        int
	Name      string
	Active    bool
	Revoked   bool
	ExpiresAt *time.Time
	Type      GitlabTargetType
	Path      string
}

// GitlabAPI spec for the used Gitlab API
type GitlabAPI interface {
	Auth(token string) error
	InitGitlab(baseURL, token string) (GitlabAPI, error)
	GetRepoVar(path string, varName string) (*GitlabCICDVar, error)
	GetGroupVar(path string, varName string) (*GitlabCICDVar, error)
	UpdateGroupVar(path string, varName string, value string) error
	UpdateRepoVar(path string, varName string, value string) error
	RotatePersonalToken(tokenID int, expiredAt time.Time) (string, error)
	RotateRepoToken(path string, tokenID int, expiredAt time.Time) (string, error)
	RotateGroupToken(path string, tokenID int, expiredAt time.Time) (string, error)
	ListPersonalAccessToken() ([]GitlabAccessToken, error)
	ListRepoAccessToken(path string) ([]GitlabAccessToken, error)
	ListGroupAccessToken(path string) ([]GitlabAccessToken, error)
}

// Gitlab implement GitlabAPI interface
type Gitlab struct {
	baseURL string
	client  *gl.Client
}

// Auth initiate gitlab API client
func (g *Gitlab) Auth(token string) error {
	client, err := gl.NewClient(token, gl.WithBaseURL(g.baseURL))
	if err != nil {
		return err
	}
	g.client = client
	return nil
}

// GetRepoVar get repo/project CICD var
func (g Gitlab) GetRepoVar(path string, varName string) (*GitlabCICDVar, error) {
	cicdVar, _, err := g.client.ProjectVariables.GetVariable(path, varName, nil)
	if err != nil {
		return nil, err
	}

	return &GitlabCICDVar{
		Key:   cicdVar.Key,
		Value: cicdVar.Value,
		Type:  GitlabTargetTypeRepo,
	}, nil
}

// GetGroupVar get group CICD var
func (g Gitlab) GetGroupVar(path string, varName string) (*GitlabCICDVar, error) {
	cicdVar, _, err := g.client.GroupVariables.GetVariable(path, varName, nil)
	if err != nil {
		return nil, err
	}

	return &GitlabCICDVar{
		Key:   cicdVar.Key,
		Value: cicdVar.Value,
		Type:  GitlabTargetTypeGroup,
	}, nil
}

// RotatePersonalToken rotate/renew personal access token
func (g Gitlab) RotatePersonalToken(tokenID int, expiredAt time.Time) (string, error) {
	convTime := gl.ISOTime(expiredAt)
	newToken, _, err := g.client.PersonalAccessTokens.RotatePersonalAccessToken(tokenID, &gl.RotatePersonalAccessTokenOptions{
		ExpiresAt: &convTime,
	})

	if err != nil {
		return "", err
	}

	return newToken.Token, nil
}

// RotateRepoToken rotate/renew project access token
func (g Gitlab) RotateRepoToken(path string, tokenID int, expiredAt time.Time) (string, error) {
	convTime := gl.ISOTime(expiredAt)
	newToken, _, err := g.client.ProjectAccessTokens.RotateProjectAccessToken(path, tokenID, &gl.RotateProjectAccessTokenOptions{
		ExpiresAt: &convTime,
	})
	if err != nil {
		return "", err
	}

	return newToken.Token, nil
}

// RotateGroupToken rotate/renew group access token
func (g Gitlab) RotateGroupToken(path string, tokenID int, expiredAt time.Time) (string, error) {
	convTime := gl.ISOTime(expiredAt)
	newToken, _, err := g.client.GroupAccessTokens.RotateGroupAccessToken(path, tokenID, &gl.RotateGroupAccessTokenOptions{
		ExpiresAt: &convTime,
	})
	if err != nil {
		return "", err
	}

	return newToken.Token, nil
}

// UpdateGroupVar update CICD variable in a group
func (g Gitlab) UpdateGroupVar(path string, varName string, value string) error {
	_, _, err := g.client.GroupVariables.UpdateVariable(path, varName, &gl.UpdateGroupVariableOptions{
		Value: &value,
	})
	return err
}

// UpdateRepoVar update CICD variable in a repo/project
func (g Gitlab) UpdateRepoVar(path string, varName string, value string) error {
	_, _, err := g.client.ProjectVariables.UpdateVariable(path, varName, &gl.UpdateProjectVariableOptions{
		Value: &value,
	})
	return err
}

// ListRepoAccessToken get list of repo/project access token
func (g Gitlab) ListRepoAccessToken(path string) (gat []GitlabAccessToken, err error) {
	listOptions := &gl.ListProjectAccessTokensOptions{
		Page:    1,
		PerPage: fetchPerPage,
	}

	for {
		tokens, resp, err := g.client.ProjectAccessTokens.ListProjectAccessTokens(path, listOptions)
		if err != nil {
			return nil, err
		}
		for idx := range tokens {
			tk := tokens[idx]
			// ignore the revoked one
			if tk.Revoked && !includeRevoked {
				continue
			}
			gat = append(gat, GitlabAccessToken{
				ID:        tk.ID,
				Name:      tk.Name,
				Active:    tk.Active,
				Revoked:   tk.Revoked,
				ExpiresAt: (*time.Time)(tk.ExpiresAt),
				Type:      GitlabTargetTypeRepo,
				Path:      path,
			})
		}

		// Check if there are more pages to fetch
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}

	return gat, nil
}

// ListGroupAccessToken get list of group access token
func (g Gitlab) ListGroupAccessToken(path string) (gat []GitlabAccessToken, err error) {
	listOptions := &gl.ListGroupAccessTokensOptions{
		Page:    1,
		PerPage: fetchPerPage,
	}

	for {
		tokens, resp, err := g.client.GroupAccessTokens.ListGroupAccessTokens(path, listOptions)
		if err != nil {
			return nil, err
		}

		for idx := range tokens {
			tk := tokens[idx]
			// ignore the revoked one
			if tk.Revoked && !includeRevoked {
				continue
			}
			gat = append(gat, GitlabAccessToken{
				ID:        tk.ID,
				Name:      tk.Name,
				Active:    tk.Active,
				Revoked:   tk.Revoked,
				ExpiresAt: (*time.Time)(tk.ExpiresAt),
				Type:      GitlabTargetTypeGroup,
				Path:      path,
			})
		}

		// Check if there are more pages to fetch
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}

	return gat, nil
}

// ListPersonalAccessToken get list of group access token
func (g Gitlab) ListPersonalAccessToken() (pat []GitlabAccessToken, err error) {
	listOptions := &gl.ListPersonalAccessTokensOptions{
		ListOptions: gl.ListOptions{
			Page:    1,
			PerPage: fetchPerPage,
		},
		Revoked: gl.Ptr(includeRevoked),
	}

	for {
		tokens, resp, err := g.client.PersonalAccessTokens.ListPersonalAccessTokens(listOptions, nil)

		if err != nil {
			return nil, err
		}

		for idx := range tokens {
			tk := tokens[idx]
			// ignore the revoked one
			if tk.Revoked && !includeRevoked {
				continue
			}
			pat = append(pat, GitlabAccessToken{
				ID:        tk.ID,
				Name:      tk.Name,
				Active:    tk.Active,
				Revoked:   tk.Revoked,
				Path:      "@personal",
				ExpiresAt: (*time.Time)(tk.ExpiresAt),
				Type:      GitlabTargetTypePersonal,
			})
		}

		// Check if there are more pages to fetch
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}

	return pat, nil
}

// InitGitlab initiating external/another Gitlab instance
func (g *Gitlab) InitGitlab(baseURL, token string) (GitlabAPI, error) {
	return NewGitlabAPI(baseURL, token)
}

// NewGitlabAPI returning gitlab API object
func NewGitlabAPI(baseURL, token string) (GitlabAPI, error) {
	gl := &Gitlab{baseURL: baseURL}
	if err := gl.Auth(token); err != nil {
		return nil, err
	}

	return gl, nil
}
