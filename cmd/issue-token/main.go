package main

import (
	"fmt"
	"log"

	"github.com/parkiroid/parkiroid-server/internal/auth"
)

func main() {
	token, err := auth.GenerateEmbeddedAPIToken()
	if err != nil {
		log.Fatalf("failed to generate token: %v", err)
	}

	fmt.Println(token)
	fmt.Println()
	fmt.Println("Set this value on the server:")
	fmt.Printf("  PARKIROID_EMBEDDED_API_TOKEN=%s\n", token)
	fmt.Println()
	fmt.Println("Embed the same value in your app and send it as:")
	fmt.Printf("  Authorization: Bearer %s\n", token)
}
