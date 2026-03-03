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
│   ├── server.go
│   ├── main_test.go
│   ├── go.mod / go.sum
│   ├── Dockerfile
│   └── sonar-project.properties
├── notification/
│   ├── main.go
│   ├── server.go
│   ├── main_test.go
│   ├── go.mod / go.sum
│   ├── Dockerfile
│   └── sonar-project.properties
├── Jenkinsfile
├── .gitignore
└── README.md
```

**Intentionally absent:** `k8s.yaml`, `istio.yaml` — these belong in Repo 3 (Global Principle #2).

---

## CI Pipeline

The `Jenkinsfile` defines a CI-only pipeline. Jenkins runs on the `jenkins-vm` (192.168.242.13:8080) using Docker-outside-of-Docker (DooD) agents.

| Stage | What it does | Tool |
|---|---|---|
| **Checkout** | Clones this repo, computes `IMAGE_TAG` (`BUILD_NUMBER-GIT_SHORT`) | Git |
| **Test Dispatch** | `go vet` + `go test -coverprofile` | Go 1.25.7 |
| **Test Notification** | Same as above | Go 1.25.7 |
| **Scan Dependencies** | `govulncheck ./...` on both modules | govulncheck |
| **SonarQube Analysis** | Static analysis via `sonar-scanner` | SonarQube (30090) |
| **Build Images** | `docker build` + `docker save` to tar | Docker 26 |
| **Scan Images** | HIGH/CRITICAL CVE gate | Trivy 0.48.3 |
| **Push Images** | Tag + push to `docker.io/ama2352` | Docker 26 |
| **GitOps Update** | Clone Repo 3, `sed` image tags, `git commit` + `git push` | Git |

### What the pipeline does NOT do

- **No `kubectl` commands** — deployment is pull-based (Global Principle #3).
- **No cluster access** — the pipeline never touches the Kubernetes API.
- Deployment is handled entirely by ArgoCD watching Repo 3.

---

## Credentials Required in Jenkins

| ID | Type | Purpose |
|---|---|---|
| `docker-registry-credentials` | usernamePassword | Push images to Docker Hub |
| `sonarqube-token` | secret text | Authenticate with SonarQube |
| `gitops-repo-credentials` | usernamePassword | Push image-tag commits to Repo 3 |

Slack notifications are configured globally in Jenkins (Slack plugin).
Security gate failures are reported per stage; success/failure summaries post automatically.

---

## Local Development

```bash
# Run dispatch service
cd dispatch
go run .

# Run notification service
cd notification
go run .

# Run all tests
cd dispatch && go test -v ./...
cd ../notification && go test -v ./...
```

---

## Key Design Decisions

| Decision | Rationale |
|---|---|
| DooD for CI agents | Sibling containers share the Docker daemon; avoids DinD privilege risks |
| Per-stage Slack notifications | Security gate failures reported immediately, not just at pipeline end |
| `when { branch 'main' }` on GitOps Update | PRs run CI only; deployment triggers only on merges to main |
| Multi-stage Docker build | `golang:alpine` build → `alpine` runtime; minimal attack surface |
| No `kubectl` in pipeline | Pull-based CD — Jenkins never touches the cluster (Principle #3) |
| govulncheck + Trivy + SonarQube | Defense in depth — dependencies, images, and code quality all gated |

---

## Global Principles

1. **Declarative** — Every state is described in Git. No manual `kubectl` or ad-hoc `sh` for final cluster state.
2. **Repo Separation** — Each repo owns a single concern: infrastructure, code, or desired state.
3. **Pull-Based CD** — Jenkins pushes images; ArgoCD pulls manifests from Repo 3.
4. **Folders > Branches** — Environment differences are directory overlays in Repo 3.
