# demo-service-008

End-to-end documentation verification

> Bootstrapped by the **repo-creator** platform service.

This repository ships a runnable Go HTTP service with `/healthz`, `/readyz`, and
`/metrics` endpoints, structured JSON logging, a passing test suite, a
multi-stage distroless container build, branch-protected CI (`go vet`,
`staticcheck`, `govulncheck`, `go test`), Dependabot, and Kubernetes manifests
ready for GitOps.

---

## Quick start (run locally in 60 seconds)

```bash
# 1. Clone
git clone https://github.com/MichaelF21/demo-service-008.git
cd demo-service-008

# 2. Hydrate go.sum from go.mod (we don't commit go.sum at bootstrap time)
go mod tidy

# 3. Run the tests (sanity-check before running)
go test ./...

# 4. Start the server
go run .
```

In another terminal:

```bash
curl http://localhost:8080/
# {"service":"demo-service-008","message":"Hello from demo-service-008!"}

curl http://localhost:8080/healthz
# ok

curl http://localhost:8080/readyz
# ready

curl -s http://localhost:8080/metrics | grep demo_service_008
# demo_service_008_http_requests_total{path="/",status="200"} 1
```

To run on a different port: `PORT=8090 go run .`

---

## Build and run a container

```bash
# Multi-stage build: builder = golang:1.23-alpine, runtime = distroless
docker build -t demo-service-008:latest .

# Final image is ~17 MB (distroless + statically-linked Go binary)
docker images demo-service-008:latest

# Run it (map host 8080 -> container 8080)
docker run --rm -p 8080:8080 demo-service-008:latest

# In another terminal:
curl http://localhost:8080/
```

The container runs as the distroless `nonroot` user (UID 65532) on a read-only
root filesystem with all Linux capabilities dropped — see `deploy/k8s/deployment.yaml`
for the full security context.

---

## Deploy to Kubernetes

The manifests under `deploy/k8s/` declare a `Deployment` + `Service`, pre-wired
with Prometheus scrape annotations, liveness/readiness probes pointing at our
endpoints, and the same security posture as the container.

```bash
# 1. Build the image (skip if you already did the section above)
docker build -t demo-service-008:latest .
# Docker Desktop's Kubernetes shares the daemon, so the image is available;
# for kind/k3d you'd `kind load docker-image demo-service-008:latest` here.

# 2. Apply manifests into a namespace
kubectl create namespace demo-service-008
kubectl -n demo-service-008 apply -f deploy/k8s/

# 3. Wait for the rollout
kubectl -n demo-service-008 rollout status deploy/demo-service-008

# 4. Port-forward and hit it
kubectl -n demo-service-008 port-forward svc/demo-service-008 8080:80
# In another terminal:
curl http://localhost:8080/
```

To uninstall: `kubectl delete namespace demo-service-008`.

### GitOps

`deploy/argocd-application.yaml` is an ArgoCD `Application` template pointing at
this repo's `deploy/k8s/` directory. Commit it into your GitOps repo (or apply
it directly: `kubectl apply -f deploy/argocd-application.yaml`) and ArgoCD will
keep the cluster in sync with `main`.

---

## Endpoints

| Path       | Purpose                                  |
|------------|------------------------------------------|
| `/`        | JSON greeting (the demo functional path) |
| `/healthz` | Liveness probe — always 200 if alive     |
| `/readyz`  | Readiness probe                          |
| `/metrics` | Prometheus exposition                    |

---

## CI and branch protection

`.github/workflows/ci.yml` runs on every pull request to `main` and every push
to `main`:

1. `go vet ./...`
2. `staticcheck` (via the SHA-pinned `dominikh/staticcheck-action`)
3. `govulncheck ./...` — fails the build if any imported function is on the
   Go vulnerability database
4. `go test ./...`

Branch protection on `main` requires the `lint-and-test` status check to pass
and at least one approving review before merging. Force-pushes are blocked.
GitHub Actions are pinned by full commit SHA (with version-tagged comments)
to prevent supply-chain attacks via tag rewriting.

`.github/dependabot.yml` opens weekly PRs to bump `go.mod`, `github-actions`,
and `docker` deps. Each PR runs the full CI suite before it's mergeable.

---

## Observability

- **Logs.** Structured JSON via `slog` to stdout. Pipe through your log
  aggregator; every line has `time`, `level`, `msg`, plus structured fields
  for the event.
- **Metrics.** Prometheus exposition at `/metrics`. The pod-template
  annotations in `deploy/k8s/deployment.yaml` (`prometheus.io/scrape: "true"`)
  let a default-config Prometheus auto-discover this service. Custom counter:
  `demo_service_008_http_requests_total{path,status}`.
- **Health.** `/healthz` is liveness (restart on failure); `/readyz` is
  readiness (pull from load balancer on failure). Both are wired into the
  Deployment's probes.

---

## Configuration

| Env var | Default | Purpose                          |
|---------|---------|----------------------------------|
| `PORT`  | `8080`  | TCP port the HTTP server binds to |

---

## Layout

```
demo-service-008/
├── main.go                         # entrypoint + handlers
├── main_test.go                    # handler unit tests
├── go.mod
├── Dockerfile                      # multi-stage, distroless
├── .gitignore
├── SECURITY.md
├── .github/
│   ├── workflows/ci.yml            # lint-and-test workflow
│   └── dependabot.yml
└── deploy/
    ├── k8s/
    │   ├── deployment.yaml
    │   └── service.yaml
    └── argocd-application.yaml
```
