package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const githubAPI = "https://api.github.com"

// Repository represents a GitHub repository
type Repository struct {
	Name string `json:"name"`
}

// Commit represents a GitHub commit
type Commit struct {
	CommitData struct {
		Committer struct {
			Email string `json:"email"`
		} `json:"committer"`
	} `json:"commit"`
}

func main() {
	// Define and parse command-line flags
	username := flag.String("u", "", "GitHub username or organization")
	token := flag.String("t", "", "GitHub API token")
	outputFile := flag.String("o", "emails.txt", "Output file to save unique emails")
	flag.Parse()

	// Validate inputs
	if *username == "" || *token == "" {
		log.Fatalf("Usage: ghpemails -u <username> -t <token> -o <output file>")
	}

	// Track unique emails using a map
	uniqueEmails := make(map[string]bool)

	// Fetch repositories
	repos := fetchRepos(*username, *token)
	for _, repo := range repos {
		fmt.Printf("Processing repository: %s\n", repo.Name)
		// Fetch commits for each repository
		commits := fetchCommits(*username, repo.Name, *token)
		for _, commit := range commits {
			email := commit.CommitData.Committer.Email
			if email != "" && !uniqueEmails[email] {
				uniqueEmails[email] = true
				fmt.Printf("Unique Committer Email: %s\n", email)
			}
		}
	}

	// Save unique emails to the specified output file
	saveUniqueEmails(uniqueEmails, *outputFile)
	fmt.Printf("\nUnique emails saved to %s\n", *outputFile)
}

// fetchRepos fetches all repositories for a user or organization
func fetchRepos(userOrOrg, token string) []Repository {
	url := fmt.Sprintf("%s/users/%s/repos", githubAPI, userOrOrg)
	response := sendRequest(url, token)

	var repos []Repository
	if err := json.Unmarshal(response, &repos); err != nil {
		log.Fatalf("Error unmarshaling repositories: %v", err)
	}
	return repos
}

// fetchCommits fetches all commits for a given repository
func fetchCommits(userOrOrg, repo, token string) []Commit {
	url := fmt.Sprintf("%s/repos/%s/%s/commits", githubAPI, userOrOrg, repo)
	response := sendRequest(url, token)

	var commits []Commit
	if err := json.Unmarshal(response, &commits); err != nil {
		log.Printf("Error unmarshaling commits for repo %s: %v", repo, err)
	}
	return commits
}

// sendRequest sends an HTTP GET request to the provided URL with the GitHub token
func sendRequest(url, token string) []byte {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("GitHub API returned status code %d for URL %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	return body
}

// saveUniqueEmails saves unique emails to a specified file
func saveUniqueEmails(emails map[string]bool, outputFile string) {
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer file.Close()

	for email := range emails {
		if _, err := file.WriteString(email + "\n"); err != nil {
			log.Fatalf("Error writing to output file: %v", err)
		}
	}
}
