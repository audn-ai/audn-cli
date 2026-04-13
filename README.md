# audn-cli

Command-line interface for [Audn.ai](https://audn.ai) — voice and text AI agent security testing. Run adversarial campaigns, gate CI/CD pipelines, and manage agents from your terminal.

## Install

**macOS / Linux:**

```bash
curl -L https://audn.ai/cli/install.sh | sh
```

**Windows:**

```powershell
iwr -useb https://audn.ai/cli/install.ps1 | iex
```

**From source:**

```bash
go install github.com/audn-ai/audn-cli@latest
```

**Or build locally:**

```bash
git clone https://github.com/audn-ai/audn-cli.git
cd audn-cli
make build
```

## Authenticate

```bash
# Interactive login (opens browser via Device Code flow)
audn-cli login

# Or use PKCE flow
audn-cli login --pkce

# Check who you're logged in as
audn-cli whoami

# Logout
audn-cli logout
```

**CI/CD authentication** (no browser):

```bash
# Option 1: API secret key
export AUDN_API_SECRET="sk_your_secret_key"

# Option 2: Auth0 M2M credentials
export AUTH0_M2M_CLIENT_ID="your-client-id"
export AUTH0_M2M_CLIENT_SECRET="your-client-secret"

# Option 3: Pre-obtained bearer token
export AUDN_BEARER_TOKEN="your-token"
```

## Commands

### Gate Check (CI/CD)

Run a security gate check — executes a campaign, waits for completion, and exits with a non-zero code if the gate fails.

```bash
# Execute campaign and gate on results
audn-cli gate check --campaign-id 675e383f-aab3-4545-bc19-3eea543abd31

# Fail if any high or critical severity finding
audn-cli gate check --campaign-id $CAMPAIGN_ID --fail-on-severity high

# Require minimum VAST grade
audn-cli gate check --campaign-id $CAMPAIGN_ID --min-grade B

# JSON output for machine parsing
audn-cli gate check --campaign-id $CAMPAIGN_ID --json

# Don't wait for completion
audn-cli gate check --campaign-id $CAMPAIGN_ID --wait=false
```

### Campaigns

```bash
# List campaigns
audn-cli campaigns list

# Create a campaign
audn-cli campaigns create --name "Q1 Audit" --agent-id $AGENT_ID

# Get campaign details
audn-cli campaigns get --id $CAMPAIGN_ID

# Execute a campaign
audn-cli campaigns execute --id $CAMPAIGN_ID

# Delete a campaign
audn-cli campaigns delete --id $CAMPAIGN_ID
```

### Agents

```bash
# List agents
audn-cli agents list

# Create agent
audn-cli agents create --name "IVR Bot" --platform twilio --phone +14155551234

# Get agent details
audn-cli agents get --id $AGENT_ID

# Delete agent
audn-cli agents delete --id $AGENT_ID
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Voice Agent Security Test

on: [push]

jobs:
  security-test:
    runs-on: ubuntu-latest
    steps:
      - name: Install Audn CLI
        run: curl -L https://audn.ai/cli/install.sh | sh

      - name: Run Security Gate
        env:
          AUDN_API_SECRET: ${{ secrets.AUDN_API_SECRET }}
        run: |
          audn-cli gate check \
            --campaign-id ${{ vars.CAMPAIGN_ID }} \
            --fail-on-severity high \
            --min-grade B
```

### GitLab CI

```yaml
security-test:
  image: golang:1.22
  script:
    - curl -L https://audn.ai/cli/install.sh | sh
    - audn-cli gate check --campaign-id $CAMPAIGN_ID --fail-on-severity high
  variables:
    AUDN_API_SECRET: $AUDN_API_SECRET
```

### Jenkins

```groovy
pipeline {
    agent any
    environment {
        AUDN_API_SECRET = credentials('audn-api-secret')
    }
    stages {
        stage('Security Test') {
            steps {
                sh 'curl -L https://audn.ai/cli/install.sh | sh'
                sh 'audn-cli gate check --campaign-id ${CAMPAIGN_ID} --fail-on-severity high'
            }
        }
    }
}
```

## Global Flags

| Flag | Env Variable | Description |
|------|-------------|-------------|
| `--api-secret` | `AUDN_API_SECRET` | API secret key (sk_...) |
| `--bearer-token` | `AUDN_BEARER_TOKEN` | Pre-obtained bearer token |
| `--m2m-client-id` | `AUTH0_M2M_CLIENT_ID` | Auth0 M2M client ID |
| `--m2m-client-secret` | `AUTH0_M2M_CLIENT_SECRET` | Auth0 M2M client secret |
| `--api-url` | `AUDN_API_URL` | API base URL (default: https://audn.ai) |
| `--json` | `AUDN_OUTPUT_JSON` | JSON output |

## Gate Check Flags

| Flag | Env Variable | Default | Description |
|------|-------------|---------|-------------|
| `--campaign-id` | - | - | Campaign to execute |
| `--agent-id` | `AUDN_AGENT_ID` | - | Agent to test (legacy) |
| `--fail-on-severity` | `AUDN_FAIL_ON_SEVERITY` | `critical` | Fail threshold |
| `--min-grade` | `AUDN_MIN_GRADE` | - | Minimum VAST grade (A-F) |
| `--wait` | `AUDN_WAIT` | `true` | Wait for completion |
| `--timeout` | `AUDN_TIMEOUT` | `600` | Timeout in seconds |

## Credential Storage

Credentials from `audn-cli login` are stored at:

| OS | Path |
|----|------|
| macOS / Linux | `~/.config/audn/credentials.json` |
| Windows | `%APPDATA%\audn\credentials.json` |

File permissions are set to `600` (owner read/write only).

## Direct Downloads

| Platform | Architecture | Download |
|----------|-------------|----------|
| macOS | Apple Silicon (arm64) | [audn-darwin-arm64](https://audn.ai/cli/latest/audn-darwin-arm64) |
| macOS | Intel (amd64) | [audn-darwin-amd64](https://audn.ai/cli/latest/audn-darwin-amd64) |
| Linux | x86_64 | [audn-linux-amd64](https://audn.ai/cli/latest/audn-linux-amd64) |
| Linux | arm64 | [audn-linux-arm64](https://audn.ai/cli/latest/audn-linux-arm64) |
| Windows | x86_64 | [audn-windows-amd64.exe](https://audn.ai/cli/latest/audn-windows-amd64.exe) |

## Development

```bash
# Build
make build

# Build all platforms
make build-all

# Run tests
make test

# Lint
make lint

# Full CI check
make ci
```

## License

GPL-3.0 — see [LICENSE](LICENSE). The CLI is open-source. The Audn.ai API is subject to [Audn.ai Terms of Service](https://audn.ai/terms).

---

**[audn.ai](https://audn.ai)** — Continuous security testing for AI agents.
