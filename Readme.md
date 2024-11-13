**GEmails** is a Go-based CLI tool for retrieving unique committer email addresses from all repositories of a specified GitHub user or organization. The tool uses the GitHub API to fetch repositories and commits and outputs unique email addresses to both stdout and a file.

## Features

- Fetches all repositories for a GitHub user or organization.
- Retrieves all commits for each repository.
- Extracts and outputs unique committer email addresses.
- Saves unique emails to a specified output file.

## Installation

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/mux0x/gemails.git
   cd gemails
   ```
2. Build and Install:

3. Ensure Go is installed and properly configured, then run:
```
    go install github.com/mux0x/gemails@latest
```
    This installs the gemails binary in your Go bin directory.

Usage
```
gemails -u <username> -t <token> -o <output_file>
```
Options

    -u: GitHub username or organization (required).
    -t: GitHub API token (required).
    -o: Output file to save unique emails (optional, defaults to unique_emails.txt).

Example
```
gemails -u octocat -t ghp_12345abcde67890fghijk -o emails.txt
```
This will:

    Retrieve all repositories for the user octocat.
    Fetch all commits for each repository.
    Extract unique committer email addresses.
    Print unique emails to stdout.
    Save unique emails to emails.txt.

Generating a GitHub Token

To use the GitHub API, you need a personal access token:

    Go to your GitHub Developer Settings.
    Generate a new token with the repo scope (if you want to access private repositories).
    Copy the token and pass it to the -t flag.

Contributing

Contributions are welcome! Feel free to open issues or submit pull requests to improve the tool.
License

This project is licensed under the MIT License. See the LICENSE file for details.
Author

mux0x
