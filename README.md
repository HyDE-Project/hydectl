# hydectl

hydectl is a powerful CLI tool for managing HyDE configurations and scripts. It provides an extensible interface for executing both built-in commands and user-defined plugin scripts from specified directories.

## Features

- **Plugin System**: Execute custom scripts from configured directories
- **Command-Line Interface**: Intuitive commands and subcommands
- **Dynamic Command Loading**: Automatically discovers and loads available scripts
- **Theme Management**: Import, select, and customize themes
- **Wallpaper Management**: Easily manage desktop wallpapers
- **Window Management**: Group/ungroup windows and control zoom in Hyprland
- **Logging System**: Configurable logging levels for debugging

## Installation

### Using Go

To build from source, make sure Go is installed, then clone the repository and build the project:

```sh
pacman -S --needed go  # or your system's package manager
git clone https://github.com/HyDE-Project/hydectl.git
cd hydectl
make all
```

To install the binary to ~/.local/lib:

```sh
make install
```

### Direct Binary Installation

Alternatively, you can copy the pre-built binary directly:

```sh
cp /bin/hydectl ~/.local/bin/
chmod +x ~/.local/bin/hydectl
```

### Uninstallation

To uninstall the binary:

```sh
make uninstall
```

## Usage

hydectl provides a command-line interface for executing commands and scripts.

### Execute a Script with Dispatch

```sh
hydectl dispatch <script_name> [args...]
```

### Using Built-in Commands

```sh
hydectl [command] [subcommand] [flags]
```

For example:

```sh
hydectl theme select [theme_name]
hydectl theme import --name mytheme --url https://example.com/theme
```

### Advanced Usage

To pass additional arguments directly to the command:

```sh
hydectl [command] -- [additional args]
```

### List Available Scripts

```sh
hydectl --list
```

### Help

```sh
hydectl --help
hydectl [command] --help
```

## Configuration

hydectl searches for scripts in the following directories:

- `${XDG_CONFIG_HOME}/lib/hydectl/scripts`
- `${HOME}/.local/lib/hydectl/scripts`
- `/usr/local/lib/hydectl/scripts`
- `/usr/lib/hydectl/scripts`

You can add your custom scripts to any of these directories to make them available for execution.

### Logging Configuration

Set the `LOG_LEVEL` environment variable to control logging verbosity:

```sh
LOG_LEVEL=debug hydectl [command]
```

Valid log levels: `debug`, `info`, `error`, `silent` (default)

## Available Commands

- **completion**: Generate the autocompletion script for the specified shell
- **dispatch**: Dispatch a plugin command (executes external scripts)
- **reload**: Reload the HyDE configuration
- **select**: Select various items
- **tabs**: Group or ungroup all windows in the current workspace
- **theme**: Manage themes
- **version**: Print the version number
- **wallpaper**: Manage wallpapers
- **zoom**: Zoom in/out Hyprland

### Theme Management

- `hydectl theme select`: Select a theme
- `hydectl theme next`: Switch to the next theme
- `hydectl theme prev`: Switch to the previous theme
- `hydectl theme set`: Set a specific theme
- `hydectl theme import`: Import themes from repositories or URLs

## Plugin Development

You can extend hydectl by creating custom script plugins. Script plugins should:

1. Be executable (with proper permissions)
2. Be placed in one of the script directories
3. Implement the required interface for command usage

Example Plugin Script:

```sh
#!/bin/bash
# This is an example script for hydectl.
# You can modify this script to create your own custom commands.

echo "Hello from example_script.sh!"
echo "This script can be executed via the hydectl dispatch <script>."
```

Once saved in one of the script directories, you can execute it with:

```sh
hydectl dispatch example_script
```

> [!Note]
> This is not limited to bash scripts. You can use any language that can be executed from the command line.

## Contributing

Contributions are welcome! Please see our [CONTRIBUTING.md](./CONTRIBUTING.md) file for details on:

- Reporting bugs
- Suggesting enhancements
- Submitting code changes
- Commit message guidelines
- Pull request process

## License

This project is licensed under the project's license. See the LICENSE file for details.

## Tips

- Use "hydectl [command] --help" for more information about a command.
- To PASS additional arguments directly to the command, append '--' before the arguments.
- Current version: r23.b3fb401
