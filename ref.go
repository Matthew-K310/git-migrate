package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Config struct {
	SourceType      string // "github", "gitlab", "gitea", "forgejo"
	SourceDomain    string
	SourceUsername  string
	SourceToken     string
	
	TargetType      string // "gitea", "forgejo", "gitlab"
	TargetDomain    string
	TargetUsername  string
	TargetToken     string
	TargetRepoOwner string
	
	MakePrivate     bool
	EnableWiki      bool
	EnableMirror    bool
}

type Repository struct {
	Name    string
	CloneURL string
	SSHURL   string
}

// ForgeClient interface for different git forges
type ForgeClient interface {
	FetchRepos(config Config) ([]Repository, error)
	MigrateRepo(config Config, repo Repository) error
}

func main() {
	// Configuration from environment variables or config file
	config := Config{
		SourceType:      getEnv("SOURCE_TYPE", "github"),
		SourceDomain:    getEnv("SOURCE_DOMAIN", "github.com"),
		SourceUsername:  getEnv("SOURCE_USERNAME", ""),
		SourceToken:     getEnv("SOURCE_TOKEN", ""),
		
		TargetType:      getEnv("TARGET_TYPE", "gitea"),
		TargetDomain:    getEnv("TARGET_DOMAIN", ""),
		TargetUsername:  getEnv("TARGET_USERNAME", ""),
		TargetToken:     getEnv("TARGET_TOKEN", ""),
		TargetRepoOwner: getEnv("TARGET_REPO_OWNER", ""),
		
		MakePrivate:     getEnv("MAKE_PRIVATE", "true") == "true",
		EnableWiki:      getEnv("ENABLE_WIKI", "true") == "true",
		EnableMirror:    getEnv("ENABLE_MIRROR", "false") == "true",
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Get source forge client
	sourceClient := getForgeClient(config.SourceType)
	if sourceClient == nil {
		log.Fatalf("Unsupported source forge type: %s", config.SourceType)
	}

	// Get target forge client
	targetClient := getForgeClient(config.TargetType)
	if targetClient == nil {
		log.Fatalf("Unsupported target forge type: %s", config.TargetType)
	}

	// Fetch repositories from source
	repos, err := sourceClient.FetchRepos(config)
	if err != nil {
		log.Fatalf("Failed to fetch repos from %s: %v", config.SourceType, err)
	}

	fmt.Printf("Found %d repositories on %s\n", len(repos), config.SourceDomain)

	// Migrate each repository
	for _, repo := range repos {
		fmt.Printf("Migrating %s...\n", repo.Name)
		if err := targetClient.MigrateRepo(config, repo); err != nil {
			log.Printf("Failed to migrate %s: %v", repo.Name, err)
		} else {
			fmt.Printf("✓ Successfully migrated %s\n", repo.Name)
		}
	}
}

func getForgeClient(forgeType string) ForgeClient {
	switch forgeType {
	case "github":
		return &GitHubClient{}
	case "gitlab":
		return &GitLabClient{}
	case "gitea", "forgejo":
		return &GiteaClient{}
	default:
		return nil
	}
}

// GitHub Client Implementation
type GitHubClient struct{}

type GitHubRepo struct {
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
	CloneURL string `json:"clone_url"`
}

func (c *GitHubClient) FetchRepos(config Config) ([]Repository, error) {
	url := fmt.Sprintf("https://%s/api/v3/users/%s/repos?per_page=200&type=all", 
		config.SourceDomain, config.SourceUsername)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(config.SourceUsername, config.SourceToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
