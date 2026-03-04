# gcnf

**gcnf** is a CLI tool that uses Google Sheets as a configuration source. It lets you store, retrieve, and inject configuration values from a Google Sheets spreadsheet into your applications and CI/CD pipelines.

---

## Install

Supported platforms: macOS (amd64, arm64), Linux (amd64, arm64), Windows.

**Bash**

```bash
curl https://raw.githubusercontent.com/NeuralInnovations/gcnf/refs/heads/master/install.sh | bash
```

**Zsh**

```zsh
curl https://raw.githubusercontent.com/NeuralInnovations/gcnf/refs/heads/master/install.sh | zsh
```

Verify the installation:

```bash
gcnf version
```

## Uninstall

Remove the installed binary:

```bash
# macOS / Linux
sudo rm /usr/local/bin/gcnf

# Windows
rm "$HOME/bin/gcnf.exe"
```

Remove stored credentials and cached data:

```bash
rm -rf ~/.gcnf
```

---

## Quick Start

```bash
# 1. Install gcnf
curl https://raw.githubusercontent.com/NeuralInnovations/gcnf/refs/heads/master/install.sh | bash

# 2. Authenticate with your Google account
gcnf login

# 3. Set the Google Sheet document ID (found in the sheet URL)
gcnf config --sheetId YOUR_GOOGLE_SHEET_ID

# 4. Get a single configuration value
gcnf get --sheet Env --env Staging --category Database --name Host

# 5. Inject configuration into an .env template
gcnf inject .env.template > .env
```

---

## Google Sheet Format

[Google Sheet Example](https://docs.google.com/spreadsheets/d/1ykOxHza5fxa-HbXPPlSGfTSm2YKG0AteNhDHI6tntUk/edit?gid=0#gid=0)

### Required columns

Every sheet must have **Category** and **Name** columns. All other columns represent environments (e.g. `Env1`, `Staging`, `Production`, `Develop`).

| Category  | Name  | Env1            | Staging           | Production       | Develop         |
|-----------|-------|-----------------|-------------------|------------------|-----------------|
| Category1 | Name1 | value1          | val1              | val2             | v1              |
| Category2 | Name1 | example1        | env1              | env2             | v2              |
| Database  | Host  | http://env1.com | https://stage.com | https://prod.com | https://dev.com |

---

## Authentication

gcnf supports two authentication methods:

### 1. Interactive OAuth2 Login (for development)

```bash
gcnf login
```

Opens a browser for Google OAuth2 authentication. The token is stored at `~/.gcnf/.token`.

To remove the stored token:

```bash
gcnf logout
```

### 2. Service Account (for CI/CD)

Set the `GCNF_GOOGLE_CREDENTIAL_BASE64` environment variable to the Base64-encoded contents of your Google service account JSON key file:

```bash
export GCNF_GOOGLE_CREDENTIAL_BASE64=$(base64 < service_account.json)
```

---

## Commands

### `gcnf login`

Authenticate with Google via OAuth2. Opens a browser to complete the login flow.

```bash
gcnf login
```

### `gcnf logout`

Remove the stored OAuth2 token.

```bash
gcnf logout
```

### `gcnf config` -- Set persistent configuration

View or set the Google Sheet ID that gcnf uses by default.

```bash
# View current sheet ID
gcnf config

# Set the sheet ID (persisted to ~/.gcnf/.google_sheet_id)
gcnf config --sheetId YOUR_GOOGLE_SHEET_ID
gcnf config -i YOUR_GOOGLE_SHEET_ID
```

### `gcnf status` -- Show tool status

Display the current configuration and credential status.

```bash
# Default YAML output
gcnf status

# JSON output
gcnf status --format json
gcnf status -f json
```

Example output:

```yaml
credentials_status: user_token
google_credential_b64: empty
google_sheet_id: 1ykOxHza5fxa-HbXPPlSGfTSm2YKG0AteNhDHI6tntUk
name: gcnf
storage_config_file: ./gcnf_config.json
user_token_file: valid
version: 0.0.8
```

### `gcnf load` (alias: `l`) -- Download config locally

Download configuration data from a Google Sheet and cache it in a local file.

```bash
gcnf load --sheet Env --env staging
gcnf load -s Env -e staging
gcnf l -s Env -e production
```

| Flag      | Short | Required | Description                         |
|-----------|-------|----------|-------------------------------------|
| `--sheet` | `-s`  | Yes      | Sheet (tab) name in the spreadsheet |
| `--env`   | `-e`  | Yes      | Environment column to download      |

### `gcnf unload` (alias: `d`) -- Delete local config cache

Remove locally cached configuration files. Without flags, deletes all cache files. With `--sheet` and `--env`, deletes only that specific cache.

```bash
# Delete all cache files
gcnf unload
gcnf d

# Delete a specific cache
gcnf unload --sheet Env --env staging
```

| Flag      | Short | Required | Description                          |
|-----------|-------|----------|--------------------------------------|
| `--sheet` | `-s`  | No       | Sheet name (for targeted deletion)   |
| `--env`   | `-e`  | No       | Environment (for targeted deletion)  |

### `gcnf get` (alias: `g`) -- Get a single value

Retrieve a specific configuration value from Google Sheets.

```bash
gcnf get --sheet Env --env staging --category Database --name Host
gcnf get -s Env -e staging -c Database -n Host
gcnf g -s Env -e staging -c Database -n Host
```

| Flag         | Short | Required | Description         |
|--------------|-------|----------|---------------------|
| `--sheet`    | `-s`  | Yes      | Sheet (tab) name    |
| `--env`      | `-e`  | Yes      | Environment column  |
| `--category` | `-c`  | Yes      | Category row filter |
| `--name`     | `-n`  | Yes      | Name row filter     |

### `gcnf read` (alias: `r`) -- Read value by URL

Read a configuration value using the `gcnf://` URL format.

```bash
gcnf read "gcnf://Env/Staging/Database/ConnectionString"
gcnf r "gcnf://Env/Production/App/ApiKey"
```

URL format: `gcnf://SHEET/ENV/CATEGORY/NAME`

### `gcnf inject` (alias: `i`) -- Process a template file

Read a template file, resolve all `gcnf://` URLs and environment variable references, and print the result to stdout.

```bash
gcnf inject .env.template > .env
gcnf inject -o .env .env.template
gcnf i .env.template > .env

# Skip comment lines in output
gcnf inject -c .env.template > .env
```

| Flag               | Short | Default  | Description                         |
|--------------------|-------|----------|-------------------------------------|
| `--skip-comments`  | `-c`  | `false`  | Omit comment lines from output      |
| `--output`         | `-o`  | _(none)_ | Write output to file instead of stdout |

**Template syntax example** (`.env.template`):

```bash
ENV_TABLE=UnitTests
ENV_CATEGORY_NAME=Category1
ENV_CATEGORY=$ENV_CATEGORY_NAME
ENV_NAME=unittest
ENV_TEST_NAME=Value2
ENV_VALUE_NAME=$ENV_TEST_NAME
ENV_1=hello
ENV_2=op://Env/$APP_ENV/Database/ConnectionString
ENV_3=gcnf://$ENV_TABLE/${ENV_NAME:?error_env_not_found}/${ENV_CATEGORY}/${ENV_VALUE:-Value1}
# support = comments
ENV_4="gcnf://$ENV_TABLE/$ENV_NAME/$ENV_CATEGORY/$ENV_VALUE_NAME"
```

Templates support:

- Environment variable expansion: `$VAR`, `${VAR}`
- Default values: `${VAR:-default}`
- Required variables: `${VAR:?error message}`
- `gcnf://` URLs for Google Sheets values
- Comment lines starting with `#`

### `gcnf token generate` -- Generate a composite token

Bundle the current service account credentials, sheet ID, and config file path into a single encoded token. This is useful in CI/CD environments where setting one variable is easier than setting several.

```bash
# First set up credentials and sheet ID, then generate:
export GCNF_GOOGLE_CREDENTIAL_BASE64=$(base64 < service_account.json)
export GCNF_GOOGLE_SHEET_ID=YOUR_SHEET_ID
gcnf token generate
```

The output token can then be used as `GCNF_TOKEN` in other environments:

```bash
export GCNF_TOKEN=<generated_token>
```

### `gcnf list` -- List sheets, environments, or categories

Discover what is available in the spreadsheet.

```bash
# List all sheet tabs
gcnf list sheets

# List environment columns in a sheet
gcnf list envs --sheet Env

# List categories in a sheet
gcnf list categories --sheet Env
```

| Subcommand   | Flag      | Short | Required | Description        |
|--------------|-----------|-------|----------|--------------------|
| `sheets`     |           |       |          | No flags needed    |
| `envs`       | `--sheet` | `-s`  | Yes      | Sheet (tab) name   |
| `categories` | `--sheet` | `-s`  | Yes      | Sheet (tab) name   |

### `gcnf diff` -- Compare two environments

Compare configuration values between two environments in the same sheet.

```bash
gcnf diff --sheet Env --env1 staging --env2 production
gcnf diff -s Env --env1 dev --env2 staging
```

| Flag      | Short | Required | Description          |
|-----------|-------|----------|----------------------|
| `--sheet` | `-s`  | Yes      | Sheet (tab) name     |
| `--env1`  |       | Yes      | First environment    |
| `--env2`  |       | Yes      | Second environment   |

### `gcnf validate` -- Validate a template file

Check a template file for missing variables, malformed `gcnf://` URLs, and empty values. Exits with code 0 if valid, 1 if issues found.

```bash
gcnf validate .env.template
```

### `gcnf completion` -- Generate shell completions

Generate shell completion scripts for bash, zsh, fish, or PowerShell.

```bash
# Bash (add to ~/.bashrc)
source <(gcnf completion bash)

# Zsh (add to ~/.zshrc)
source <(gcnf completion zsh)

# Fish
gcnf completion fish | source

# PowerShell
gcnf completion powershell | Invoke-Expression
```

### `gcnf update`

Download and install the latest version of gcnf.

```bash
gcnf update
```

### `gcnf version`

Print the current version.

```bash
gcnf version
```

---

## Global Flags

These flags can be used with any command:

| Flag        | Short | Description                          |
|-------------|-------|--------------------------------------|
| `--quiet`   | `-q`  | Suppress non-data output             |
| `--verbose` |       | Enable verbose diagnostic logging    |
| `--version` | `-v`  | Print version                        |

```bash
# Suppress all warnings/info during inject
gcnf inject --quiet .env.template > .env

# See diagnostic details (cache hits, API calls)
gcnf get --verbose --sheet Env --env staging --category Database --name Host
```

---

## Command Aliases

| Command        | Alias    |
|----------------|----------|
| `gcnf get`     | `gcnf g` |
| `gcnf load`    | `gcnf l` |
| `gcnf read`    | `gcnf r` |
| `gcnf inject`  | `gcnf i` |
| `gcnf unload`  | `gcnf d` |

---

## Environment Variables

**`GCNF_GOOGLE_CREDENTIAL_BASE64`** -- Default: _(none)_

Base64-encoded Google service account JSON key. Used for non-interactive (CI/CD) authentication.

**`GCNF_GOOGLE_SHEET_ID`** -- Default: _(none)_

The Google Sheets document ID. Found in the sheet URL: `https://docs.google.com/spreadsheets/d/<ID>/`

**`GCNF_STORE_CONFIG_FILE`** -- Default: `./gcnf_config.json`

Path to the local config cache file created by `gcnf load`.

**`GCNF_CACHE_TTL`** -- Default: _(none, cache never expires)_

Duration after which cached data expires and is re-fetched from Google Sheets. Uses Go duration format: `30m`, `1h`, `24h`, etc.

```bash
export GCNF_CACHE_TTL=1h
```

**`GCNF_TOKEN`** -- Default: _(none)_

Composite token that encodes credentials, sheet ID, and config file path into a single value. Generated with `gcnf token generate`. When set, it provides defaults for the three variables above. Individual environment variables take precedence over values decoded from `GCNF_TOKEN`.

### Configuration precedence

gcnf resolves configuration in the following order (highest priority first):

1. Individual environment variables (`GCNF_GOOGLE_CREDENTIAL_BASE64`, `GCNF_GOOGLE_SHEET_ID`, `GCNF_STORE_CONFIG_FILE`)
2. Values decoded from `GCNF_TOKEN`
3. Persistent config files stored in `~/.gcnf/`
4. Built-in defaults

---

## CI/CD Usage

### Using individual environment variables

```bash
export GCNF_GOOGLE_CREDENTIAL_BASE64=$(base64 < service_account.json)
export GCNF_GOOGLE_SHEET_ID=1ykOxHza5fxa-HbXPPlSGfTSm2YKG0AteNhDHI6tntUk

gcnf inject .env.template > .env
```

### Using GCNF_TOKEN

```bash
# Generate the token once (locally):
export GCNF_GOOGLE_CREDENTIAL_BASE64=$(base64 < service_account.json)
export GCNF_GOOGLE_SHEET_ID=1ykOxHza5fxa-HbXPPlSGfTSm2YKG0AteNhDHI6tntUk
gcnf token generate
# Copy the output and store it as a CI/CD secret

# In CI/CD, set only one variable:
export GCNF_TOKEN=<token_from_above>
gcnf inject .env.template > .env
```

---

## Stored Files

gcnf stores the following files in `~/.gcnf/`:

| File                         | Purpose                                           |
|------------------------------|---------------------------------------------------|
| `~/.gcnf/.token`             | OAuth2 user token (created by `gcnf login`)       |
| `~/.gcnf/.google_sheet_id`   | Persisted Google Sheet ID (set by `gcnf config`)  |
| `~/.gcnf/.gcnf_config.json`  | Persisted config file path                        |

The local config cache is stored at the path specified by `GCNF_STORE_CONFIG_FILE` (default: `./gcnf_config.json` in the current working directory).
