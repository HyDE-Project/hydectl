# hydectl

CLI too for Hyde.

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

### Help

```sh
hydectl --help
hydectl [command] --help
```

## Contributing

Contributions are welcome! Please see our [CONTRIBUTING.md](./CONTRIBUTING.md) file for details on:

- Reporting bugs
- Suggesting enhancements
- Submitting code changes
- Commit message guidelines
- Pull request process
