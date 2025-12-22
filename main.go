package main


import (
	"log"
	
	"github.com/charmbracelet/huh"
	"github.com/joho/godotenv"
) 

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter your username...").
				Prompt("Username").
				// Password(true).
				// Value(&userName).
				Run(),
			huh.NewInput().
				Title("Enter your access token...").
				Prompt("API Token:").
				// Password(true).
				EchoMode(huh.EchoModePassword).
				// Value(&apiToken).
				Run(),
		),
	)
	
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
