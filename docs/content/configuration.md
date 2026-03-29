# Configuration

All configuration is done through environment variables. Copy `.env.example` to `.env` and edit it, or set the variables directly.

```bash
cp .env.example .env
```

## Application

| Variable | Default | Required | Description |
|---|---|---|---|
| `PORT` | `3000` | No | Port the HTTP server listens on |
| `DEV_MODE` | `false` | No | Enable development mode with mock GitHub, mock workspace, mock engine, and auth bypass. See [Testing](testing.md) for details |
| `DEV_REPO_DIR` | `backend/testdata/fake-infra` | No | Local directory used as the fake repository in dev mode. The mock workspace symlinks to this path instead of cloning |
| `REPOS` | _(empty)_ | No | Comma-separated list of repositories in `owner/name` format (e.g., `myorg/infra,myorg/platform`). These are the repos shown in the UI |
| `SESSION_SECRET` | `dev-secret-change-in-prod` | **Yes (in prod)** | Secret key used to HMAC-sign session cookies. Must be changed from the default in production |
| `LOG_LEVEL` | `info` | No | Log verbosity. Accepted values: `debug`, `info`, `warn`, `error` |

## GitHub OAuth and API

These variables are **required** when `DEV_MODE` is not `true`. In dev mode they are ignored.

| Variable | Default | Required | Description |
|---|---|---|---|
| `GITHUB_CLIENT_ID` | _(none)_ | **Yes** | OAuth App client ID. Create one at [github.com/settings/developers](https://github.com/settings/developers) |
| `GITHUB_CLIENT_SECRET` | _(none)_ | **Yes** | OAuth App client secret |
| `GITHUB_TOKEN` | _(none)_ | No | Personal access token with repo read access. Used by the backend to clone private repos and call the GitHub API for PR data |
| `ALLOWED_ORG` | _(none)_ | **Yes** | GitHub organisation slug. Only members of this org can log in |
| `BASE_URL` | `http://localhost:3000` | No | Public URL of the kiln instance. Used to construct the OAuth callback URL (`<BASE_URL>/auth/callback`). Must match the callback URL configured in your GitHub OAuth App |

## Cloud Provider Credentials

These variables are passed through to Terraform and Pulumi at runtime. Set whichever ones your infrastructure requires.

### AWS

| Variable | Default | Description |
|---|---|---|
| `AWS_ACCESS_KEY_ID` | _(none)_ | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | _(none)_ | AWS secret key |
| `AWS_DEFAULT_REGION` | _(none)_ | AWS region (e.g., `us-east-1`) |

### GCP

| Variable | Default | Description |
|---|---|---|
| `GOOGLE_APPLICATION_CREDENTIALS` | _(none)_ | Path to a GCP service account JSON key file |

### Pulumi

| Variable | Default | Description |
|---|---|---|
| `PULUMI_ACCESS_TOKEN` | _(none)_ | Pulumi Cloud access token for state management |

## Docker Compose

The `docker-compose.yml` passes all of the above variables through to the container. Variables with `:-` defaults in the compose file will use empty strings if not set. See the compose files for the full mapping:

- `docker-compose.yml` -- Production configuration
- `docker-compose.dev.yml` -- Development configuration with source mounting and optional ngrok
