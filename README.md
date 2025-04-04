# hydectl

hydectl is a CLI tool for managing HyDE configurations and scripts. It allows you to execute built-in commands and user-defined scripts from specified directories.

## Installation

To install hydectl, make sure "go" is installed, clone the repository and build the project:

```sh
pacman -S --needed go
git clone https://github.com/HyDE-Project/hydectl.git
cd hydectl
make all
```

To install the binary to ~/.local/lib:

```sh
make install
```

To uninstall the binary:

```sh
make uninstall
```

## Usage

hydectl provides a command-line interface for executing commands and scripts. Below are some examples of how to use hydectl.

### Execute a Script

To execute a script from the configured script directories:

```sh
./hydectl <script_name> [args...]
```

### List Available Scripts

To list all available scripts:

```sh
./hydectl --list
```

### Help

To display the help message:

```sh
./hydectl --help
```

## Configuration

hydectl searches for scripts in the following directories:

- `${XDG_CONFIG_HOME}/lib/hydectl/scripts`
- `${XDG_DATA_HOME}/lib/hydectl/scripts`
- `${XDG_DATA_HOME}/lib/hyde`
- `/usr/local/lib/hydectl/scripts`
- `/usr/lib/hydectl/scripts`

You can add your scripts to any of these directories to make them available for execution.

## Available Commands

- `hydectl <script_name> [args...]`: Execute a script from the configured script directories.
- `hydectl --list`: List all available scripts.
- `hydectl help`: Display the help message.
- `hydectl completion <shell>`: Generate a completion script for the specified shell (bash, zsh, fish, powershell).

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.
