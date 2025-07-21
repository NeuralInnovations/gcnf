#!/usr/bin/env bash

# Set the repository owner and name
OWNER="NeuralInnovations"
REPOSITORY="gcnf"
CPU_ARCH=$(arch)
RETRY_COUNT=3
RETRY_DELAY=5

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

# Get the latest release information using GitHub API, with retries
for i in $(seq 1 $RETRY_COUNT); do
    release_info=$(curl -s "https://api.github.com/repos/$OWNER/$REPOSITORY/releases/latest")
    if [[ -n "$release_info" ]]; then
        break
    fi
    if [[ $i -lt $RETRY_COUNT ]]; then
        echo "Attempt $i failed to fetch release info. Retrying in $RETRY_DELAY seconds..."
        sleep "$RETRY_DELAY"
        ((RETRY_DELAY *= 2))  # Exponential backoff
    fi
done

if [[ -z "$release_info" ]]; then
    echo "Failed to fetch release info after $RETRY_COUNT attempts."
    exit 1
fi

# Determine the filename for the current platform
filename="gcnf-$platform"

# Extract the download URL for the specific platform binary
download_url=$(echo "$release_info" | grep "browser_download_url" | grep "$filename" | cut -d '"' -f 4)

# Check if the download URL was found
if [[ -z "$download_url" ]]; then
    echo "------------------------------"
    echo "Error: Unable to find download URL for platform: $platform"
    echo "------------------------------"
    echo "Arch $CPU_ARCH"
    echo "Download URL for platform: $platform not found, file name: $filename."
    echo "Content of the latest release:"
    echo "'"
    echo "$release_info"
    echo "'"
    echo "Available platforms in the last releases:"
    echo $("$release_info" | grep "browser_download_url")
    echo "Please check the repository for available releases."
    echo "------------------------------"
    exit 1
fi

# Download the binary
echo "Downloading $filename..."
curl -L -o gcnf "$download_url"

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
        # Reload .bashrc to update PATH in the current session
        source ~/.bashrc
    fi
fi

echo "Installation complete. You can now use 'gcnf'."