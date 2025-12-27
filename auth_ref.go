package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/joho/godotenv"
)

type Forge struct {
	SourceType     string
	SourceUsername string
	SourceToken    string
	SourceDomain   string

	TargetType     string
	TargetUsername string
	TargetToken    string
	TargetDomain   string
	MakePrivate    bool
	EnableMirror   bool
}

// Simplified version using a map
func SourceAuth(forge *Forge) error {
	// Define all fields to collect
	fields := []struct {
		title  string
		prompt string
		envKey string
		value  *string
	}{
		{"Enter your Git forge type", "Forge type (github/gitlab/gitea/forgejo)...", "SRC_TYPE", &forge.SourceType},
		{"Enter your forge username", "Forge username...", "SRC_USER", &forge.SourceUsername},
		{"Enter your forge domain", "Forge domain (leave empty for default)...", "SRC_DOMAIN", &forge.SourceDomain},
		{"Enter your forge token", "Forge token...", "SRC_TOKEN", &forge.SourceToken},
	}

	// Collect all inputs
	envVars := make(map[string]string)
	for _, field := range fields {
		if err := huh.NewInput().
			Title(field.title).
			Prompt(field.prompt).
			Value(field.value).
			Run(); err != nil {
			return fmt.Errorf("input failed: %w", err)
		}
		envVars[field.envKey] = *field.value
	}

	// Write all to .env at once
	return writeEnvFile(envVars)
}

// Even more simplified using huh.Form
func SourceAuthWithForm(forge *Forge) error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter your Git forge type").
				Prompt("Forge type (github/gitlab/gitea/forgejo)...").
				Value(&forge.SourceType),

			huh.NewInput().
				Title("Enter your forge username").
				Prompt("Forge username...").
				Value(&forge.SourceUsername),

			huh.NewInput().
				Title("Enter your forge domain").
				Prompt("Forge domain (leave empty for default)...").
				Value(&forge.SourceDomain),

			huh.NewInput().
				Title("Enter your forge token").
				Prompt("Forge token...").
				EchoMode(huh.EchoModePassword).
				Value(&forge.SourceToken),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("form failed: %w", err)
	}

	// Write to .env
	envVars := map[string]string{
		"SRC_TYPE":   forge.SourceType,
		"SRC_USER":   forge.SourceUsername,
		"SRC_DOMAIN": forge.SourceDomain,
		"SRC_TOKEN":  forge.SourceToken,
	}

	return writeEnvFile(envVars)
}

// Helper function to write/append to .env file
func writeEnvFile(vars map[string]string) error {
	// Load existing .env if it exists
	existingVars := make(map[string]string)
	if _, err := os.Stat(".env"); err == nil {
		existingVars, _ = godotenv.Read(".env")
	}

	// Merge new vars with existing
	for k, v := range vars {
		existingVars[k] = v
	}

	// Open file for writing (truncate)
	file, err := os.OpenFile(".env", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open .env file: %w", err)
	}
	defer file.Close()

	// Write all vars
	for k, v := range existingVars {
		if _, err := fmt.Fprintf(file, "%s=%s\n", k, v); err != nil {
			return fmt.Errorf("failed to write to .env: %w", err)
		}
	}

	fmt.Println("✓ Saved configuration to .env")
	return nil
}

// Complete source and target auth in one form
func CompleteAuth(forge *Forge) error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Source Forge Configuration").
				Description("Configure the source git forge to migrate FROM"),

			huh.NewSelect[string]().
				Title("Source forge type").
				Options(
					huh.NewOption("GitHub", "github"),
					huh.NewOption("GitLab", "gitlab"),
					huh.NewOption("Gitea", "gitea"),
					huh.NewOption("Forgejo", "forgejo"),
				).
				Value(&forge.SourceType),

			huh.NewInput().
				Title("Source username").
				Value(&forge.SourceUsername),

			huh.NewInput().
				Title("Source domain (optional)").
				Placeholder("github.com").
				Value(&forge.SourceDomain),

			huh.NewInput().
				Title("Source token").
				EchoMode(huh.EchoModePassword).
				Value(&forge.SourceToken),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Target Forge Configuration").
				Description("Configure the target git forge to migrate TO"),

			huh.NewSelect[string]().
				Title("Target forge type").
				Options(
					huh.NewOption("Gitea", "gitea"),
					huh.NewOption("Forgejo", "forgejo"),
					huh.NewOption("GitLab", "gitlab"),
				).
				Value(&forge.TargetType),

			huh.NewInput().
				Title("Target username").
				Value(&forge.TargetUsername),

			huh.NewInput().
				Title("Target domain").
				Placeholder("git.example.com").
				Value(&forge.TargetDomain),

			huh.NewInput().
				Title("Target token").
				EchoMode(huh.EchoModePassword).
				Value(&forge.TargetToken),

			huh.NewConfirm().
				Title("Make repositories private?").
				Value(&forge.MakePrivate),

			huh.NewConfirm().
				Title("Enable mirroring?").
				Value(&forge.EnableMirror),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("form failed: %w", err)
	}

	// Write all config to .env
	envVars := map[string]string{
		"SRC_TYPE":      forge.SourceType,
		"SRC_USER":      forge.SourceUsername,
		"SRC_DOMAIN":    forge.SourceDomain,
		"SRC_TOKEN":     forge.SourceToken,
		"TARGET_TYPE":   forge.TargetType,
		"TARGET_USER":   forge.TargetUsername,
		"TARGET_DOMAIN": forge.TargetDomain,
		"TARGET_TOKEN":  forge.TargetToken,
		"MAKE_PRIVATE":  fmt.Sprintf("%t", forge.MakePrivate),
		"ENABLE_MIRROR": fmt.Sprintf("%t", forge.EnableMirror),
	}

	return writeEnvFile(envVars)
}

func main() {
	forge := &Forge{}

	// Use the complete auth form
	if err := CompleteAuth(forge); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("\n✓ Configuration complete!")
	fmt.Printf("Source: %s (%s)\n", forge.SourceType, forge.SourceUsername)
	fmt.Printf("Target: %s (%s)\n", forge.TargetType, forge.TargetUsername)
}
