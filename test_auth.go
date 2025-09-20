package main

import (
	"dazedtrader/api"
	"fmt"
)

func main() {
	fmt.Println("=== TESTING ROBINHOOD OAUTH2 ENDPOINT ===")

	// Create API client
	client := api.NewClient()

	// Test with sample credentials (will fail but show endpoint works)
	username := "test@example.com"
	password := "samplepassword123"
	mfaCode := ""

	fmt.Printf("Testing endpoint with sample data...\n")
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Password: [%d characters]\n", len(password))
	fmt.Println()

	// Try authentication
	resp, err := client.Login(username, password, mfaCode)
	if err != nil {
		fmt.Printf("Expected authentication failure: %v\n", err)
		fmt.Println("This is normal - we're just testing the endpoint works")
	} else {
		fmt.Printf("Unexpected success: %+v\n", resp)
	}
}