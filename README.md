# ride-hail-services (Repo 2 of 3)

> Application source code and CI pipelines for the Ride-Hailing platform.

---

## 🔗 Related Repositories
- [ride-hail-platform](https://github.com/ama2352/ride-hail-platform) (Repo 1 - Infrastructure)
- **ride-hail-services** (Repo 2 - You are here)
- [ride-hail-gitops](https://github.com/ama2352/ride-hail-gitops) (Repo 3 - K8s Manifests & App Config)

---

## 🏛️ Architecture Overview

This repository defines the core business logic (Golang) and the Continuous Integration models that integrate natively into our GitOps flow.

```text
┌─────────────────────┐     ┌──────────────────────┐     ┌─────────────────────┐
│  ride-hail-platform │     │  ride-hail-services  │     │  ride-hail-gitops   │
│      (Repo 1)       │     │  >>> THIS REPO <<<   │     │      (Repo 3)       │
│                     │     │                      │     │                     │
│  Vagrant, Ansible,  │     │  Go source code,     │     │  K8s manifests,     │
│  K8s bootstrap,     │     │  Dockerfiles,        │     │  Helm values,       │
│  ArgoCD install     │     │  Jenkins & GitLab CI │     │  ArgoCD App defs    │
└─────────────────────┘     └──────────┬───────────┘     └──────────▲──────────┘
                                       │  git commit image tag      │
                                       └───────────────────────────►┘
                                              ArgoCD reconciles
```

---

## 🚀 Dual-CI Workflows

This repo contains configurations to execute fully compatible CI pipelines across GitLab or Jenkins. Below is the behavioral model:

### Jenkins CI (`Jenkinsfile`)
- **Pipeline:** Native parallel stages orchestrating build, test, and docker publishing via the Jenkins master node.
- **GitOps Handoff:** Dynamically mutates `apps/dispatch/overlays/dev/kustomization.yaml` inside `ride-hail-gitops` to record structural container versions.
- **Trigger:** Automated via GitHub webhooks or periodic triggers.

### GitLab CI (`.gitlab-ci.yml`)
- **Pipeline:** Implements a strict `verify -> sonar -> build -> scan -> push -> gitops` process using localized Docker-in-Docker runners.
- **Security Gates:** Ensures code is vetted fully against `govulncheck`, `SonarQube` standards, and `Trivy` container scanning prior to the push.
- **GitOps Handoff:** A dedicated job resolves modifications within Repo 3 (`gitops_update_dev`) on the main branch, automating delivery securely. 

*Crucially: CI pipelines never apply changes directly into Kubernetes (no `kubectl apply`). They handoff strictly to ArgoCD using Repo 3.*

---

## ⚙️ Setup Guide (Fresh Environment)

### Local Development:
The services (built with Go 1.25.x) run natively standard `go run` models:
```bash
# Start dispatch service locally
cd dispatch && go run .

# Start notification service locally 
cd notification && go run .
```

### End-to-End Delivery:
1. Validate infrastructure is running (from Repo 1).
2. Create/Commit changes here in `ride-hail-services`.
3. Monitor your active CI runtime (Jenkins or GitLab Runner).
4. ArgoCD orchestrates deployment inside the cluster as soon as the pipelines push the updated tag to `ride-hail-gitops`.
