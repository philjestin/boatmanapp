# Boatman

A modern desktop interface for Claude Code agents.

## About

Boatman is a native desktop application built with Wails (Go + React) that provides a graphical user interface for interacting with Claude Code CLI. It offers an intuitive way to manage coding sessions, track tasks, review changes, and collaborate with AI agents on your development projects.

## Features

- **Interactive Chat Interface**: Communicate with Claude Code agents in a clean, modern UI
- **Project Management**: Open and manage multiple coding projects
- **Session Management**: Create, switch between, and manage multiple agent sessions
- **Task Tracking**: View and monitor tasks created by the agent during sessions
- **Git Integration**: View repository status and file diffs directly in the app
- **Approval Workflow**: Review and approve/reject agent actions before they execute
- **MCP Server Support**: Configure and manage Model Context Protocol servers
- **Customizable Settings**: Configure API keys, approval modes, and preferences
- **Onboarding Flow**: Guided setup for first-time users

## Prerequisites

- **Go 1.18+**: Required to build the application
- **Node.js 16+**: Required for the frontend
- **Wails CLI**: Install with `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Claude CLI**: The Claude Code command-line interface must be installed and accessible in your PATH

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd boatmanapp
   ```

2. Install dependencies:
   ```bash
   wails doctor  # Check if all dependencies are installed
   cd frontend && npm install && cd ..
   ```

3. Build the application:
   ```bash
   wails build
   ```

The built application will be in the `build/bin` directory.

## Getting Started

### First Launch

1. Launch Boatman
2. Complete the onboarding wizard:
   - Configure your API authentication (API Key or GCP)
   - Set your approval mode preference
   - Choose optional settings

### Creating Your First Session

1. Click **"New Session"** in the header or use the welcome screen
2. Select a project folder when prompted (or use the currently open project)
3. Start chatting with the agent in the chat interface
4. Review and approve actions when the agent requests permission

### Opening a Project

1. Click **"Open Project"** in the header
2. Navigate to and select your project folder
3. The project will be added to your recent projects list in the sidebar

## Usage

### Chat Interface

- Type messages in the input area at the bottom
- Press Enter or click the send button to send messages
- View agent responses, code blocks, and diffs in the chat history
- Switch between **Chat** and **Tasks** tabs to view different information

### Managing Sessions

- **Create**: Click "New Session" to start a new agent session
- **Switch**: Click on a session in the sidebar to switch between sessions
- **Delete**: Right-click or use the delete option to remove a session

### Approval Workflow

When an agent requests approval for an action:
1. An approval bar appears at the bottom of the window
2. Review the proposed action in the chat
3. Click **"Approve"** to proceed or **"Reject"** to deny

### Settings

Access settings by clicking the gear icon in the header:

- **Authentication**: Configure API key or GCP credentials
- **Approval Mode**: Set when approval is required (always, auto, or never)
- **MCP Servers**: Configure Model Context Protocol servers
- **Preferences**: Customize app behavior

### Git Integration

For projects in a git repository:
- View repository status in the project panel
- See modified, added, deleted, and untracked files
- View diffs for individual files

## Development

### Live Development Mode

Run the app in development mode with hot reload:

```bash
wails dev
```

This starts:
- A Vite development server for fast frontend hot reload
- A dev server at http://localhost:34115 for browser-based development

### Project Structure

```
boatmanapp/
├── frontend/          # React TypeScript frontend
│   ├── src/
│   │   ├── components/  # UI components
│   │   ├── hooks/       # React hooks
│   │   └── store/       # State management
├── agent/             # Agent session management
├── project/           # Project and workspace management
├── config/            # Configuration and preferences
├── git/               # Git integration
├── diff/              # Diff parsing and rendering
├── mcp/               # MCP server management
├── app.go             # Main application logic
└── main.go            # Application entry point
```

### Configuration

Edit `wails.json` to configure build settings. More information: https://wails.io/docs/reference/project-config

## Building

### Development Build

```bash
wails build
```

### Production Build

```bash
wails build -clean -production
```

### Platform-Specific Builds

```bash
# macOS
wails build -platform darwin/universal

# Windows
wails build -platform windows/amd64

# Linux
wails build -platform linux/amd64
```

Built applications will be in `build/bin/`.

## Troubleshooting

### Claude CLI Not Found

Ensure the Claude CLI is installed and in your system PATH:
```bash
which claude  # macOS/Linux
where claude  # Windows
```

### Dependencies Issues

Run the Wails doctor to check your setup:
```bash
wails doctor
```

## License

See LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.
