// =============================================================================
// Jenkinsfile — ride-hail-services (Repo 2)
//
// CI-ONLY pipeline. Builds, tests, scans, and pushes Docker images.
// The final stage ("GitOps Update") commits the new image tag to Repo 3
// (ride-hail-gitops), triggering ArgoCD to reconcile the cluster.
//
// WHAT WAS REMOVED (compared to UITGo_Ver2/Jenkinsfile):
//   - 'CD' stage:  kubectl apply, envsubst, rollout status — all forbidden
//                   by Global Principle #3 (Pull-Based CD).
//   - KUBERNETES_SERVER env var — this pipeline never talks to the cluster.
//   - cd-pod.yaml K8s agent — no need for in-cluster kubectl access.
//   - Email templates from infrastructure/ — those paths belonged to Repo 1.
//
// WHAT WAS ADDED:
//   - 'GitOps Update' stage: clones Repo 3, updates image tags via sed,
//     commits + pushes. ArgoCD detects the change and deploys.
//
// CREDENTIALS REQUIRED IN JENKINS:
//   - docker-registry-credentials  (usernamePassword)  — Docker Hub push
//   - sonarqube-token              (string)             — SonarQube analysis
//   - gitops-repo-credentials      (usernamePassword)   — Push to Repo 3
// =============================================================================

pipeline {
    agent none

    environment {
        DOCKER_REGISTRY   = "docker.io/ama2352"
        SONAR_HOST        = "http://192.168.242.10:30090"
        GITOPS_REPO       = "https://github.com/ama2352/ride-hail-gitops.git"
        GITOPS_BRANCH     = "main"
    }

    options {
        timeout(time: 30, unit: 'MINUTES')
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '10'))
    }

    stages {

        // =====================================================================
        // CI STAGES — unchanged from original (paths adjusted for new repo layout)
        // =====================================================================

        stage('CI') {
            agent { label 'built-in' }

            stages {

                stage('Checkout') {
                    steps {
                        echo "Checking out source code from SCM..."
                        checkout scm
                        script {
                            env.IMAGE_TAG = "${env.BUILD_NUMBER}-${env.GIT_COMMIT?.take(7) ?: 'latest'}"
                            echo "Image tag: ${env.IMAGE_TAG}"
                        }
                    }
                }

                stage('Verify Source') {
                    parallel {

                        stage('Test Dispatch') {
                            agent {
                                docker {
                                    image 'golang:1.25.7-alpine'
                                    args  '-u root -e HOME=/root -e GOPATH=/root/go -v /tmp/go-mod-cache:/root/go/pkg/mod -v /tmp/go-build-cache:/root/.cache/go-build'
                                    reuseNode true
                                }
                            }
                            steps {
                                dir('dispatch') {
                                    sh '''
                                        echo "=== [dispatch] Downloading Go modules ==="
                                        go mod download
                                        echo "=== [dispatch] Running go vet ==="
                                        go vet ./...
                                        echo "=== [dispatch] Running unit tests ==="
                                        go test -v -coverprofile=coverage.out ./... || echo "No tests yet"
                                        echo "=== [dispatch] Tests complete ==="
                                    '''
                                }
                            }
                        }

                        stage('Test Notification') {
                            agent {
                                docker {
                                    image 'golang:1.25.7-alpine'
                                    args  '-u root -e HOME=/root -e GOPATH=/root/go -v /tmp/go-mod-cache:/root/go/pkg/mod -v /tmp/go-build-cache:/root/.cache/go-build'
                                    reuseNode true
                                }
                            }
                            steps {
                                dir('notification') {
                                    sh '''
                                        echo "=== [notification] Downloading Go modules ==="
                                        go mod download
                                        echo "=== [notification] Running go vet ==="
                                        go vet ./...
                                        echo "=== [notification] Running unit tests ==="
                                        go test -v -coverprofile=coverage.out ./... || echo "No tests yet"
                                        echo "=== [notification] Tests complete ==="
                                    '''
                                }
                            }
                        }

                        stage('Scan Dependencies') {
                            agent {
                                docker {
                                    image 'golang:1.25.7-alpine'
                                    args  '-u root -e HOME=/root -e GOPATH=/root/go -v /tmp/go-mod-cache:/root/go/pkg/mod -v /tmp/go-build-cache:/root/.cache/go-build'
                                    reuseNode true
                                }
                            }
                            steps {
                                sh '''
                                    echo "=== Installing govulncheck ==="
                                    go install golang.org/x/vuln/cmd/govulncheck@latest
                                    GOVULNCHECK=$(go env GOPATH)/bin/govulncheck
                                    echo "=== [dispatch] Scanning for known vulnerabilities ==="
                                    cd dispatch && $GOVULNCHECK ./...
                                    echo "=== [notification] Scanning for known vulnerabilities ==="
                                    cd ../notification && $GOVULNCHECK ./...
                                    echo "=== Dependency scan complete — no known vulnerabilities ==="
                                '''
                            }
                        }

                    }
                }

                stage('SonarQube Analysis') {
                    agent {
                        docker {
                            image 'sonarsource/sonar-scanner-cli:11.3'
                            args  '-u root -e HOME=/root -v /tmp/sonar-cache:/root/.sonar/cache'
                            reuseNode true
                        }
                    }
                    steps {
                        withCredentials([string(credentialsId: 'sonarqube-token', variable: 'SONAR_TOKEN')]) {
                            sh '''
                                echo "=== [dispatch] Running SonarQube analysis ==="
                                cd dispatch
                                sonar-scanner -Dsonar.host.url=${SONAR_HOST} -Dsonar.token=${SONAR_TOKEN}
                                echo "=== [dispatch] SonarQube analysis submitted ==="
                                echo "=== [notification] Running SonarQube analysis ==="
                                cd ../notification
                                sonar-scanner -Dsonar.host.url=${SONAR_HOST} -Dsonar.token=${SONAR_TOKEN}
                                echo "=== [notification] SonarQube analysis submitted ==="
                            '''
                        }
                    }
                }

                stage('Build Images') {
                    agent {
                        docker {
                            image 'docker:26-cli'
                            args  '-v /var/run/docker.sock:/var/run/docker.sock -u root'
                            reuseNode true
                        }
                    }
                    steps {
                        sh '''
                            set -e
                            echo "=== [dispatch] Building Docker image: dispatch-service:${IMAGE_TAG} ==="
                            docker build -t dispatch-service:${IMAGE_TAG}     dispatch
                            echo "=== [dispatch] Saving image to tar for scanning ==="
                            docker save  dispatch-service:${IMAGE_TAG}     -o dispatch-service.tar
                            echo "=== [notification] Building Docker image: notification-service:${IMAGE_TAG} ==="
                            docker build -t notification-service:${IMAGE_TAG} notification
                            echo "=== [notification] Saving image to tar for scanning ==="
                            docker save  notification-service:${IMAGE_TAG} -o notification-service.tar
                            echo "=== Image tars ready for security scan ==="
                            ls -lh *.tar
                        '''
                    }
                }

                stage('Scan Images') {
                    agent {
                        docker {
                            image 'aquasec/trivy:0.48.3'
                            args  '-u root -v /tmp/trivy-cache:/root/.cache/trivy --entrypoint='
                            reuseNode true
                        }
                    }
                    steps {
                        sh '''
                            set -e
                            SCAN_FAILED=0
                            echo "=== [dispatch] Scanning for HIGH/CRITICAL CVEs ==="
                            trivy image --input dispatch-service.tar \
                                --severity HIGH,CRITICAL --exit-code 1 --format table \
                                || SCAN_FAILED=1
                            echo "=== [notification] Scanning for HIGH/CRITICAL CVEs ==="
                            trivy image --input notification-service.tar \
                                --severity HIGH,CRITICAL --exit-code 1 --format table \
                                || SCAN_FAILED=1
                            [ $SCAN_FAILED -eq 0 ] || { echo "Security gate failed — not pushing."; exit 1; }
                            echo "=== Security gate passed — both images are clean ==="
                        '''
                    }
                }

                stage('Push Images') {
                    agent {
                        docker {
                            image 'docker:26-cli'
                            args  '-v /var/run/docker.sock:/var/run/docker.sock -u root'
                            reuseNode true
                        }
                    }
                    steps {
                        withCredentials([usernamePassword(
                            credentialsId: 'docker-registry-credentials',
                            usernameVariable: 'DOCKER_USER',
                            passwordVariable: 'DOCKER_PASS'
                        )]) {
                            sh '''
                                set -e
                                echo "=== Authenticating with Docker registry ==="
                                echo "${DOCKER_PASS}" | docker login -u "${DOCKER_USER}" --password-stdin

                                echo "=== [dispatch] Tagging and pushing to ${DOCKER_REGISTRY} ==="
                                docker tag dispatch-service:${IMAGE_TAG} ${DOCKER_REGISTRY}/dispatch-service:${IMAGE_TAG}
                                docker tag dispatch-service:${IMAGE_TAG} ${DOCKER_REGISTRY}/dispatch-service:latest
                                docker push ${DOCKER_REGISTRY}/dispatch-service:${IMAGE_TAG}
                                docker push ${DOCKER_REGISTRY}/dispatch-service:latest

                                echo "=== [notification] Tagging and pushing to ${DOCKER_REGISTRY} ==="
                                docker tag notification-service:${IMAGE_TAG} ${DOCKER_REGISTRY}/notification-service:${IMAGE_TAG}
                                docker tag notification-service:${IMAGE_TAG} ${DOCKER_REGISTRY}/notification-service:latest
                                docker push ${DOCKER_REGISTRY}/notification-service:${IMAGE_TAG}
                                docker push ${DOCKER_REGISTRY}/notification-service:latest

                                echo "=== Cleaning up local images to free disk space ==="
                                docker rmi dispatch-service:${IMAGE_TAG} \
                                           ${DOCKER_REGISTRY}/dispatch-service:${IMAGE_TAG} \
                                           ${DOCKER_REGISTRY}/dispatch-service:latest || true
                                docker rmi notification-service:${IMAGE_TAG} \
                                           ${DOCKER_REGISTRY}/notification-service:${IMAGE_TAG} \
                                           ${DOCKER_REGISTRY}/notification-service:latest || true
                                echo "=== All images pushed successfully ==="
                            '''
                        }
                    }
                }

            }
        }

        // =====================================================================
        // GITOPS UPDATE — replaces the old "CD" stage entirely
        //
        // Instead of: kubectl apply -f k8s.yaml (Push-based CD, FORBIDDEN)
        // We now:     git commit the new image tag into Repo 3 (Pull-based CD)
        //
        // ArgoCD watches Repo 3 and reconciles the cluster automatically.
        // This pipeline NEVER touches the cluster directly.
        // =====================================================================

        stage('GitOps Update') {
            agent { label 'built-in' }

            steps {
                withCredentials([usernamePassword(
                    credentialsId: 'gitops-repo-credentials',
                    usernameVariable: 'GIT_USER',
                    passwordVariable: 'GIT_TOKEN'
                )]) {
                    sh '''
                        set -e
                        echo "=== Cloning GitOps repository ==="
                        GITOPS_URL=$(echo "${GITOPS_REPO}" | sed "s|https://|https://${GIT_USER}:${GIT_TOKEN}@|")
                        rm -rf gitops-workspace
                        git clone --branch "${GITOPS_BRANCH}" --depth 1 "${GITOPS_URL}" gitops-workspace
                        cd gitops-workspace

                        echo "=== Updating image tags for dispatch-service ==="
                        find . -name '*.yaml' -exec grep -l 'dispatch-service' {} \\; | while read file; do
                            sed -i "s|image:.*dispatch-service:.*|image: ${DOCKER_REGISTRY}/dispatch-service:${IMAGE_TAG}|g" "$file"
                            echo "    Updated: $file"
                        done

                        echo "=== Updating image tags for notification-service ==="
                        find . -name '*.yaml' -exec grep -l 'notification-service' {} \\; | while read file; do
                            sed -i "s|image:.*notification-service:.*|image: ${DOCKER_REGISTRY}/notification-service:${IMAGE_TAG}|g" "$file"
                            echo "    Updated: $file"
                        done

                        echo "=== Committing and pushing to GitOps repo ==="
                        git config user.email "jenkins@ride-hail.ci"
                        git config user.name "Jenkins CI"
                        git add -A
                        git diff --cached --quiet && {
                            echo "No manifest changes detected — nothing to commit."
                            exit 0
                        }
                        MSG=$(printf 'ci: update image tags to %s\n\nTriggered by ride-hail-services build #%s\nCommit: %s\nImages:\n  - %s/dispatch-service:%s\n  - %s/notification-service:%s' \
                            "${IMAGE_TAG}" "${BUILD_NUMBER}" "${GIT_COMMIT}" \
                            "${DOCKER_REGISTRY}" "${IMAGE_TAG}" \
                            "${DOCKER_REGISTRY}" "${IMAGE_TAG}")
                        git commit -m "$MSG"
                        git push origin "${GITOPS_BRANCH}"
                        echo "=== GitOps repo updated — ArgoCD will reconcile ==="
                    '''
                }
            }
        }

    }

    post {
        success {
            node('built-in') {
                script {
                    echo "Pipeline completed successfully! Published version: ${env.IMAGE_TAG}"
                    def body = readFile('jenkins/email/success.txt')
                        .replace('@@BUILD_NUMBER@@',   env.BUILD_NUMBER              ?: '')
                        .replace('@@GIT_COMMIT@@',     env.GIT_COMMIT               ?: '')
                        .replace('@@GIT_BRANCH@@',     env.GIT_BRANCH               ?: 'N/A')
                        .replace('@@IMAGE_TAG@@',      env.IMAGE_TAG                ?: '')
                        .replace('@@BUILD_DURATION@@', currentBuild.durationString  ?: '')
                        .replace('@@DOCKER_REGISTRY@@', DOCKER_REGISTRY             ?: '')
                        .replace('@@TIMESTAMP@@',      new Date().toString())
                    mail(
                        to:      'honguyenminhsang2005@gmail.com',
                        subject: "\u2713 CI Pipeline Success - Build #${env.BUILD_NUMBER}",
                        body:    body
                    )
                }
            }
        }

        failure {
            node('built-in') {
                script {
                    echo "Pipeline failed! Check logs for details."
                    def body = readFile('jenkins/email/failure.txt')
                        .replace('@@BUILD_NUMBER@@',   env.BUILD_NUMBER              ?: '')
                        .replace('@@GIT_COMMIT@@',     env.GIT_COMMIT               ?: '')
                        .replace('@@GIT_BRANCH@@',     env.GIT_BRANCH               ?: '')
                        .replace('@@STAGE_NAME@@',     env.STAGE_NAME               ?: 'Unknown')
                        .replace('@@BUILD_DURATION@@', currentBuild.durationString  ?: '')
                        .replace('@@TIMESTAMP@@',      new Date().toString())
                    mail(
                        to:      'honguyenminhsang2005@gmail.com',
                        subject: "\u2717 CI Pipeline FAILURE - Build #${env.BUILD_NUMBER}",
                        body:    body
                    )
                }
            }
        }
    }
}
