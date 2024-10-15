package downloader

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"jetbrains-offline/internal/logger"
	"jetbrains-offline/internal/models"
	xmlgenerator "jetbrains-offline/internal/xml"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// IdeaVersion represents the <idea-version> tag with since-build and until-build attributes
type IdeaVersion struct {
	SinceBuild string `xml:"since-build,attr"`
	UntilBuild string `xml:"until-build,attr"`
}

// PluginVersion represents the plugin version data returned by the XML API
type PluginVersion struct {
	Version     string      `xml:"version"`
	URL         string      `xml:"url"`
	IdeaVersion IdeaVersion `xml:"idea-version"`
	Description string      `xml:"description"`
}

// Category represents a category containing plugins
type Category struct {
	Plugins []PluginVersion `xml:"idea-plugin"`
}

// PluginRepository represents the root element of the XML response
type PluginRepository struct {
	Categories []Category `xml:"category"`
}

// SyncPlugins is responsible for downloading plugins and generating the index.xml
func SyncPlugins(config models.Config) error {
	// Create the plugin directory where everything will be stored
	outputDir := "output/plugins"
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Download each plugin
	for _, plugin := range config.Plugins {
		err := downloadPlugin(plugin, config)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to download plugin: %s", plugin.ID), err)
		}
	}

	// After downloading, generate the index.xml file
	xmlFilePath := filepath.Join(outputDir, "index.xml")
	err = xmlgenerator.GenerateIndexXML(xmlFilePath)
	if err != nil {
		return fmt.Errorf("failed to generate index.xml: %v", err)
	}

	logger.Info("Successfully generated index.xml")
	return nil
}

// downloadPlugin fetches and downloads the plugins compatible with the specified IntelliJ builds
func downloadPlugin(plugin models.Plugin, config models.Config) error {
	logger.Info(fmt.Sprintf("Syncing Plugin: %s", plugin.ID))

	// Fetch available plugin versions from the IntelliJ Plugin API
	pluginVersions, err := fetchPluginVersions(plugin.ID)
	if err != nil {
		return err
	}

	// Download applicable versions
	for _, version := range pluginVersions {
		if isCompatible(version, config) {
			logger.Info(fmt.Sprintf("Downloading Plugin %s version %s", plugin.ID, version.Version))
			err := downloadFile(plugin.ID, version.Version, version.IdeaVersion.SinceBuild, version.IdeaVersion.UntilBuild, version.Description)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to download plugin %s version %s", plugin.ID, version.Version), err)
			}
		}
	}
	return nil
}

// fetchPluginVersions gets all versions of the plugin using the JetBrains Plugin XML API
func fetchPluginVersions(pluginID string) ([]PluginVersion, error) {
	apiURL := fmt.Sprintf("https://plugins.jetbrains.com/plugins/list?pluginId=%s", pluginID)

	// Log the URL being called for debugging purposes
	logger.Debug(fmt.Sprintf("Fetching plugin versions from URL: %s", apiURL))

	// Perform the HTTP request
	resp, err := http.Get(apiURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to make request to %s: %v", apiURL, err))
		return nil, fmt.Errorf("failed to fetch plugin versions: %v", err)
	}
	defer resp.Body.Close()

	// If the response is not 200 OK, log the error and body
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		logger.Error(fmt.Sprintf("Failed response for %s: %s", apiURL, string(bodyBytes)))
		return nil, errors.New("failed to fetch plugin versions: invalid response")
	}

	// Parse the XML response
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	var repository PluginRepository
	err = xml.Unmarshal(bodyBytes, &repository)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to decode XML from %s: %v", apiURL, err))
		return nil, fmt.Errorf("failed to decode plugin versions: %v", err)
	}

	// Aggregate all plugins from all categories
	var pluginVersions []PluginVersion
	for _, category := range repository.Categories {
		pluginVersions = append(pluginVersions, category.Plugins...)
	}

	return pluginVersions, nil
}

func isCompatible(version PluginVersion, config models.Config) bool {
	buildRanges := extractBuildRange(config)

	// Debugging: Log the build ranges and plugin version's since-build/until-build
	logger.Debug(fmt.Sprintf("Checking compatibility for plugin version %s", version.Version))
	logger.Debug(fmt.Sprintf("Plugin since-build: %s, until-build: %s", version.IdeaVersion.SinceBuild, version.IdeaVersion.UntilBuild))
	logger.Debug(fmt.Sprintf("IntelliJ build ranges: %+v", buildRanges))

	for _, buildRange := range buildRanges {
		// Handle wildcard in plugin until-build (e.g., "243.*")
		untilBuild := version.IdeaVersion.UntilBuild
		untilBuild = strings.TrimSuffix(untilBuild, ".*")

		// Handle empty plugin until-build as open-ended (like *)
		if untilBuild == "" {
			untilBuild = "999999" // Treat empty until-build as open-ended
		}

		// Convert plugin since/until-build to floats (handle decimal points)
		pluginSinceBuildFloat, err := strconv.ParseFloat(version.IdeaVersion.SinceBuild, 64)
		if err != nil {
			logger.Debug(fmt.Sprintf("Skipping plugin version %s due to invalid since-build", version.Version))
			continue
		}

		pluginUntilBuildFloat, _ := strconv.ParseFloat(untilBuild, 64)
		ideaSinceBuildInt, _ := strconv.Atoi(buildRange.SinceBuild)

		// Handle "*" in config until-build, meaning open-ended compatibility
		ideaUntilBuildInt := 999999 // Default to a large number if until-build is "*"
		if buildRange.UntilBuild != "*" {
			ideaUntilBuildInt, _ = strconv.Atoi(buildRange.UntilBuild)
		}

		// Convert floats to integers for comparison, rounding the plugin build numbers
		pluginSinceBuildInt := int(math.Floor(pluginSinceBuildFloat))
		pluginUntilBuildInt := int(math.Floor(pluginUntilBuildFloat))

		// Debugging: Log the integer values being compared
		logger.Debug(fmt.Sprintf("Comparing plugin (since: %d, until: %d) with IntelliJ (since: %d, until: %d)",
			pluginSinceBuildInt, pluginUntilBuildInt, ideaSinceBuildInt, ideaUntilBuildInt))

		// Check if the plugin version is compatible with the idea build range
		if pluginSinceBuildInt >= ideaSinceBuildInt && pluginUntilBuildInt <= ideaUntilBuildInt {
			logger.Debug(fmt.Sprintf("Plugin version %s is compatible with IntelliJ", version.Version))
			return true
		}
	}

	// No compatible build range found
	logger.Debug("No compatible plugin version found for IntelliJ")
	return false
}

func extractBuildRange(config models.Config) []models.BuildRange {
	// Return the list of build ranges directly from IntelliJ in the config
	return config.IntelliJ.Builds
}

// downloadFile constructs the correct download URL and saves the plugin file locally, along with its metadata
func downloadFile(pluginID, pluginVersion, sinceBuild, untilBuild, description string) error {
	downloadURL := fmt.Sprintf("https://plugins.jetbrains.com/plugin/download?pluginId=%s&version=%s", pluginID, pluginVersion)
	logger.Debug(fmt.Sprintf("Starting download from URL: %s", downloadURL))

	resp, err := http.Get(downloadURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to make request to download URL %s: %v", downloadURL, err))
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download file: invalid response code %d", resp.StatusCode)
	}

	// Create the plugin's directory
	outputDir := filepath.Join("output", "plugins", pluginID, pluginVersion)
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Save the zip file
	fileName := fmt.Sprintf("%s-intellij-bin-%s.zip", pluginID, pluginVersion)
	filePath := filepath.Join(outputDir, fileName)
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	// Save the metadata file
	metadata := models.PluginMetadata{
		ID:          pluginID,
		Version:     pluginVersion,
		SinceBuild:  sinceBuild,
		UntilBuild:  untilBuild,
		Description: description,
	}
	metadataFile := filepath.Join(outputDir, "metadata.json")
	metaFile, err := os.Create(metadataFile)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %v", err)
	}
	defer metaFile.Close()

	encoder := json.NewEncoder(metaFile)
	err = encoder.Encode(metadata)
	if err != nil {
		return fmt.Errorf("failed to encode metadata: %v", err)
	}

	logger.Info(fmt.Sprintf("Downloaded plugin %s version %s to %s", pluginID, pluginVersion, filePath))
	return nil
}
