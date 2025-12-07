package main

import (
	"fmt"

	"filippo.io/age"
)

func main() {
	// Generate a new age identity
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		fmt.Printf("Error generating identity: %v\n", err)
		return
	}

	fmt.Println("Age Vault - A simple age encryption example")
	fmt.Printf("Generated identity public key: %s\n", identity.Recipient())
}
