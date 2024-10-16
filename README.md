# IntelliJ Offline Plugin Syncer & Server

This project provides a solution for syncing and serving JetBrains IntelliJ-based IDE plugins in air-gapped environments. The system consists of two components:
1. **Syncer**: Downloads plugins from the JetBrains Marketplace, stores them locally, and generates an `index.xml` file with metadata.
2. **Server**: Hosts the downloaded plugins and serves the `index.xml` file as an offline plugin marketplace.

## Features
- **Air-gapped Support**: Sync and serve IntelliJ plugins without an internet connection.
- **Automatic Plugin Compatibility Check**: Downloads only plugins that are compatible with the specified IntelliJ builds.
- **Metadata Handling**: Stores plugin metadata for easy generation of `index.xml`.
- **Simple HTTP Server**: Serves the plugins and the `index.xml` file, which can be accessed by IntelliJ-based IDEs like PyCharm.

## Requirements
- Go 1.18+ installed.
- Access to JetBrains Marketplace to run the syncer (can be done on a machine with internet access).

## Project Structure
```
.
├── cmd/
│   ├── syncer/          # Main entry point for the syncer
│   └── server/          # Main entry point for the server
├── internal/
│   ├── downloader/      # Handles downloading plugins and saving metadata
│   ├── xmlgenerator/    # Generates the index.xml file from metadata
│   └── models/          # Shared data models (PluginMetadata, etc.)
├── output/
│   └── plugins/         # Where downloaded plugins and metadata are stored
│       └── <plugin-id>/ 
│           └── <version>/
│               ├── <plugin>.zip
│               └── metadata.json
├── README.md            # Project documentation
└── config.json          # Configuration file for the syncer
```

## Setup & Usage

### 1. Sync Plugins from JetBrains Marketplace
First, run the syncer to download the plugins and generate the `index.xml` file.

#### Example `config.json`
Define the plugins you want to sync in a `config.json` file:

```
{
  "intellij": {
    "builds": [
      {
        "since-build": "233",
        "until-build": "*"
      }
    ]
  },
  "plugins": [
    {
      "id": "dev.turingcomplete.intellijdevelopertoolsplugins"
    },
    {
      "id": "org.intellij.scala"
    }
  ]
}
```

#### Running the Syncer
```
go run cmd/syncer/main.go
```
This will:
- Download plugins listed in `config.json` (if compatible with the specified builds).
- Store each plugin's `.zip` file and its metadata in the `output/plugins` directory.
- Generate `output/index.xml`.

### 2. Serve the Plugins Locally
Next, run the server to serve the downloaded plugins and the generated `index.xml`.

#### Running the Server
```
go run cmd/server/main.go
```
This will:
- Start a simple HTTP server on `http://localhost:8080` that serves the plugins and `index.xml`.
- Allow IntelliJ-based IDEs to access the offline marketplace.

### 3. Configure IntelliJ IDE to Use the Offline Marketplace
1. Open your JetBrains IDE (PyCharm, IntelliJ, etc.).
2. Go to `Settings` > `Plugins` > `⚙️` (gear icon) > `Manage Plugin Repositories`.
3. Add `http://localhost:8080/index.xml` as a custom repository.
4. Your IDE will now fetch plugins from your local server.

## Troubleshooting
- Ensure the syncer is run in an environment with internet access.
- Make sure the server is running and accessible from the IDE.
- Check `output/plugins/` for the downloaded plugins and `metadata.json` files.

## Contributing
Contributions are welcome! Please feel free to open issues or submit pull requests.

## License
This project is licensed under the MIT License.
