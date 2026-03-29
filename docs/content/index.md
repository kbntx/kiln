# Kiln

Kiln is a self-hosted Infrastructure as Code (IaC) runner with a web UI. It lets teams authenticate via GitHub OAuth, configure repositories, browse open pull requests, discover Pulumi and Terraform projects within those repos, and run `plan` or `apply` operations with real-time log streaming -- all from a single binary with an embedded frontend.

Kiln is designed for teams that want lightweight, self-hosted IaC automation without the overhead of a full CI/CD platform. It ships as one Docker image, requires no database, and keeps all state in memory.

## Quick Start

```bash
cp .env.example .env
# Edit .env with your GitHub OAuth credentials, token, and allowed org

docker compose up
```

The UI is available at [http://localhost:3000](http://localhost:3000).

For development mode (no GitHub credentials needed):

```bash
tilt up
```

## Features

- **GitHub OAuth authentication** with organisation-based access control
- **Repository management** -- configure which repos kiln monitors via a simple env var
- **Pull request browsing** -- list open PRs and their approval status
- **Automatic project discovery** -- detects Pulumi (`Pulumi.yaml`) and Terraform (`*.tf`) projects in a repo
- **Plan and Apply** -- run IaC operations against discovered projects and stacks
- **Real-time log streaming** -- SSE-based streaming of plan/apply output to the browser
- **Single binary deployment** -- Go backend with the React frontend embedded at build time
- **No database** -- all run state is held in memory; kiln is stateless across restarts
- **Dev mode** -- full mock stack (GitHub API, git workspace, IaC engine) for local development without any external dependencies

## Documentation

| Document | Description |
|---|---|
| [Architecture](architecture.md) | System design, package map, and Mermaid diagrams |
| [Configuration](configuration.md) | All environment variables |
| [Testing](testing.md) | Local development, dev mode, and testing strategy |
