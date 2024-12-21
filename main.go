package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/likexian/whois"
	"github.com/fatih/color"
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
	repo := flag.String("r", "", "Specific repository to process (leave empty to process all repositories)")
	flag.Parse()

	// Validate inputs
	if *username == "" || *token == "" {
		log.Fatalf("Usage: gemails -u <username> -t <token> -o <output file> [-r <repo>]")
	}

	// Track unique emails using a map
	uniqueEmails := make(map[string]bool)
	uniqueDomains := make(map[string]bool)

	var repos []Repository
	if *repo != "" {
		// Process only the specific repository
		repos = append(repos, Repository{Name: *repo})
	} else {
		// Fetch all repositories
		repos = fetchRepos(*username, *token)
	}

	// Process each repository
	for _, repo := range repos {
		fmt.Printf("Processing repository: %s\n", repo.Name)
		// Fetch commits for each repository
		commits := fetchCommits(*username, repo.Name, *token)
		for _, commit := range commits {
			email := commit.CommitData.Committer.Email
			if email != "" && !uniqueEmails[email] {
				uniqueEmails[email] = true
				// Extract domain and add it to uniqueDomains map
				domain := extractDomainFromEmail(email)
				if domain != "" {
					uniqueDomains[domain] = true
				}
			}
		}
	}

	// Save unique emails to the specified output file
	saveUniqueEmails(uniqueEmails, *outputFile)
	fmt.Printf("\nUnique emails saved to %s\n", *outputFile)

	// Now, check the domain expiry for each unique domain
	checkDomainsExpiry(uniqueDomains)
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

	// Handle different HTTP status codes, especially 409 Conflict
	if resp.StatusCode == http.StatusConflict { // 409 Conflict
		log.Printf("Warning: 409 Conflict encountered for URL: %s. Skipping.", url)
		return nil // Skip this request and return an empty response
	} else if resp.StatusCode != http.StatusOK {
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

// checkDomainsExpiry checks WHOIS info for each domain and compares expiry date
func checkDomainsExpiry(domains map[string]bool) {
	for domain := range domains {
		// Perform WHOIS lookup
		whoisInfo, err := whois.Whois(domain)
		if err != nil {
			log.Printf("Error fetching WHOIS info for domain %s: %v", domain, err)
			continue
		}

		// Try to find the expiry date in the WHOIS info (simplified)
		expiryDate := extractExpiryDateFromWhois(whoisInfo)
		if expiryDate.IsZero() {
			log.Printf("No expiry date found for domain %s", domain)
			continue
		}

		// Compare the expiry date with today's date
		daysUntilExpiry := time.Until(expiryDate).Hours() / 24
		if daysUntilExpiry < 30 {
			color.Red("Domain %s is nearing expiry (Expires on %s, %d days left)", domain, expiryDate.Format("2006-01-02"), int(daysUntilExpiry))
		} else {
			color.Green("Domain %s has a valid expiry date (Expires on %s, %d days left)", domain, expiryDate.Format("2006-01-02"), int(daysUntilExpiry))
		}
	}
}

// extractDomainFromEmail extracts the domain from an email address
func extractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// extractExpiryDateFromWhois extracts the expiry date from the WHOIS information
func extractExpiryDateFromWhois(whoisInfo string) time.Time {
	// Simple regex pattern to match expiry date (in ISO 8601 format or similar)
	expiryRegex := regexp.MustCompile(`(?i)(?:(expiration|expire|expiry)[^\w]*(date|time)[^\w]*[:\s]+)(\d{4}-\d{2}-\d{2})`)
	matches := expiryRegex.FindStringSubmatch(whoisInfo)

	if len(matches) > 3 {
		expiryDateStr := matches[3]
		expiryDate, err := time.Parse("2006-01-02", expiryDateStr)
		if err != nil {
			log.Printf("Error parsing expiry date: %v", err)
			return time.Time{}
		}
		return expiryDate
	}

	return time.Time{} // return zero value if no expiry date is found
}
