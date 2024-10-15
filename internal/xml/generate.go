package xmlgenerator

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"jetbrains-offline/internal/models"
	"os"
	"path/filepath"
)

type IdeaPlugin struct {
	ID          string `xml:"id"`
	Name        string `xml:"name"`
	Version     string `xml:"version"`
	URL         string `xml:"url,attr"`
	IdeaVersion struct {
		SinceBuild string `xml:"since-build,attr"`
		UntilBuild string `xml:"until-build,attr"`
	} `xml:"idea-version"`
	Vendor struct {
		Email string `xml:"email,attr"`
		URL   string `xml:"url,attr"`
	} `xml:"vendor"`
	Description string `xml:"description"`
}

type PluginRepository struct {
	XMLName  xml.Name `xml:"plugin-repository"`
	Category Category `xml:"category"`
}

type Category struct {
	Name        string       `xml:"name,attr"`
	IdeaPlugins []IdeaPlugin `xml:"idea-plugin"`
}

// GenerateIndexXML generates the index.xml file based on metadata files in the plugins directory
func GenerateIndexXML(filePath string) error {
	pluginsDir := "output/plugins"
	plugins := []IdeaPlugin{}

	// Walk through each plugin directory to read metadata
	err := filepath.Walk(pluginsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If we find a metadata.json file, process it
		if info.Name() == "metadata.json" {
			metadata, err := readMetadata(path)
			if err != nil {
				return fmt.Errorf("failed to read metadata: %v", err)
			}

			pluginURL := fmt.Sprintf("http://localhost:8080/plugins/%s/%s/%s-intellij-bin-%s.zip",
				metadata.ID, metadata.Version, metadata.ID, metadata.Version)

			plugins = append(plugins, IdeaPlugin{
				ID:      metadata.ID,
				Name:    metadata.ID, // Use plugin ID as the name for simplicity
				Version: metadata.Version,
				URL:     pluginURL,
				IdeaVersion: struct {
					SinceBuild string `xml:"since-build,attr"`
					UntilBuild string `xml:"until-build,attr"`
				}{SinceBuild: metadata.SinceBuild, UntilBuild: metadata.UntilBuild},
				Vendor: struct {
					Email string `xml:"email,attr"`
					URL   string `xml:"url,attr"`
				}{Email: "support@jetbrains.com", URL: "https://www.jetbrains.com"},
				Description: metadata.Description,
			})
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk plugin directory: %v", err)
	}

	// Generate the XML file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create XML file: %v", err)
	}
	defer file.Close()

	repo := PluginRepository{
		Category: Category{
			Name:        "Programming Languages",
			IdeaPlugins: plugins,
		},
	}

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	err = encoder.Encode(repo)
	if err != nil {
		return fmt.Errorf("failed to encode XML: %v", err)
	}

	fmt.Println("Generated index.xml file")
	return nil
}

// readMetadata reads a plugin's metadata.json file and returns the PluginMetadata struct
func readMetadata(metadataFile string) (models.PluginMetadata, error) {
	var metadata models.PluginMetadata
	data, err := ioutil.ReadFile(metadataFile)
	if err != nil {
		return metadata, err
	}

	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}
