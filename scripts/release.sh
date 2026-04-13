#!/bin/bash

# Release script for audn-cli
# This script creates a GitHub release and uploads built binaries

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="audn"
GITHUB_OWNER="audn-ai"
GITHUB_REPO="audn.ai"

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if gh CLI is installed
check_gh_cli() {
    if ! command -v gh &> /dev/null; then
        print_error "GitHub CLI (gh) is not installed"
        echo "Install it from: https://cli.github.com/"
        echo "Or run: brew install gh"
        exit 1
    fi
    
    # Check if authenticated
    if ! gh auth status &> /dev/null; then
        print_error "Not authenticated with GitHub"
        echo "Run: gh auth login"
        exit 1
    fi
}

# Parse arguments
VERSION=""
DRAFT=false
PRERELEASE=false
NOTES=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -d|--draft)
            DRAFT=true
            shift
            ;;
        -p|--prerelease)
            PRERELEASE=true
            shift
            ;;
        -n|--notes)
            NOTES="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 -v VERSION [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -v, --version VERSION    Version to release (e.g., v1.0.0) [required]"
            echo "  -d, --draft             Create as draft release"
            echo "  -p, --prerelease        Mark as pre-release"
            echo "  -n, --notes NOTES       Release notes (optional)"
            echo "  -h, --help              Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Validate version
if [ -z "$VERSION" ]; then
    print_error "Version is required. Use -v or --version"
    exit 1
fi

# Ensure version starts with 'v'
if [[ ! "$VERSION" =~ ^v ]]; then
    VERSION="v$VERSION"
fi

print_info "Preparing release for version: $VERSION"

# Check prerequisites
check_gh_cli

# Build all binaries
print_info "Building binaries for all platforms..."
make build-all

# Check if dist directory exists and has files
if [ ! -d "dist" ] || [ -z "$(ls -A dist)" ]; then
    print_error "No binaries found in dist/ directory"
    print_info "Run 'make build-all' first"
    exit 1
fi

# Create compressed archives
print_info "Creating compressed archives..."
cd dist

for binary in ${BINARY_NAME}-*; do
    if [ -f "$binary" ]; then
        if [[ "$binary" == *.exe ]]; then
            # Windows binary - create zip
            zip_name="${binary%.exe}.zip"
            print_info "Creating $zip_name"
            zip -q "$zip_name" "$binary"
            rm "$binary"
        else
            # Unix binary - create tar.gz
            tar_name="${binary}.tar.gz"
            print_info "Creating $tar_name"
            tar -czf "$tar_name" "$binary"
            rm "$binary"
        fi
    fi
done

# Generate checksums
print_info "Generating checksums..."
shasum -a 256 *.{tar.gz,zip} 2>/dev/null > checksums.txt || sha256sum *.{tar.gz,zip} > checksums.txt

cd ..

# Generate release notes if not provided
if [ -z "$NOTES" ]; then
    print_info "Generating release notes..."
    NOTES=$(cat <<EOF
# audn CLI ${VERSION}

## Installation

### macOS/Linux
\`\`\`bash
# Download the appropriate binary for your system
curl -L https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/cli/${VERSION}/${BINARY_NAME}-\$(uname -s | tr '[:upper:]' '[:lower:]')-\$(uname -m | sed 's/x86_64/amd64/').tar.gz | tar xz
sudo mv ${BINARY_NAME} /usr/local/bin/
${BINARY_NAME} --version
\`\`\`

### Windows
Download the Windows binary from the assets below and add it to your PATH.

### Using npm
\`\`\`bash
npm install -g @audn-ai/cli
\`\`\`

### Using Docker
\`\`\`bash
docker run ghcr.io/${GITHUB_OWNER}/${BINARY_NAME}-cli:${VERSION}
\`\`\`

## Checksums
\`\`\`
$(cat dist/checksums.txt)
\`\`\`

## What's Changed
Check the commit history for detailed changes.
EOF
)
fi

# Create git tag
TAG_NAME="cli/${VERSION}"
print_info "Creating git tag: $TAG_NAME"

# Check if tag already exists
if git rev-parse "$TAG_NAME" >/dev/null 2>&1; then
    print_warn "Tag $TAG_NAME already exists locally"
    read -p "Do you want to delete and recreate it? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag -d "$TAG_NAME"
        git push origin --delete "$TAG_NAME" 2>/dev/null || true
    else
        print_info "Using existing tag"
    fi
else
    git tag -a "$TAG_NAME" -m "Release ${VERSION}"
fi

# Push tag
print_info "Pushing tag to GitHub..."
git push origin "$TAG_NAME"

# Build release command
RELEASE_CMD="gh release create $TAG_NAME"

if [ "$DRAFT" = true ]; then
    RELEASE_CMD="$RELEASE_CMD --draft"
fi

if [ "$PRERELEASE" = true ]; then
    RELEASE_CMD="$RELEASE_CMD --prerelease"
fi

RELEASE_CMD="$RELEASE_CMD --title \"audn CLI ${VERSION}\""
RELEASE_CMD="$RELEASE_CMD --notes \"$NOTES\""

# Add all artifacts
for file in dist/*.{tar.gz,zip,txt} 2>/dev/null; do
    if [ -f "$file" ]; then
        RELEASE_CMD="$RELEASE_CMD \"$file\""
    fi
done

# Create GitHub release
print_info "Creating GitHub release..."
eval $RELEASE_CMD

if [ $? -eq 0 ]; then
    print_info "✅ Release created successfully!"
    print_info "View release at: https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/tag/${TAG_NAME}"
    
    # Cleanup
    read -p "Do you want to clean up the dist directory? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf dist/
        print_info "Cleaned up dist directory"
    fi
else
    print_error "Failed to create release"
    exit 1
fi

print_info "Release process complete!"