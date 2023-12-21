package fctl

import (
	"fmt"
	"os/exec"
)

func Exec(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command %s failed: %s", cmd, string(output))
	}
	return string(output), nil
}

func Run(name string, args ...string) error {
	_, err := Exec(name, args...)
	return err
}
