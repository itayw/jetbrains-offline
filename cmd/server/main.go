package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Serve the static files from the "output" directory
	pluginDir := "output"

	// Use FileServer to serve the files
	http.Handle("/", http.FileServer(http.Dir(pluginDir)))

	// Start the server on port 8080
	fmt.Println("Starting plugin server on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
