package fctl

import (
	"fmt"
	"os/exec"
)

func Exec(words ...string) (string, error) {
	if len(words) == 0 {
		return "", fmt.Errorf("command is empty")
	}
	cmd := exec.Command(words[0], words[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command %s failed: %s", cmd, string(output))
	}
	return string(output), nil
}

func Run(words ...string) error {
	_, err := Exec(words...)
	return err
}
