package main

import (
	"bufio"
	"dazedtrader/api"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func main() {
	fmt.Println("=== ROBINHOOD API DEBUG MODE ===")
	fmt.Println("This will show all API requests/responses without TUI interference")
	fmt.Println()

	// Create API client
	client := api.NewClient()

	// Get credentials
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Enter password: ")
	passwordBytes, _ := term.ReadPassword(int(syscall.Stdin))
	password := string(passwordBytes)
	fmt.Println() // New line after password input

	fmt.Print("Enter MFA code (if required, or press enter to skip): ")
	mfaCode, _ := reader.ReadString('\n')
	mfaCode = strings.TrimSpace(mfaCode)

	fmt.Println()
	fmt.Println("=== ATTEMPTING LOGIN ===")
	fmt.Println()

	// Try to login with full debug output
	resp, err := client.Login(username, password, mfaCode)
	if err != nil {
		fmt.Printf("LOGIN FAILED: %v\n", err)
		return
	}

	fmt.Printf("LOGIN SUCCESS!\n")
	fmt.Printf("Token: %s\n", resp.Token)
	fmt.Printf("Access Token: %s\n", resp.AccessToken)
	fmt.Printf("MFA Required: %v\n", resp.MFARequired)
	fmt.Printf("MFA Type: %s\n", resp.MFAType)
}