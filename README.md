# Simple Screenshot Folder Watcher and Automatically Sync With Google Drive

(Currently) A CLI-based simple folder watcher, entirely written in Golang. This is the initial version, focusing on monitoring a local screenshot folder. Whenever a file (screenshot) is added or removed in the specified folder, the program will automatically sync these changes with Google Drive. Additionally, the folder can be shared with other Google accounts specified when starting the program.

This initial version is straightforward, supporting one-way synchronization from your local folder to Google Drive. New files in the local folder will be uploaded to Google Drive, and deleted files will also be removed from Google Drive. However, changes made directly in the Google Drive folder are not synced back to the local folder.

File metadata and registered accounts for shared folder access are stored in a lightweight SQLite database, providing a historical record and laying the groundwork for potential future enhancements.

## Getting Started

### 1. Configure Your Google Service Account
Currently, the program uses a Google Service Account for accessing the Google Drive API. (In the future, the goal is to support direct integration with your Google account via OAuth2.)

To get started, prepare a JSON file for your [Google Service Account](https://cloud.google.com/iam/docs/service-account-overview). Then, set the path to this JSON file in the **main.go** configuration.

### 2. Install Dependencies
Run the following command to ensure all necessary modules are installed:

```bash
go mod tidy
```

### 3. Start watcher
To start watcher, run:

```bash
go run main.go
```

This will start the program and automatically reload changes when you rerun the command after making updates.

Alternatively, for live reloading during development, you can use **air** by running:

```bash
air
```

Make sure to configure air according to your project's needs by adjusting the settings in the `.air.toml` file.

### 4. Start with Binary
You can build the binary and run it:

#### On Windows:
```bash
go build -o lorem.exe
lorem.exe
```

#### On Linux/macOS:
```bash
go build -o lorem
./lorem
```