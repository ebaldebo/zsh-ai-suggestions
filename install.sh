#!/bin/bash

set -e

VERSION="v0.1.0-beta"
INSTALL_DIR="$HOME/.local/bin"
ZSH_PLUGIN_DIR="$HOME/.zsh/plugins/zsh-ai-suggestions"
REPO_URL="https://github.com/ebaldebo/zsh-ai-suggestions/releases/download/${VERSION}"
TEMP_DIR=$(mktemp -d)

cleanup() {
  rm -rf "$TEMP_DIR"
}

trap cleanup EXIT

OS=$(uname -s)
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="X86_64" ;;
    amd64) ARCH="X86_64" ;;
    arm64) ARCH="arm64" ;;
    aarch64) ARCH="arm64" ;;
    i386|i686) ARCH="i386" ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

ARCHIVE_FORMAT="tar.gz"

DOWNLOAD_URL="${REPO_URL}/zsh-ai-suggestions_${OS}_${ARCH}.${ARCHIVE_FORMAT}"
CHECKSUM_URL="${REPO_URL}/zsh-ai-suggestions_${VERSION}_checksums.txt"

echo "Downloading zsh-ai-suggestions ${VERSION}_${OS}_${ARCH}..."
curl -sL "$DOWNLOAD_URL" -o "$TEMP_DIR/archive.$ARCHIVE_FORMAT"
curl -sL "$CHECKSUM_URL" -o "$TEMP_DIR/checksums.txt"

cd "$TEMP_DIR"
# if command -v sha256sum > /dev/null; then
#     CHECKSUM_TOOL="sha256sum"
# elif command -v shasum > /dev/null; then
#     CHECKSUM_TOOL="shasum -a 256"
# else
#     echo "Warning: Could not find sha256sum or shasum, skipping checksum verification"
#     CHECKSUM_TOOL=""
# fi

# if [ -n "$CHECKSUM_TOOL" ]; then
#     echo "Verifying checksum..."
#     grep "zsh-ai-suggestions_${OS}_${ARCH}.${ARCHIVE_FORMAT}" checksums.txt | $CHECKSUM_TOOL -c
#     if [ $? -ne 0 ]; then
#         echo "Checksum verification failed"
#         exit 1
#     fi
# fi

echo "Extracting archive..."
if [[ "$ARCHIVE_FORMAT" == "tar.gz" ]]; then
    tar -xzf "archive.$ARCHIVE_FORMAT"
else
    echo "Unsupported archive format: $ARCHIVE_FORMAT"
    exit 1
fi

mkdir -p "$INSTALL_DIR"

echo "Installing zsh-ai-suggestions to $INSTALL_DIR..."
install -m 755 "zsh-ai-suggestions" "$INSTALL_DIR/"
echo "Installation complete"
echo "The binary is located at $INSTALL_DIR/zsh-ai-suggestions"