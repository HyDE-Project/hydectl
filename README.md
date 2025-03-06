# hydectl - A Command Line Interface for Managing Scripts

## Overview

`hydectl` is a command-line interface (CLI) tool designed to manage and execute scripts dynamically. It features a plugin system that allows users to run custom scripts located in `~/.local/share/hyprctl/scripts/`. The CLI is built using the Cobra library, providing a structured and user-friendly command interface.

## Features

- **Dynamic Plugin System**: Automatically discovers and executes scripts from the specified directory.
- **Built-in Commands**: Comes with several built-in commands for common tasks.
- **Help System**: Automatically generates help messages based on available commands and plugins.

## Installation

To install `hydectl`, clone the repository and build the project:

```bash
git clone https://github.com/yourusername/hydectl.git
cd hydectl
go mod tidy
go build -o hydectl
```

## Usage

After installation, you can run `hydectl` from the command line:

```bash
./hydectl
```

### Built-in Commands

- `help`: Displays help information for all commands.
- `version`: Shows the current version of `hydectl`.

### Running Scripts

To run a script, simply use the command followed by the script name:

```bash
./hydectl run <script_name>
```

Make sure your scripts are located in `~/.local/share/hyprctl/scripts/`.

## Example Script

An example script is provided in the `scripts` directory. You can use it as a template to create your own scripts.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue for any enhancements or bug fixes.
