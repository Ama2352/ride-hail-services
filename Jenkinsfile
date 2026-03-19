// =============================================================================
// Jenkinsfile — ride-hail-services (Benchmark Optimized)
// =============================================================================

pipeline {
    agent none

    environment {
        DOCKER_REGISTRY   = "docker.io/ama2352"
        SONAR_HOST        = "https://sonarcloud.io"
        SONAR_ORG         = "ama2352"
        GITOPS_REPO       = "https://github.com/ama2352/ride-hail-gitops.git"
        GITOPS_BRANCH     = "benchmark"
    }

    options {
        timeout(time: 30, unit: 'MINUTES')
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '10'))
    }

    stages {
        stage('Checkout') {
            agent any
            steps {
                checkout scm
                script {
                    env.GIT_SHORT = env.GIT_COMMIT?.take(7) ?: 'unknown'
                    env.IMAGE_TAG = "${env.BUILD_NUMBER}-${env.GIT_SHORT}"
                }
            }
        }

        stage('Verify') {
            parallel {
                stage('Test Dispatch') {
                    agent {
                        docker { 
                            image 'golang:1.25.8-alpine'
                            args '-u root -v /tmp/go-mod-cache:/go/pkg/mod'
                        }
                    }
                    steps {
                        dir('dispatch') {
                            sh '''
                                go mod download
                                go vet ./...
                                go test -v -coverprofile=coverage.out ./...
                            '''
                        }
                        // Stash coverage report để SonarQube dùng ở stage sau
                        stash name: 'coverage-dispatch', includes: 'dispatch/coverage.out', allowEmpty: true
                    }
                }
                
                stage('Test Notification') {
                    agent {
                        docker { 
                            image 'golang:1.25.8-alpine'
                            args '-u root -v /tmp/go-mod-cache:/go/pkg/mod'
                        }
                    }
                    steps {
                        dir('notification') {
                            sh '''
                                go mod download
                                go vet ./...
                                go test -v -coverprofile=coverage.out ./...
                            '''
                        }
                        stash name: 'coverage-notification', includes: 'notification/coverage.out', allowEmpty: true
                    }
                }

                stage('Scan Dependencies') {
                    agent {
                        docker { 
                            image 'golang:1.25.8-alpine'
                            args '-u root -v /tmp/go-mod-cache:/go/pkg/mod'
                        }
                    }
                    steps {
                        sh '''
                            go install golang.org/x/vuln/cmd/govulncheck@latest
                            export PATH=$PATH:$(go env GOPATH)/bin
                            cd dispatch && govulncheck ./...
                            cd ../notification && govulncheck ./...
                        '''
                    }
                }
            }
        }

        stage('SonarQube Analysis') {
            agent {
                docker { 
                    image 'sonarsource/sonar-scanner-cli:11.3'
                    args '-u root'
                }
            }
            steps {
                unstash 'coverage-dispatch'
                unstash 'coverage-notification'
                
                withCredentials([string(credentialsId: 'sonarqube-token', variable: 'SONAR_TOKEN')]) {
                    sh '''
                        # Quét Dispatch
                        cd dispatch
                        sonar-scanner \
                            -Dsonar.projectKey=ama2352_ridehail-dispatch-service \
                            -Dsonar.organization=${SONAR_ORG} \
                            -Dsonar.host.url=${SONAR_HOST} \
                            -Dsonar.token=${SONAR_TOKEN} \
                            -Dsonar.qualitygate.wait=false \
                            -Dsonar.go.coverage.reportPaths=coverage.out

                        # Quét Notification
                        cd ../notification
                        sonar-scanner \
                            -Dsonar.projectKey=ama2352_ridehail-notification-service \
                            -Dsonar.organization=${SONAR_ORG} \
                            -Dsonar.host.url=${SONAR_HOST} \
                            -Dsonar.token=${SONAR_TOKEN} \
                            -Dsonar.qualitygate.wait=false \
                            -Dsonar.go.coverage.reportPaths=coverage.out
                    '''
                }
            }
        }

        stage('Build Images') {
            agent {
                docker {
                    image 'docker:26-cli'
                    // Mount socket để áp dụng chuẩn DooD giống GitLab Runner
                    args '-v /var/run/docker.sock:/var/run/docker.sock -u root'
                }
            }
            steps {
                sh '''
                    docker build -t dispatch-service:${IMAGE_TAG} dispatch
                    docker save dispatch-service:${IMAGE_TAG} -o dispatch-service.tar
                    
                    docker build -t notification-service:${IMAGE_TAG} notification
                    docker save notification-service:${IMAGE_TAG} -o notification-service.tar
                '''
                // Pass file tar sang stage Scan giống như artifacts trong GitLab
                stash name: 'tar-images', includes: '*.tar'
            }
        }

        stage('Scan Images') {
            agent {
                docker {
                    image 'aquasec/trivy:0.48.3'
                    args '-u root --entrypoint='
                }
            }
            steps {
                unstash 'tar-images'
                sh '''
                    trivy image --input dispatch-service.tar --severity HIGH,CRITICAL --exit-code 1 --format table
                    trivy image --input notification-service.tar --severity HIGH,CRITICAL --exit-code 1 --format table
                '''
            }
        }

        stage('Push Images') {
            when { branch 'benchmark' }
            agent {
                docker {
                    image 'docker:26-cli'
                    args '-v /var/run/docker.sock:/var/run/docker.sock -u root'
                }
            }
            steps {
                withCredentials([usernamePassword(credentialsId: 'docker-registry-credentials', usernameVariable: 'DOCKER_USER', passwordVariable: 'DOCKER_PASS')]) {
                    sh '''
                        echo "${DOCKER_PASS}" | docker login -u "${DOCKER_USER}" --password-stdin
                        
                        docker tag dispatch-service:${IMAGE_TAG} ${DOCKER_REGISTRY}/dispatch-service:${IMAGE_TAG}
                        docker tag dispatch-service:${IMAGE_TAG} ${DOCKER_REGISTRY}/dispatch-service:latest
                        docker push ${DOCKER_REGISTRY}/dispatch-service:${IMAGE_TAG}
                        docker push ${DOCKER_REGISTRY}/dispatch-service:latest

                        docker tag notification-service:${IMAGE_TAG} ${DOCKER_REGISTRY}/notification-service:${IMAGE_TAG}
                        docker tag notification-service:${IMAGE_TAG} ${DOCKER_REGISTRY}/notification-service:latest
                        docker push ${DOCKER_REGISTRY}/notification-service:${IMAGE_TAG}
                        docker push ${DOCKER_REGISTRY}/notification-service:latest
                    '''
                }
            }
        }

        stage('GitOps Update') {
            when { branch 'benchmark' }
            agent {
                docker { image 'alpine:3.20' }
            }
            steps {
                withCredentials([usernamePassword(credentialsId: 'gitops-repo-credentials', usernameVariable: 'GIT_USER', passwordVariable: 'GIT_TOKEN')]) {
                    sh '''
                        apk add --no-cache git sed
                        
                        GITOPS_URL=$(echo "${GITOPS_REPO}" | sed "s|https://|https://${GIT_USER}:${GIT_TOKEN}@|")
                        git clone --branch "${GITOPS_BRANCH}" --depth 1 "${GITOPS_URL}" gitops-workspace
                        cd gitops-workspace

                        sed -i "s|newTag:.*|newTag: \\"${IMAGE_TAG}\\"|" apps/dispatch/overlays/dev/kustomization.yaml
                        sed -i "s|newTag:.*|newTag: \\"${IMAGE_TAG}\\"|" apps/notification/overlays/dev/kustomization.yaml

                        git config user.email "jenkins-ci@ride-hail.ci"
                        git config user.name "Jenkins CI"
                        git add apps/dispatch/overlays/dev/kustomization.yaml apps/notification/overlays/dev/kustomization.yaml
                        git diff --cached --quiet && echo "No GitOps changes" && exit 0
                        
                        git commit -m "ci: update image tags to ${IMAGE_TAG}

Triggered by Jenkins pipeline #${BUILD_NUMBER}
Commit: ${GIT_SHORT}"
                        
                        git push origin "${GITOPS_BRANCH}"
                    '''
                }
            }
        }
    }
}