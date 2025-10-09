package util

import (
	"fmt"
	"os"
	"os/exec"
	"prostic/internal/config"
)

func RunResticCommand(showOutput bool, args ...string) error {
	env := os.Environ()
	for key, val := range config.Get().Restic.EnvVars {
		env = append(env, key+"="+val)
	}

	cmd := exec.Command("restic", args...)
	cmd.Env = env

	if showOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("restic command failed: %v", err)
		}
		return nil
	}

	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restic command failed: %v", err)
	}

	return nil
}
