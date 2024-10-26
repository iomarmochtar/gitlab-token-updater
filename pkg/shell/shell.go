// Package shell all related to shell execution, included with it's interface
package shell

import (
	"fmt"
	"os"
	"os/exec"
)

//go:generate mockgen -destination ../../test/mocks/shell/shell.go -source=shell.go

// Shell abstracting shell command execution
type Shell interface {
	Exec(command string, envVars map[string]string) (results []byte, err error)
	FileMustExists(path string) error
}

// SHExecutor implements shell interface
type SHExecutor struct{}

// Exec executing shell command, also the passing environment variables to the executor
func (s SHExecutor) Exec(command string, envVars map[string]string) (results []byte, err error) {
	cmd := exec.Command(command)
	for k, v := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	output, err := cmd.Output()

	if err != nil {
		return nil, err
	}

	return output, nil
}

// FileMustExists check whether the file is exists or not
func (s SHExecutor) FileMustExists(filePath string) error {
	_, err := os.Stat(filePath)
	return err
}
