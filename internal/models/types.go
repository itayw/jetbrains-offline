package models

type BuildRange struct {
	SinceBuild string `json:"since-build"`
	UntilBuild string `json:"until-build"` // This can now be "*" to signify an open-ended range
}

type IntelliJ struct {
	Builds []BuildRange `json:"builds"`
}

type Config struct {
	IntelliJ IntelliJ `json:"intellij"`
	Plugins  []Plugin `json:"plugins"`
}

type Plugin struct {
	ID       string    `json:"id"`
	Versions []Version `json:"versions"` // Add a Versions field to track plugin versions
}

type Version struct {
	Version string `json:"version"`
}

// PluginMetadata stores metadata about the downloaded plugin
type PluginMetadata struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	SinceBuild  string `json:"since_build"`
	UntilBuild  string `json:"until_build"`
	Description string `json:"description"`
}
