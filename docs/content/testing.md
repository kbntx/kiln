# Testing and Local Development

## Dev Mode

Setting `DEV_MODE=true` activates a full mock stack that replaces all external dependencies. This lets you run kiln locally without a GitHub OAuth app, a GitHub token, or any real infrastructure.

When dev mode is enabled:

| Component | Production | Dev Mode |
|---|---|---|
| GitHub API | `RealClient` (go-github, requires `GITHUB_TOKEN`) | `MockGitHubClient` -- returns 5 fake PRs per repo, always-true org membership |
| Git workspace | `RealWorkspace` -- shallow clones into `/tmp/kiln` | `MockWorkspace` -- symlinks to local `testdata/fake-infra` (or `DEV_REPO_DIR`) |
| IaC engine | `TerraformEngine` / `PulumiEngine` (real CLI) | `MockEngine` -- streams realistic Terraform-like output with ANSI colours |
| Authentication | Full GitHub OAuth flow with org check | `/auth/login` immediately creates a `dev-user` session, no redirect to GitHub |
| Required env vars | `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `ALLOWED_ORG` | None required |

## Running Locally with Docker Compose

The dev compose file mounts the source tree and runs `go run` directly:

```bash
docker compose -f docker-compose.dev.yml up
```

This starts the backend on port 3000 with `DEV_MODE=true`. The frontend dev server (Vite) is available on port 5173 if started separately.

## Running with Tilt

The recommended way to develop locally is with [Tilt](https://tilt.dev):

```bash
tilt up
```

This starts both the backend (`go run` with `DEV_MODE=true`) and the frontend Vite dev server with hot reload. Tilt watches file changes and restarts the relevant service automatically.

The Vite dev server proxies API requests to the backend on port 3000.

## Testing Real OAuth Locally

GitHub OAuth apps accept `http://localhost:3000/auth/callback` as a valid redirect URI. To test the real OAuth flow:

1. Create a GitHub OAuth App at [https://github.com/settings/developers](https://github.com/settings/developers).
   - Set the callback URL to `http://localhost:3000/auth/callback`.

2. Set the required environment variables:

   ```bash
   export DEV_MODE=false
   export GITHUB_CLIENT_ID=your-client-id
   export GITHUB_CLIENT_SECRET=your-client-secret
   export GITHUB_TOKEN=ghp_your-token
   export ALLOWED_ORG=your-org
   export BASE_URL=http://localhost:3000
   ```

3. Run the backend normally — OAuth will redirect through GitHub and back to localhost.

## Unit Tests

Run all tests:

```bash
cd backend && go test ./...
```

### Test structure

Tests live alongside their source files following standard Go conventions:

| Package | File | What it tests |
|---|---|---|
| `internal/config` | `config_test.go` | Repo string parsing, dev mode loading, required-var validation |
| `internal/discovery` | `discover_test.go` | End-to-end discovery against `testdata/fake-infra` (both Pulumi and Terraform) |

### Test data

The `backend/testdata/fake-infra/` directory contains a minimal Pulumi project (with `dev` and `prod` stacks) and a Terraform directory. It is used by both the discovery tests and the dev mode mock workspace.

### Adding tests

- Place tests in the same package as the code under test (`_test.go` suffix).
- Use the `testdata/` directory for fixture files.
- Interfaces (`GitHubClient`, `WorkspaceManager`, `Engine`) are designed to be swapped with mocks. The `devmode` package provides ready-made mock implementations that can be reused in tests.
