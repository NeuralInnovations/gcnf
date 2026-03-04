#!/usr/bin/env bash

# Set the repository owner and name
OWNER="NeuralInnovations"
REPOSITORY="gcnf"
CPU_ARCH=$(arch)

# Determine the current platform
platform=""
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    platform="linux-amd64"
    install_dir="/usr/local/bin"
    if [[ "$CPU_ARCH" == "aarch64" ]] || [[ "$CPU_ARCH" == "arm64" ]]; then
        platform="linux-arm64"
    fi
elif [[ "$OSTYPE" == "darwin"* ]]; then
    platform="darwin-amd64"
    install_dir="/usr/local/bin"
    if [[ "$CPU_ARCH" == "aarch64" ]] || [[ "$CPU_ARCH" == "arm64" ]]; then
        platform="darwin-arm64"
    fi
elif [[ "$OSTYPE" == "msys"* || "$OSTYPE" == "cygwin"* || "$OSTYPE" == "win32" ]]; then
    platform="windows"
    install_dir="$HOME/bin"
else
    echo "Unsupported platform: $OSTYPE"
    exit 1
fi

# Determine the filename for the current platform
filename="gcnf-$platform"

download_url="https://github.com/$OWNER/$REPOSITORY/releases/latest/download/$filename"

# Download the binary
echo "Downloading $filename..."
if ! curl -fSL -o gcnf "$download_url"; then
    echo "------------------------------"
    echo "Error: Unable to download gcnf for platform: $platform"
    echo "------------------------------"
    echo "Arch: $CPU_ARCH"
    echo "Download URL: $download_url"
    echo "Please check the repository for available releases:"
    echo "https://github.com/$OWNER/$REPOSITORY/releases"
    echo "------------------------------"
    exit 1
fi

# Make the binary executable (skipped on Windows)
if [[ "$platform" != "windows" ]]; then
    chmod +x gcnf
fi

# Create the install directory if it doesn't exist (for Windows)
if [[ "$platform" == "windows" && ! -d "$install_dir" ]]; then
    mkdir -p "$install_dir"
fi

# Move the binary to the install directory
echo "Installing to $install_dir..."
if [[ "$platform" == "windows" ]]; then
    mv gcnf "$install_dir/gcnf.exe"
else
    sudo mv gcnf "$install_dir/gcnf"
fi

# Add the directory to PATH (for Windows) if it's not already
if [[ "$platform" == "windows" ]]; then
    if [[ ":$PATH:" != *":$install_dir:"* ]]; then
        echo "Adding $install_dir to PATH in .bashrc..."
        echo "export PATH=\$PATH:$install_dir" >> ~/.bashrc
    fi
fi

echo "Installation complete. You can now use 'gcnf'."
