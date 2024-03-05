package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	rt "tfc-pipeline-run-task/runTask"
)

type CommandResult struct {
	Output   string
	ExitCode int
}

func RunCommands(workspace string, commandsFilePath string, runTaskPayload rt.RunTaskPayload) ([]rt.RunTaskOutcome, error) {

	outcomes := []rt.RunTaskOutcome{}

	// Read the commands from the file
	commands, err := readCommands(commandsFilePath)
	if err != nil {
		return nil, err
	}

	// Run each command
	for index, command := range commands {
		level := "none"
		outcome := rt.RunTaskOutcome{}
		outcome.Type = "task-result-outcomes"
		outcome.Attributes.OutcomeID = fmt.Sprintf("COMMAND-%d", index)
		result, err := runCommand(workspace, command)
		if err != nil {
			log.Println("Error executing command:", err)
			outcome.Attributes.Description = fmt.Sprintf("Error executing command: %s", err)
			level = "error"

			outcome.Attributes.Tags.Status = []rt.RunTaskRichLabel{
				{
					Label: "Status",
					Level: level,
				},
			}

		}
		if result != nil {

			outcome.Attributes.Description = fmt.Sprintf("Command: %s", command)
			outcome.Attributes.Body = result.Output

			if result.ExitCode == 0 {
				level = "info"
			} else if result.ExitCode > 0 {
				level = "warning"
			} else if result.ExitCode < 0 {
				level = "error"
			}

			outcome.Attributes.Tags.Status = []rt.RunTaskRichLabel{
				{
					Label: "Status",
					Level: level,
				},
			}

			// For debugging
			// log.Println("Output:", result.Output)
			// log.Println("Exit Code:", result.ExitCode)
		}
		outcomes = append(outcomes, outcome)

	}
	return outcomes, nil
}

func runCommand(workingDirectory string, command string) (*CommandResult, error) {
	parts := strings.Fields(command)
	head := parts[0]
	parts = parts[1:]

	cmd := exec.Command(head, parts...)
	cmd.Dir = workingDirectory
	output, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, err
		}
	}

	return &CommandResult{
		Output:   string(output),
		ExitCode: exitCode,
	}, nil
}

func readCommands(filePath string) ([]string, error) {
	commands := []string{}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		command := scanner.Text()
		commands = append(commands, command)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commands, nil
}

func FileMatchesSHA256(filePath string, expectedSHA256 string) (bool, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	// Calculate the SHA-256 checksum
	hash := sha256.Sum256(data)
	calculatedSHA256 := hex.EncodeToString(hash[:])

	// Compare the calculated checksum with the expected one
	return calculatedSHA256 == expectedSHA256, nil
}
