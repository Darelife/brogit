package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"brogit/pkg/api"
)

const ServerURL = "http://localhost:8080"

func main() {
	pushCmd := flag.NewFlagSet("push", flag.ExitOnError)
	pushUser := pushCmd.String("user", "default_user", "User making the change")
	pushFile := pushCmd.String("file", "", "File to push")

	commitCmd := flag.NewFlagSet("commit", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("expected 'push' or 'commit' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "push":
		pushCmd.Parse(os.Args[2:])
		if *pushFile == "" {
			fmt.Println("Please provide a file to push using -file")
			os.Exit(1)
		}
		handlePush(*pushUser, *pushFile)
	case "commit":
		commitCmd.Parse(os.Args[2:])
		handleCommit()
	default:
		fmt.Println("expected 'push' or 'commit' subcommands")
		os.Exit(1)
	}
}

func handlePush(user, filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	entry := api.DiffEntry{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		UserID:    user,
		Timestamp: time.Now(),
		Type:      api.ChangeTypeFileEdit,
		FilePath:  filePath,
		Content:   string(content),
	}

	reqBody := api.PushRequest{Entry: entry}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(ServerURL+"/api/push", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to push to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Server returned error: %s", resp.Status)
	}

	fmt.Println("Successfully pushed change for", filePath)
}

func handleCommit() {
	resp, err := http.Post(ServerURL+"/api/commit", "application/json", nil)
	if err != nil {
		log.Fatalf("Failed to commit: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Commit triggered. Server pending entries:\n%s\n", string(body))
}
