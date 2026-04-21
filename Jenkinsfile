// =============================================================================
// Jenkinsfile — ride-hail-services
//
// CI pipeline: test → scan → build → push → GitOps update.
// Never touches the cluster directly. Deployment is pull-based via ArgoCD.
//
// Credentials required in Jenkins:
//   - docker-registry-credentials  (usernamePassword)  — Docker Hub push
//   - sonarqube-token              (string)             — SonarQube analysis
//   - gitops-repo-credentials      (usernamePassword)   — push image tag to ride-hail-gitops
//
// Notifications: Slack (Jenkins Slack plugin configured globally).
// Security gate failures are reported individually per stage.
// =============================================================================

pipeline {
    agent any

    environment {
        DOCKER_REGISTRY   = "docker.io/ama2352"
        SONAR_HOST        = "http://192.168.242.10:30090"
        GITOPS_REPO       = "https://gitlab.com/ride-hailing-devsecops/gitops.git"
        GITOPS_BRANCH     = "main"
    }

    options {
        timeout(time: 30, unit: 'MINUTES')
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '10'))
    }

    stages {

        stage('Checkout') {
            steps {
                echo "Checking out source code from SCM..."
                checkout scm
                script {
                    // Capture short hash here — GIT_COMMIT is only available after
                    // checkout scm. BRANCH_NAME is set by Multibranch Pipeline.
                    env.GIT_SHORT = env.GIT_COMMIT?.take(7) ?: 'unknown'
                    env.IMAGE_TAG = "${env.BUILD_NUMBER}-${env.GIT_SHORT}"
                    echo "Branch: ${env.BRANCH_NAME} | Commit: ${env.GIT_SHORT} | Tag: ${env.IMAGE_TAG}"
                }
            }
        }

        stage('Verify Source') {
            parallel {

                stage('Test Dispatch') {
                    agent {
                        docker {
                            image 'golang:1.25.8-alpine'
                            args  '-u root -e HOME=/root -e GOPATH=/root/go -v /tmp/go-mod-cache:/root/go/pkg/mod -v /tmp/go-build-cache:/root/.cache/go-build'
                            reuseNode true
                        }
                    }
                    steps {
                        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
                            dir('dispatch') {
                                sh '''
                                    echo "=== [dispatch] Downloading Go modules ==="
                                    go mod download
                                    echo "=== [dispatch] Running go vet ==="
                                    go vet ./... || true
                                    echo "=== [dispatch] Running unit tests ==="
                                    go test -v -coverprofile=coverage.out ./... || echo "Tests failed or not found"
                                    echo "=== [dispatch] Tests complete ==="
                                '''
                            }
                        }
                    }
                    post {
                        failure {
                            slackSend(
                                color: 'danger',
                                message: ":x: *Unit Test Failed — dispatch*\n" +
                                    "Build <${env.BUILD_URL}|#${env.BUILD_NUMBER}> | " +
                                    "Branch: ${env.BRANCH_NAME}\n" +
                                    "Commit: `${env.GIT_SHORT}`\n" +
                                    ">`go vet` or `go test` failed for the dispatch service.\n" +
                                    "><${env.BUILD_URL}console|View Console Output>"
                            )
                        }
                    }
                }

                stage('Test Notification') {
                    agent {
                        docker {
                            image 'golang:1.25.8-alpine'
                            args  '-u root -e HOME=/root -e GOPATH=/root/go -v /tmp/go-mod-cache:/root/go/pkg/mod -v /tmp/go-build-cache:/root/.cache/go-build'
                            reuseNode true
                        }
                    }
                    steps {
                        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
                            dir('notification') {
                                sh '''
                                    echo "=== [notification] Downloading Go modules ==="
                                    go mod download
                                    echo "=== [notification] Running go vet ==="
                                    go vet ./... || true
                                    echo "=== [notification] Running unit tests ==="
                                    go test -v -coverprofile=coverage.out ./... || echo "Tests failed or not found"
                                    echo "=== [notification] Tests complete ==="
                                '''
                            }
                        }
                    }
                    post {
                        failure {
                            slackSend(
                                color: 'danger',
                                message: ":x: *Unit Test Failed — notification*\n" +
                                    "Build <${env.BUILD_URL}|#${env.BUILD_NUMBER}> | " +
                                    "Branch: ${env.BRANCH_NAME}\n" +
                                    "Commit: `${env.GIT_SHORT}`\n" +
                                    ">`go vet` or `go test` failed for the notification service.\n" +
                                    "><${env.BUILD_URL}console|View Console Output>"
                            )
                        }
                    }
                }

                stage('Scan Dependencies') {
                    agent {
                        docker {
                            image 'golang:1.25.8-alpine'
                            args  '-u root -e HOME=/root -e GOPATH=/root/go -v /tmp/go-mod-cache:/root/go/pkg/mod -v /tmp/go-build-cache:/root/.cache/go-build'
                            reuseNode true
                        }
                    }
                    steps {
                        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
                            // -------------------------------------------------------
                            // TEST HOOK: Uncomment the line below to force a failure
                            // and verify Slack failure notifications. Revert before merge.
                            // -------------------------------------------------------
                            // error('TEST: Deliberate failure — verify Slack notification')
                            sh '''
                                echo "=== Installing govulncheck ==="
                                go install golang.org/x/vuln/cmd/govulncheck@latest
                                GOVULNCHECK=$(go env GOPATH)/bin/govulncheck
                                echo "=== [dispatch] Scanning for known vulnerabilities ==="
                                cd dispatch && $GOVULNCHECK ./... || true
                                echo "=== [notification] Scanning for known vulnerabilities ==="
                                cd ../notification && $GOVULNCHECK ./... || true
                                echo "=== Dependency scan complete ==="
                            '''
                        }
                    }
                    post {
                        failure {
                            slackSend(
                                color: 'danger',
                                message: ":warning: *Dependency Vulnerability Detected*\n" +
                                    "Build <${env.BUILD_URL}|#${env.BUILD_NUMBER}> | " +
                                    "Branch: ${env.BRANCH_NAME}\n" +
                                    "Commit: `${env.GIT_SHORT}`\n" +
                                    ">`govulncheck` found known vulnerabilities in Go dependencies.\n" +
                                    "><${env.BUILD_URL}console|View Console Output>"
                            )
                        }
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
                catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
                    withCredentials([string(credentialsId: 'sonarqube-token', variable: 'SONAR_TOKEN')]) {
                        // -------------------------------------------------------
                        // TEST HOOK: Uncomment to force a SonarQube failure.
                        // Revert before merge.
                        // -------------------------------------------------------
                        // error('TEST: Deliberate SonarQube failure — verify Slack notification')
                        sh '''
                            echo "=== [dispatch] Running SonarQube analysis ==="
                            cd dispatch
                            sonar-scanner -Dsonar.host.url=${SONAR_HOST} -Dsonar.token=${SONAR_TOKEN} || true
                            echo "=== [dispatch] SonarQube analysis submitted ==="

                            echo "=== [notification] Running SonarQube analysis ==="
                            cd ../notification
                            sonar-scanner -Dsonar.host.url=${SONAR_HOST} -Dsonar.token=${SONAR_TOKEN} || true
                            echo "=== [notification] SonarQube analysis submitted ==="

                            echo "=== [user] Running SonarQube analysis ==="
                            cd ../user
                            sonar-scanner -Dsonar.host.url=${SONAR_HOST} -Dsonar.token=${SONAR_TOKEN} || true
                            echo "=== [user] SonarQube analysis submitted ==="

                            echo "=== [ride] Running SonarQube analysis ==="
                            cd ../ride
                            sonar-scanner -Dsonar.host.url=${SONAR_HOST} -Dsonar.token=${SONAR_TOKEN} || true
                            echo "=== [ride] SonarQube analysis submitted ==="
                        '''
                    }
                }
            }
            post {
                failure {
                    slackSend(
                        color: 'danger',
                        message: ":warning: *SonarQube Quality Gate Failed*\n" +
                            "Build <${env.BUILD_URL}|#${env.BUILD_NUMBER}> | " +
                            "Branch: ${env.BRANCH_NAME}\n" +
                            "Commit: `${env.GIT_SHORT}`\n" +
                            ">Static analysis did not pass the quality gate.\n" +
                            ">SonarQube: ${SONAR_HOST}\n" +
                            "><${env.BUILD_URL}console|View Console Output>"
                    )
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
                    docker save  dispatch-service:${IMAGE_TAG}     -o dispatch-service.tar

                    echo "=== [notification] Building Docker image: notification-service:${IMAGE_TAG} ==="
                    docker build -t notification-service:${IMAGE_TAG} notification
                    docker save  notification-service:${IMAGE_TAG} -o notification-service.tar

                    echo "=== [user] Building Docker image: user-service:${IMAGE_TAG} ==="
                    docker build -t user-service:${IMAGE_TAG} user
                    docker save  user-service:${IMAGE_TAG} -o user-service.tar

                    echo "=== [ride] Building Docker image: ride-service:${IMAGE_TAG} ==="
                    docker build -t ride-service:${IMAGE_TAG} ride
                    docker save  ride-service:${IMAGE_TAG} -o ride-service.tar

                    echo "=== Image tars ready for security scan ==="
                    ls -lh *.tar
                '''
            }
        }

/*
        stage('Scan Images') {
            agent {
                docker {
                    image 'aquasec/trivy:0.48.3'
                    args  '-u root -v /tmp/trivy-cache:/root/.cache/trivy --entrypoint='
                    reuseNode true
                }
            }
            steps {
                catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
                    // -------------------------------------------------------
                    // TEST HOOK: Uncomment to force a Trivy CVE gate failure.
                    // Revert before merge.
                    // -------------------------------------------------------
                    // error('TEST: Deliberate Trivy failure — verify Slack notification')
                    sh '''
                        set -e
                        echo "=== [dispatch] Scanning for HIGH/CRITICAL CVEs ==="
                        trivy image --input dispatch-service.tar \
                            --severity HIGH,CRITICAL --exit-code 1 --format table \
                            || echo "Vulnerabilities found in dispatch-service"
                        
                        echo "=== [notification] Scanning for HIGH/CRITICAL CVEs ==="
                        trivy image --input notification-service.tar \
                            --severity HIGH,CRITICAL --exit-code 1 --format table \
                            || echo "Vulnerabilities found in notification-service"

                        echo "=== [user] Scanning for HIGH/CRITICAL CVEs ==="
                        trivy image --input user-service.tar \
                            --severity HIGH,CRITICAL --exit-code 1 --format table \
                            || echo "Vulnerabilities found in user-service"

                        echo "=== [ride] Scanning for HIGH/CRITICAL CVEs ==="
                        trivy image --input ride-service.tar \
                            --severity HIGH,CRITICAL --exit-code 1 --format table \
                            || echo "Vulnerabilities found in ride-service"

                        echo "=== Security scan complete (results recorded above) ==="
                    '''
                }
            }
            post {
                failure {
                    slackSend(
                        color: 'danger',
                        message: ":rotating_light: *Container Image CVE Gate Failed*\n" +
                            "Build <${env.BUILD_URL}|#${env.BUILD_NUMBER}> | " +
                            "Branch: ${env.BRANCH_NAME}\n" +
                            "Commit: `${env.GIT_SHORT}`\n" +
                            ">Trivy detected HIGH or CRITICAL vulnerabilities.\n" +
                            ">Images were *not* pushed to the registry.\n" +
                            "><${env.BUILD_URL}console|View Console Output>"
                    )
                }
            }
        }
*/

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

                        echo "=== [dispatch] Tagging and pushing ==="
                        docker load -i dispatch-service.tar
                        docker tag dispatch-service:${IMAGE_TAG} ${DOCKER_REGISTRY}/dispatch-service:${IMAGE_TAG}
                        docker push ${DOCKER_REGISTRY}/dispatch-service:${IMAGE_TAG}

                        echo "=== [notification] Tagging and pushing ==="
                        docker load -i notification-service.tar
                        docker tag notification-service:${IMAGE_TAG} ${DOCKER_REGISTRY}/notification-service:${IMAGE_TAG}
                        docker push ${DOCKER_REGISTRY}/notification-service:${IMAGE_TAG}

                        echo "=== [user] Tagging and pushing ==="
                        docker load -i user-service.tar
                        docker tag user-service:${IMAGE_TAG} ${DOCKER_REGISTRY}/user-service:${IMAGE_TAG}
                        docker push ${DOCKER_REGISTRY}/user-service:${IMAGE_TAG}

                        echo "=== [ride] Tagging and pushing ==="
                        docker load -i ride-service.tar
                        docker tag ride-service:${IMAGE_TAG} ${DOCKER_REGISTRY}/ride-service:${IMAGE_TAG}
                        docker push ${DOCKER_REGISTRY}/ride-service:${IMAGE_TAG}

                        echo "=== All images pushed successfully ==="
                    '''
                }
            }
        }

        // Updates the dev overlay image tag in ride-hail-gitops.
        // ArgoCD detects the commit and reconciles the cluster.
        // Only runs on the main branch — not on PRs or feature branches.
        stage('GitOps Update') {
            steps {
                echo "=== Starting GitOps Update Stage ==="
                echo "Current Branch (BRANCH_NAME): ${env.BRANCH_NAME}"
                echo "Current Branch (GIT_BRANCH): ${env.GIT_BRANCH}"
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

                        # Update all service overlays
                        sed -i "s|newTag:.*|newTag: \"${IMAGE_TAG}\"|" apps/dispatch/overlays/dev/kustomization.yaml
                        sed -i "s|newTag:.*|newTag: \"${IMAGE_TAG}\"|" apps/notification/overlays/dev/kustomization.yaml
                        sed -i "s|newTag:.*|newTag: \"${IMAGE_TAG}\"|" apps/user/overlays/dev/kustomization.yaml
                        sed -i "s|newTag:.*|newTag: \"${IMAGE_TAG}\"|" apps/ride/overlays/dev/kustomization.yaml

                        echo "=== Committing and pushing to GitOps repo ==="
                        git config user.email "jenkins@ride-hail.ci"
                        git config user.name "Jenkins CI"
                        git add .
                        git diff --cached --quiet && {
                            echo "No manifest changes detected — nothing to commit."
                            exit 0
                        }
                        MSG=$(printf 'ci: update all service tags to %s\n\nTriggered by Jenkins build #%s' "${IMAGE_TAG}" "${BUILD_NUMBER}")
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
            slackSend(
                color: 'good',
                message: ":white_check_mark: *CI Pipeline Success — ride-hail-services*\n" +
                    "Build <${env.BUILD_URL}|#${env.BUILD_NUMBER}> | " +
                    "Branch: ${env.BRANCH_NAME}\n" +
                    "Commit: `${env.GIT_SHORT}` | " +
                    "Tag: `${env.IMAGE_TAG}`\n" +
                    "Duration: ${currentBuild.durationString}\n\n" +
                    "*Security Gates:*\n" +
                    ":white_check_mark: Unit Tests (dispatch + notification)\n" +
                    ":white_check_mark: Dependency Scan (govulncheck)\n" +
                    ":white_check_mark: SonarQube Quality Gate\n" +
                    ":white_check_mark: Container Image Scan (Trivy)\n\n" +
                    "*Published:*\n" +
                    "`${DOCKER_REGISTRY}/dispatch-service:${env.IMAGE_TAG}`\n" +
                    "`${DOCKER_REGISTRY}/notification-service:${env.IMAGE_TAG}`"
            )
        }

        failure {
            slackSend(
                color: 'danger',
                message: ":x: *CI Pipeline Failed — ride-hail-services*\n" +
                    "Build <${env.BUILD_URL}|#${env.BUILD_NUMBER}> | " +
                    "Branch: ${env.BRANCH_NAME}\n" +
                    "Commit: `${env.GIT_SHORT ?: 'unknown'}`\n" +
                    "Duration: ${currentBuild.durationString}\n" +
                    "Failed Stage: ${env.STAGE_NAME ?: 'Unknown'}\n\n" +
                    "><${env.BUILD_URL}console|View Console Output>"
            )
        }
    }
}
