#!/bin/bash

# Simple script to upload existing builds to GitHub releases

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}GitHub Release Upload Script${NC}"
echo "================================"

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) is not installed${NC}"
    echo "Install it with: brew install gh"
    echo "Or visit: https://cli.github.com/"
    exit 1
fi

# Check authentication
if ! gh auth status &> /dev/null; then
    echo -e "${RED}Error: Not authenticated with GitHub${NC}"
    echo "Run: gh auth login"
    exit 1
fi

# Get version
if [ -z "$1" ]; then
    echo -e "${YELLOW}Usage: $0 <version>${NC}"
    echo "Example: $0 v1.0.0"
    exit 1
fi

VERSION=$1
if [[ ! "$VERSION" =~ ^v ]]; then
    VERSION="v$VERSION"
fi

TAG_NAME="cli/$VERSION"

echo -e "${GREEN}[1/5]${NC} Checking for existing builds..."

# Check if we have built binaries
if [ ! -d "dist" ]; then
    echo -e "${YELLOW}No dist directory found. Building binaries...${NC}"
    make build-all
fi

# Change to dist directory
cd dist

echo -e "${GREEN}[2/5]${NC} Creating archives..."

# Create archives for each binary
for binary in audn-*; do
    if [ -f "$binary" ]; then
        if [[ "$binary" == *.exe ]]; then
            # Windows - create zip
            if [ ! -f "${binary%.exe}.zip" ]; then
                echo "  Creating ${binary%.exe}.zip"
                zip -q "${binary%.exe}.zip" "$binary"
            fi
        else
            # Unix - create tar.gz
            if [ ! -f "${binary}.tar.gz" ]; then
                echo "  Creating ${binary}.tar.gz"
                tar -czf "${binary}.tar.gz" "$binary"
            fi
        fi
    fi
done

echo -e "${GREEN}[3/5]${NC} Generating checksums..."
shasum -a 256 *.tar.gz *.zip 2>/dev/null > checksums.txt || sha256sum *.tar.gz *.zip > checksums.txt

cd ..

echo -e "${GREEN}[4/5]${NC} Creating git tag..."

# Create and push tag if it doesn't exist
if ! git rev-parse "$TAG_NAME" >/dev/null 2>&1; then
    git tag -a "$TAG_NAME" -m "Release $VERSION"
    git push origin "$TAG_NAME"
    echo "  Created and pushed tag: $TAG_NAME"
else
    echo "  Tag already exists: $TAG_NAME"
fi

echo -e "${GREEN}[5/5]${NC} Creating GitHub release..."

# Check if release already exists
if gh release view "$TAG_NAME" >/dev/null 2>&1; then
    echo -e "${YELLOW}Release already exists for $TAG_NAME${NC}"
    read -p "Do you want to upload artifacts to existing release? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Upload to existing release
        gh release upload "$TAG_NAME" dist/*.tar.gz dist/*.zip dist/checksums.txt --clobber
        echo -e "${GREEN}✅ Artifacts uploaded to existing release${NC}"
    fi
else
    # Create new release
    gh release create "$TAG_NAME" \
        --title "audn CLI $VERSION" \
        --notes "# audn CLI $VERSION

## Installation

### Quick Install (macOS/Linux)
\`\`\`bash
curl -L https://github.com/\$(gh repo view --json owner,name -q '.owner.login + \"/\" + .name')/releases/download/$TAG_NAME/audn-\$(uname -s | tr '[:upper:]' '[:lower:]')-\$(uname -m | sed 's/x86_64/amd64/').tar.gz | tar xz
sudo mv audn /usr/local/bin/
\`\`\`

### Checksums
\`\`\`
$(cat dist/checksums.txt)
\`\`\`" \
        dist/*.tar.gz dist/*.zip dist/checksums.txt
    
    echo -e "${GREEN}✅ Release created successfully!${NC}"
fi

echo ""
echo -e "${GREEN}Release URL:${NC} https://github.com/$(gh repo view --json owner,name -q '.owner.login + "/" + .name')/releases/tag/$TAG_NAME"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Visit the release page to verify everything looks correct"
echo "2. Edit the release notes if needed"
echo "3. If this was a draft, publish it when ready"