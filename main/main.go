package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func main() {
	// Read the credentials from the CSV file
	file, err := os.Open("credentials.csv")
	if err != nil {
		log.Fatalf("Couldn't open the credentials file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read the CSV file: %v", err)
	}

	// Check if CSV format is correct
	if len(records) < 1 || len(records[0]) < 2 {
		log.Fatal("CSV file format is incorrect")
	}

	// Read the command from the user
	fmt.Println("Enter the command to execute:")
	scanner := bufio.NewReader(os.Stdin)
	line, _ := scanner.ReadString('\n')
	command := strings.TrimSpace(line)

	// Create or open the results file
	resultsFile, err := os.Create("results.txt")
	if err != nil {
		log.Fatalf("Failed to create results file: %v", err)
	}
	defer resultsFile.Close()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent executions

	for _, record := range records {
		if len(record) < 2 {
			continue // Skip invalid records
		}

		accessKeyID := record[0]
		secretAccessKey := record[1]

		// Configure environment variables

		wg.Add(1)
		go func(accessKeyID, secretAccessKey string) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire a slot

			// Prepare the command
			cmd := exec.Command("bash", "-c", command)
			cmd.Env = append(os.Environ(),
				"AWS_ACCESS_KEY_ID="+accessKeyID,
				"AWS_SECRET_ACCESS_KEY="+secretAccessKey,
			)
			output, err := cmd.CombinedOutput()

			// Aggregate results
			result := fmt.Sprintf("AccessKeyID: %s\nSecretAccessKey: %s\nCommand: %s\nOutput:\n%s\n\n", accessKeyID, secretAccessKey, command, output)
			if err != nil {
				result += fmt.Sprintf("Error: %v\n", err)
			}

			// Write results to file
			_, err = resultsFile.WriteString(result)
			if err != nil {
				log.Printf("Failed to write to results file: %v", err)
			}

			<-semaphore // Release the slot
		}(accessKeyID, secretAccessKey)
	}

	wg.Wait()
	fmt.Println("All commands have been executed and results are aggregated.")
}
func openTerminal(command string, index int) error {
	// Adjust this based on your OS and terminal emulator
	switch {
	case isMacOS():
		return openMacOSTerminal(command, index)
	case isLinux():
		return openLinuxTerminal(command, index)
	default:
		return fmt.Errorf("unsupported operating system")
	}
}

func isMacOS() bool {
	return strings.HasPrefix(os.Getenv("GOOS"), "darwin")
}

func isLinux() bool {
	return strings.HasPrefix(os.Getenv("GOOS"), "linux")
}

func openMacOSTerminal(command string, index int) error {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Terminal" to do script "%s"`, command))
	return cmd.Run()
}

func openLinuxTerminal(command string, index int) error {
	// Using `gnome-terminal` here; replace with your preferred terminal emulator
	cmd := exec.Command("gnome-terminal", "--", "bash", "-c", command)
	return cmd.Run()
}
