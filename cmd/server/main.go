package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("bulhufas server starting...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8420"
	}

	fmt.Printf("listening on :%s\n", port)
}
