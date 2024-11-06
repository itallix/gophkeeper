package cmd

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func promptString(cmd *cobra.Command, reader *bufio.Reader, prompt string) (string, error) {
	cmd.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func promptNumber(cmd *cobra.Command, reader *bufio.Reader, prompt string) (int, error) {
	input, err := promptString(cmd, reader, prompt)
	if err != nil {
		return 0, err
	}
	num, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return 0, err
	}
	return num, nil
}

// promptPassword reads a password securely from the terminal or from cmd.InOrStdin().
func promptPassword(cmd *cobra.Command, reader *bufio.Reader, prompt string) (string, error) {
	cmd.Print(prompt)
	// Check if cmd.InOrStdin() is a terminal
	var fd int
	if f, ok := cmd.InOrStdin().(*os.File); ok {
		fd = int(f.Fd())
	} else {
		// cmd.InOrStdin() is not a file, fallback to non-terminal input
		password, err := reader.ReadString('\n')
		if err != nil {
			cmd.PrintErrf("Error reading input: %v\n", err)
			return "", err
		}
		return strings.TrimSpace(password), nil
	}

	if term.IsTerminal(fd) {
		bytePassword, err := term.ReadPassword(fd)
		cmd.Println() // Move to the next line after password input
		if err != nil {
			cmd.PrintErrf("Error reading password: %v\n", err)
			return "", err
		}
		return string(bytePassword), nil
	}
	// Not a terminal, read from cmd.InOrStdin()
	password, err := reader.ReadString('\n')
	if err != nil {
		cmd.PrintErrf("Error reading input: %v\n", err)
		return "", err
	}
	return strings.TrimSpace(password), nil
}
