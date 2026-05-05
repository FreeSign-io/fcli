# fcli

The official command-line tool for [FreeSign](https://freesign.io), the free, open-source eSignature platform.

```bash
brew tap FreeSign-io/fcli
brew install fcli

fcli auth login
fcli docs send contract.pdf --to "Jane:jane@example.com" --send
```

## What it does

`fcli` wraps the FreeSign HTTP API so you can run the entire signing workflow from your terminal:

- Send PDFs for signature with one command
- List, get, download, cancel, resend envelopes
- Manage recipients and signature fields
- Use templates
- Manage webhooks

## Install

### Homebrew (macOS, Linux)

```bash
brew tap FreeSign-io/fcli
brew install fcli
```

### Pre-built binaries

Grab the latest release for your platform from
[GitHub Releases](https://github.com/FreeSign-io/fcli/releases).

### From source

```bash
go install github.com/FreeSign-io/fcli@latest
```

## Quickstart

```bash
# Authenticate (paste a token from /t/<team>/settings/tokens)
fcli auth login

# Confirm
fcli auth whoami

# Send a SAFE
fcli docs send ~/Documents/safe.pdf \
  --to "Araik:araik@example.com" \
  --subject "Elixion SAFE — \$25k" \
  --message "Please review and sign." \
  --signature-anchor "INVESTOR" \
  --send

# Watch for completion
fcli docs status <id>

# Download once signed
fcli docs download <id> --out signed.pdf
```

## Self-hosted FreeSign

Point `fcli` at your own deployment:

```bash
fcli config profiles add work --base-url https://sign.example.com
fcli config profiles use work
fcli auth login
```

## Output formats

- TTY → human-readable table
- Pipe → JSON (`fcli docs list | jq '.[] | .title'`)
- Override → `--output json|yaml|table`

## License

[AGPL-3.0](./LICENSE). See `NOTICE` for attribution to FreeSign and Documenso.
