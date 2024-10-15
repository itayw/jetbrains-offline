package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"jetbrains-offline/internal/downloader"
	"jetbrains-offline/internal/models"
	"os"
)

func main() {
	// Define the path to the config file
	configPath := "config/config.json"

	// Read the config file
	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Call the SyncPlugins function with the config object
	err = downloader.SyncPlugins(config)
	if err != nil {
		fmt.Printf("Failed to sync plugins: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Sync completed successfully!")
}

// loadConfig reads the config.json and returns a models.Config object
func loadConfig(filePath string) (models.Config, error) {
	var config models.Config

	// Read the config file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse the JSON into a models.Config struct
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config: %v", err)
	}

	return config, nil
}
