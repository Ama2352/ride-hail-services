# ride-hail-services (Repo 2 of 3)

> Application source code and CI pipeline for the Ride-Hailing platform.
> Part of a 3-repo GitOps architecture governed by `Global_Principles.md`.

---

## Architecture Position

```
┌─────────────────────┐     ┌──────────────────────┐     ┌─────────────────────┐
│  ride-hail-platform │     │  ride-hail-services  │     │  ride-hail-gitops   │
│      (Repo 1)       │     │  >>>  THIS REPO <<<  │     │      (Repo 3)       │
│                     │     │                      │     │                     │
│  Vagrant, Ansible,  │     │  Go source code,     │     │  K8s manifests,     │
│  K8s bootstrap,     │     │  Dockerfiles,        │     │  Helm values,       │
│  ArgoCD install     │     │  Jenkinsfile (CI)    │     │  ArgoCD App defs    │
└─────────────────────┘     └──────────┬───────────┘     └──────────▲──────────┘
                                       │  git commit image tag      │
                                       └───────────────────────────►┘
                                              ArgoCD reconciles
```

---

## Services

| Service | Port | Description |
|---|---|---|
| `dispatch-service` | `:8080` | Handles ride dispatch and matching |
| `notification-service` | `:8080` | Sends rider/driver notifications |

Both services expose `/metrics` for Prometheus scraping and `/health` for liveness probes.

**Tech stack:** Go 1.25.6, `prometheus/client_golang`, multi-stage Docker build (`golang:1.25.7-alpine` → `alpine:3.20`).

---

## Repository Structure

```
ride-hail-services/
├── dispatch/
│   ├── main.go
# ride-hail-services

> Application source code and GitLab CI pipeline for the Ride-Hailing platform.
> Repo 2 of 3 in a GitOps architecture: **platform** → **services** → **gitops**.

---

## Architecture Overview

```
┌─────────────────────┐     ┌──────────────────────┐     ┌─────────────────────┐
│  ride-hail-platform │     │  ride-hail-services  │     │  ride-hail-gitops   │
│      (Repo 1)       │     │   >>>THIS REPO<<<    │     │      (Repo 3)       │
│                     │     │                      │     │                     │
│  Vagrant · Ansible  │     │  Go source code      │     │  K8s manifests      │
│  K8s bootstrap      │     │  Dockerfiles         │     │  Helm values        │
│  ArgoCD install     │     │  GitLab CI pipeline  │     │  ArgoCD App defs    │
└─────────────────────┘     └──────────┬───────────┘     └──────────▲──────────┘
                                                                  │  git commit (image tag)    │
                                                                  └───────────────────────────►┘
                                                                               ArgoCD reconciles
```

---

## Services

| Service | Port | Description |
|---|---|---|
| `dispatch-service` | `8080` | Ride dispatch and driver-matching logic |
| `notification-service` | `8080` | Rider and driver notification delivery |

Both services expose:

| Endpoint | Purpose |
|---|---|
| `GET /health` | Liveness probe — returns service name, status, and timestamp |
| `GET /metrics` | Prometheus scrape endpoint (`http_requests_total`, `http_request_duration_seconds`) |
| `GET /ride/dispatch` | Business endpoint (dispatch) |
| `GET /notifications` | Business endpoint (notification) |

**Tech stack:** Go 1.25.8 · `prometheus/client_golang v1.23.2` · Multi-stage Docker build (`golang:1.25.8-alpine` → `alpine:3.20`) · Non-root container user (`appuser`, UID 1000)

---

## Repository Structure

```
ride-hail-services/
├── .gitlab-ci.yml               # Active CI pipeline (GitLab CI)
├── Jenkinsfile                  # Legacy — retained for historical reference
├── dispatch/
│   ├── main.go                  # Entry point — server bootstrap only
│   ├── server.go                # HTTP handlers, Prometheus metrics, middleware
│   ├── main_test.go             # Unit tests for all handler paths
│   ├── go.mod / go.sum
│   ├── Dockerfile               # Multi-stage build → alpine:3.20 runtime
│   └── sonar-project.properties # SonarQube: project key uitgo-dispatch-service
└── notification/
       ├── main.go
       ├── server.go
       ├── main_test.go
       ├── go.mod / go.sum
       ├── Dockerfile
       └── sonar-project.properties # SonarQube: project key uitgo-notification-service
```

> **Intentionally absent:** `k8s.yaml`, `istio.yaml` — Kubernetes manifests belong in Repo 3 (Global Principle #2 — Repo Separation).

---

## CI Pipeline (`.gitlab-ci.yml`)

The pipeline has **6 sequential stages** with a `needs:` DAG for parallelism within stages. Every push to a branch or merge request runs the full CI path; image push and GitOps update are gated to branch pipelines and `main` respectively.

```
verify ──► sonar ──► build ──► scan ──► push ──► gitops
```

| Job | Stage | Runs on | Tool |
|---|---|---|---|
| `test_dispatch` | verify | all branches + MRs | `golang:1.25.8-alpine` |
| `test_notification` | verify | all branches + MRs | `golang:1.25.8-alpine` |
| `scan_dependencies` | verify | all branches + MRs | `govulncheck` |
| `sonarqube_analysis` | sonar | all branches + MRs | `sonar-scanner-cli:11.3` |
| `build_images` | build | all branches + MRs | `docker:26-cli` + DinD |
| `scan_images` | scan | all branches + MRs | `aquasec/trivy:0.48.3` |
| `push_images` | push | branch pipelines only | `docker:26-cli` + DinD |
| `gitops_update_dev` | gitops | `main` branch only | `alpine:3.20` |

**Security gates:**
- `govulncheck` — fails the pipeline on known Go module vulnerabilities
- `SonarQube` — `sonar.qualitygate.wait=true` blocks the pipeline if the quality gate fails
- `Trivy` — rejects images with `HIGH` or `CRITICAL` CVEs (`--exit-code 1`)

**Image tag format:** `$CI_PIPELINE_IID-$CI_COMMIT_SHORT_SHA` (e.g. `42-a1b2c3d`)

**Image registry:** `docker.io/ama2352` — both a versioned tag and `latest` are pushed on every merge to `main`.

**Artifact passing across jobs:** Docker images are saved to `.tar` artifacts in the `build` stage (`docker save`) and reloaded in `scan` and `push` stages (`docker load`), since GitLab CI jobs run on ephemeral runners with no shared Docker daemon.

**GitOps handoff (`gitops_update_dev`):** Clones `ride-hail-gitops`, patches `newTag` in both `apps/*/overlays/dev/kustomization.yaml` via `sed`, and pushes the commit. ArgoCD detects the diff and rolls out the new image automatically.

### What the pipeline does NOT do

- No `kubectl` commands — deployment is entirely pull-based (ArgoCD watches Repo 3).
- No direct cluster access — the runner never touches the Kubernetes API.

---

## GitLab CI Variables

Configure the following CI/CD variables in **Settings → CI/CD → Variables**:

| Variable | Type | Purpose |
|---|---|---|
| `DOCKERHUB_USER` | Variable | Docker Hub username for image push |
| `DOCKERHUB_TOKEN` | Masked | Docker Hub access token |
| `SONAR_TOKEN` | Masked | SonarQube authentication token |
| `GITOPS_REPO_URL` | Variable | HTTPS URL of `ride-hail-gitops` (e.g. `https://github.com/ama2352/ride-hail-gitops.git`) |
| `GITOPS_PUSH_TOKEN` | Masked | Personal access token with write access to the GitOps repo |

---

## Local Development

```bash
# Run dispatch service (default port 8080)
cd dispatch && go run .

# Run notification service
cd notification && go run .

# Run tests with coverage
cd dispatch && go test -v -coverprofile=coverage.out ./...
cd notification && go test -v -coverprofile=coverage.out ./...

# Run vulnerability scan locally
go install golang.org/x/vuln/cmd/govulncheck@latest
cd dispatch && govulncheck ./...
```

---

## Key Design Decisions

| Decision | Rationale |
|---|---|
| GitLab CI with Docker-in-Docker (DinD) | Ephemeral runners have no shared daemon; DinD provides an isolated Docker environment per job |
| Artifact-based image passing | `docker save` → artifact → `docker load` bridges the build, scan, and push jobs across ephemeral runners |
| `needs:` DAG for parallelism | `test_dispatch` and `test_notification` run in parallel; `sonarqube_analysis` waits for both coverage reports |
| Quality gates (govulncheck + Trivy + SonarQube) | Defense in depth — Go module CVEs, container image CVEs, and code quality are all hard gates |
| `push_images` gated to branches only | MR pipelines validate but never push images; prevents tag pollution from draft MRs |
| `gitops_update_dev` gated to `main` | Only merged, reviewed code triggers a deployment to dev |
| Multi-stage Docker build | `golang:alpine` build layer → `alpine` runtime layer; strips toolchain from the final image |
| Non-root container user | `appuser` (UID 1000) reduces container breakout risk |
| No `kubectl` in pipeline | Pull-based CD — the CI runner never touches the cluster (Global Principle #3) |

---

## Global Principles

1. **Declarative** — Every cluster state is described in Git. No manual `kubectl` or ad-hoc shell for final state.
2. **Repo Separation** — Each repository owns exactly one concern: infrastructure, application code, or desired state.
3. **Pull-Based CD** — GitLab CI pushes images; ArgoCD pulls manifests from Repo 3.
4. **Folders over Branches** — Environment differences live in directory overlays in Repo 3, not git branches.
