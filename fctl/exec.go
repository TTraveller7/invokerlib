package main

import (
	"fmt"
	"os/exec"
)

func Exec(words ...string) (string, error) {
	if len(words) == 0 {
		return "", fmt.Errorf("command is empty")
	}
	cmd := exec.Command(words[0], words[1:]...)
	logs.Printf("%s", cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: output=%s, err=%v", string(output), err)
	}
	return string(output), nil
}

func ExecAtDir(dir string, words ...string) (string, error) {
	if len(words) == 0 {
		return "", fmt.Errorf("command is empty")
	}
	cmd := exec.Command(words[0], words[1:]...)
	logs.Printf("%s", cmd)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed at dir %s: output=%s, err=%v", dir, string(output), err)
	}
	return string(output), nil
}

func Run(words ...string) error {
	_, err := Exec(words...)
	return err
}

// FctlRun runs commands under FCTLHOME directory. See consts.go
func FctlRun(words ...string) error {
	_, err := ExecAtDir(FctlHome, words...)
	return err
}
