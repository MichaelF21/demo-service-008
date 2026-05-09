# Security Policy

## Reporting a vulnerability

Please report security issues privately to the repository owner via GitHub Security Advisories
(`Security → Advisories → Report a vulnerability`). Do not file a public issue for
security-sensitive reports.

## Scope

- The `demo-service-008` service binary and its container image.
- Generated Kubernetes manifests under `deploy/k8s/`.

## Defaults

- Container image is built from a distroless base and runs as non-root.
- Dependencies are tracked and updated via Dependabot (`.github/dependabot.yml`).
- Branch protection on `main` requires PR review and a passing `lint-and-test` check.
