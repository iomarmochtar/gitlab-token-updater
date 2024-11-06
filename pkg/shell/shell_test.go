package shell_test

import (
	"os"
	"testing"

	s "github.com/iomarmochtar/gitlab-token-updater/pkg/shell"
	t_helper "github.com/iomarmochtar/gitlab-token-updater/test"
	"github.com/stretchr/testify/assert"
)

func TestSHExecutor_Exec(t *testing.T) {
	sh := s.SHExecutor{}
	envs := map[string]string{
		"GL_NEW_TOKEN": "glpat-abc",
		"ANOTHER_ENV":  "abc",
	}
	result, err := sh.Exec(t_helper.FixturePath("sample_script.sh"), envs)
	assert.NoError(t, err)
	assert.Equal(t, "new token: glpat-abc, another env: abc", string(result))

	result, err = sh.Exec("/file/is/not/found", envs)
	assert.Nil(t, result)
	assert.True(t, os.IsNotExist(err))
}

func TestSHExecutor_FileMustExists(t *testing.T) {
	sh := s.SHExecutor{}
	err := sh.FileMustExists(t_helper.FixturePath("sample_script.sh"))
	assert.NoError(t, err)

	err = sh.FileMustExists("/file/is/not/found")
	assert.True(t, os.IsNotExist(err))
}
