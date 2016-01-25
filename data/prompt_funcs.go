package data

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"strings"
	"syscall"
)

func hiddenPrompt(str string) (string, error) {
	fmt.Print(str)
	byts, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("Failed obtaining hidden prompt: %v", err)
	}
	return strings.TrimSpace(string(byts)), nil
}
