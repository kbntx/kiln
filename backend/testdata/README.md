# Test Data

This directory contains fixture data used by both **unit tests** and **dev mode**.

## `fake-infra/`

A minimal directory tree that mimics a real repository containing IaC projects. It is intentionally small — just enough to exercise kiln's discovery logic.

### Structure

```
fake-infra/
├── Pulumi.yaml          # Pulumi project definition (name: fake-pulumi-project, runtime: go)
├── Pulumi.dev.yaml      # Stack config for "dev" stack
├── Pulumi.prod.yaml     # Stack config for "prod" stack
└── terraform/
    ├── main.tf           # Minimal Terraform config (null_resource)
    └── variables.tf      # Single variable with a default
```

### How discovery uses it

Kiln's `internal/discovery` package walks a directory looking for IaC projects:

1. **Pulumi**: finds `Pulumi.yaml`, reads the project `name`, then globs `Pulumi.*.yaml` to extract stack names (`dev`, `prod`).
2. **Terraform**: finds directories containing `*.tf` files, uses the directory basename as the project name, returns `["default"]` as stacks.

Running `DiscoverProjects("backend/testdata/fake-infra")` returns two projects:

| Name | Engine | Dir | Stacks |
|------|--------|-----|--------|
| fake-pulumi-project | pulumi | `.` | dev, prod |
| terraform | terraform | `terraform` | default |

### How dev mode uses it

When `DEV_MODE=true`, the `MockWorkspace` symlinks this directory instead of cloning a real repository. When a user clicks "Discover Projects" in the UI, kiln runs discovery against this tree and returns the two projects above.

You can override this with the `DEV_REPO_DIR` env var to point at your own local IaC repo instead.

### How tests use it

`internal/discovery/discover_test.go` runs `DiscoverProjects` against this directory and asserts the expected projects, engines, and stacks are found.

### Adding test fixtures

To test new discovery logic (e.g. a new engine type or nested project detection), add files here and update the corresponding test in `internal/discovery/discover_test.go`.
