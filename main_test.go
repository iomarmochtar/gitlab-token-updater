package main_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	m "github.com/iomarmochtar/gitlab-token-updater"
	"github.com/stretchr/testify/assert"
)

func fixturePath(fname string) string {
	return path.Join("test", "fixtures", fname)
}

func readFixture(fpath string) []byte {
	data, err := os.ReadFile(fixturePath(fpath))
	if err != nil {
		panic(err)
	}
	return data
}

func TestRun(t *testing.T) {
	testCases := map[string]struct {
		cmdArgs        []string
		expectedErrMsg string
		mockGitlabResp func(w http.ResponseWriter, r *http.Request)
	}{
		"ok: rotate group access token then update project CICD var": {
			cmdArgs: []string{"--config", fixturePath("cmd_test_config.yml"), "--force", "--strict"},
			mockGitlabResp: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				method := r.Method
				path := r.URL.Path
				//nolint:gocritic
				if path == `/api/v4/groups//some/group/path/access_tokens` && method == http.MethodGet {
					_, _ = w.Write(readFixture("api_responses/group_access_tokens.json"))
				} else if path == `/api/v4/groups//some/group/path/access_tokens/42/rotate` && method == http.MethodPost {
					_, _ = w.Write(readFixture("api_responses/group_access_token_rotate.json"))
				} else if path == `/api/v4/projects//some/repo/path/variables/THIS_IS_VAR` && method == http.MethodPut {
					_, _ = w.Write(readFixture("api_responses/project_cicd_var.json"))
				}
			},
		},
		"ok: dry run enabled": {
			cmdArgs: []string{"--config", fixturePath("cmd_test_config.yml"), "--dry-run", "--force"},
			mockGitlabResp: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				method := r.Method
				path := r.URL.Path
				if path == `/api/v4/groups//some/group/path/access_tokens` && method == http.MethodGet {
					_, _ = w.Write(readFixture("api_responses/group_access_tokens.json"))
				} else if path == `/api/v4/projects//some/repo/path/variables/THIS_IS_VAR` && method == http.MethodGet {
					_, _ = w.Write(readFixture("api_responses/project_cicd_var.json"))
				}
			},
		},
		"err: not providing required flags": {
			cmdArgs:        []string{},
			expectedErrMsg: `Required flag "config" not set`,
		},
		"err: required config is not set": {
			cmdArgs:        []string{"--config", fixturePath("cmd_test_config.yml"), "--debug"},
			expectedErrMsg: "empty host",
		},
		"err: while read broken configuration fiel": {
			cmdArgs:        []string{"--config", fixturePath("broken_config.yml")},
			expectedErrMsg: `error in unmarshal YAML object: yaml: line 2: mapping values are not allowed in this context`,
		},
	}

	for title, tc := range testCases {
		t.Run(title, func(t *testing.T) {
			if tc.mockGitlabResp != nil {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Log("[http_test] attempt to access. path:", r.URL.Path, ", method: ", r.Method)
					tc.mockGitlabResp(w, r)
				}))
				_ = os.Setenv("HTTP_TEST", ts.URL)

				t.Cleanup(func() {
					ts.Close()
					_ = os.Unsetenv("HTTP_TEST")
				})
			}
			cmdArgs := append([]string{m.CmdName}, tc.cmdArgs...)
			command := m.New()
			command.Writer = io.Discard
			err := command.Run(cmdArgs)

			if tc.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	// get version
	buf := new(bytes.Buffer)
	app := m.New()
	app.Writer = buf
	err := app.Run([]string{"gitlab-token-updater", "--version"})

	obj := struct {
		Version     string `json:"version"`
		Commit      string `json:"commit"`
		CompileTime string `json:"compile_time"`
	}{}
	assert.NoError(t, err)
	_ = json.Unmarshal(buf.Bytes(), &obj)
	assert.Equal(t, m.Version, obj.Version)
	assert.Equal(t, m.BuildHash, obj.Commit)
}
