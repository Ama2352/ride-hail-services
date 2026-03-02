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
│  K8s bootstrap,     │     │  Dockerfiles,        │     │  Istio configs,     │
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

## Repository Layout

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
├── jenkins/
│   └── email/
│       ├── success.txt
│       └── failure.txt
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
