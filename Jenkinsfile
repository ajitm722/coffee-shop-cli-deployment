pipeline {
  agent any
  stages {
    stage('Checkout') {
      steps { checkout scm }
    }
    stage('Build') {
      steps {
        script {
          // Give the container a proper HOME and Go caches on a writable volume
          def runArgs = '-e HOME=/var/jenkins_home ' +
                        '-e GOCACHE=/var/jenkins_home/.cache/go-build ' +
                        '-e GOMODCACHE=/var/jenkins_home/go/pkg/mod'

          docker.image('golang:1.23').inside(runArgs) {
            sh '''
              echo "HOME=$HOME"
              go env GOCACHE GOMODCACHE
              go version

              go mod download
              make build
              ls -lh bin/
            '''
          }
        }
        archiveArtifacts artifacts: 'bin/coffee', onlyIfSuccessful: true
      }
    }
    stage('Test - Unit') {
      steps {
        script {
          // reuse the same runArgs you used in Build
          def runArgs = '-e HOME=/var/jenkins_home ' +
                        '-e GOCACHE=/var/jenkins_home/.cache/go-build ' +
                        '-e GOMODCACHE=/var/jenkins_home/go/pkg/mod'

          docker.image('golang:1.23').inside(runArgs) {
            sh '''
              go test ./... -race -coverprofile=coverage.out
            '''
          }
        }
        archiveArtifacts artifacts: 'coverage.out', allowEmptyArchive: true
      }
    }
    stage('Test - Integration') {
      steps {
        script {
          def runArgs =
            "-e HOME=/var/jenkins_home " +
            "-e GOCACHE=/var/jenkins_home/.cache/go-build " +
            "-e GOMODCACHE=/var/jenkins_home/go/pkg/mod " +
            "-e DOCKER_HOST=unix:///var/run/docker.sock " +
            "-v /var/run/docker.sock:/var/run/docker.sock " +
            "--add-host=host.docker.internal:host-gateway " +
            "-e TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal " +
            "-e TESTCONTAINERS_RYUK_DISABLED=true " +           // ← disable Ryuk
            "-e TESTCONTAINERS_CHECKS_DISABLE=true " +          // ← skip external reachability check
            "-u 0:0"                                            // already using root here; OK

          docker.image('golang:1.23').inside(runArgs) {
            sh '''
              go test -tags=integration ./test/integration/... -count=1 -v
              go test -tags=integration ./test/integration/... -bench . -benchmem -run '^$' -count=1 | tee bench.txt
            '''
          }
        }
      }
      post {
        always {
          archiveArtifacts artifacts: 'bench.txt', allowEmptyArchive: true
          // Optional safety cleanup if a job is hard-killed:
          // sh 'docker ps -aq -f label=org.testcontainers.sessionId | xargs -r docker rm -f -v || true'
        }
      }
    }
    stage('Code Quality') {
      steps {
        script {
          def runArgs = '-e HOME=/var/jenkins_home ' +
                        '-e GOCACHE=/var/jenkins_home/.cache/go-build ' +
                        '-e GOMODCACHE=/var/jenkins_home/go/pkg/mod'

          docker.image('golang:1.23').inside(runArgs) {
            sh '''
              echo "== gofmt check =="
              CHANGED=$(gofmt -s -l .)
              if [ -n "$CHANGED" ]; then
                echo "Files need gofmt:"; echo "$CHANGED"
                exit 1
              fi

              echo "== golangci-lint =="
              go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
              $GOPATH/bin/golangci-lint run ./...


              echo "== go vet (static checks) =="
              # Static analysis for suspicious constructs (misused printf, unreachable code, etc).
              # Non-zero exit on issues -> fails the build.
              go vet ./...

            '''
          }
        }
      }
    }
    stage('Security – Vulns, Secrets & Code Scans') {
      steps {
        script {
          // Reuse Go build caches for speed
          def runArgs = '-e HOME=/var/jenkins_home ' +
                        '-e GOCACHE=/var/jenkins_home/.cache/go-build ' +
                        '-e GOMODCACHE=/var/jenkins_home/go/pkg/mod'

          docker.image('golang:1.24').inside(runArgs) {
            sh '''
              set -e

              echo "== gitleaks (secret scanning) =="
              # Scans the repository for hard-coded secrets (tokens, keys, passwords).
              # Fails the build if a secret is found.
              go install github.com/zricethezav/gitleaks/v8@latest
              $GOPATH/bin/gitleaks detect --redact

              echo "== govulncheck (Go module advisories) =="
              # Analyzes your code + dependency graph for known vulnerabilities (Go vuln DB).
              # Fails the build if a reachable vulnerability is detected.
              go install golang.org/x/vuln/cmd/govulncheck@latest
              $GOPATH/bin/govulncheck ./...

              echo "== gosec (SAST) =="
              # Security-focused static analysis: looks for risky patterns (SQLi, weak crypto, path traversal).
              # Non-zero exit on findings -> fails the build by default.
              go install github.com/securego/gosec/v2/cmd/gosec@latest
              $GOPATH/bin/gosec ./...
            '''
          }
        }
      }
    }
    stage('Deploy Stage - Package & Push (ACR)') {
      environment {
        APP      = 'coffee'
        ACR_NAME = 'acrcoffeedev'
        REGISTRY = "acrcoffeedev.azurecr.io"
      }
      steps {
        withCredentials([string(credentialsId: 'az-sp-json', variable: 'AZ_SP_JSON')]) {
          writeFile file: 'deploy-acr.sh', text: '''
          #!/usr/bin/env bash
          set -euo pipefail

          CLIENT_ID=$(echo "$AZ_SP_JSON" | jq -r .appId)
          CLIENT_SECRET=$(echo "$AZ_SP_JSON" | jq -r .password)
          TENANT_ID=$(echo "$AZ_SP_JSON" | jq -r .tenant)
          SUB_ID=$(echo "$AZ_SP_JSON" | jq -r .subscription)

          echo "Logging in to Azure subscription: $SUB_ID"
          az login --service-principal -u "$CLIENT_ID" -p "$CLIENT_SECRET" --tenant "$TENANT_ID" >/dev/null
          az account set --subscription "$SUB_ID"

          VERSION="$(date +%Y.%m.%d).${BUILD_NUMBER}-sha${GIT_COMMIT:0:7}"
          echo "$VERSION" > .version
          echo "Version: $VERSION"

          az acr login -n "${ACR_NAME}"

          docker build -t "${REGISTRY}/${APP}:${VERSION}" .
          docker tag   "${REGISTRY}/${APP}:${VERSION}" "${REGISTRY}/${APP}:latest"
          docker push  "${REGISTRY}/${APP}:${VERSION}"
          docker push  "${REGISTRY}/${APP}:latest"

          echo "Pushed:"
          echo "  ${REGISTRY}/${APP}:${VERSION}"
          echo "  ${REGISTRY}/${APP}:latest"
          echo "Done."
          '''
          sh 'chmod +x ./deploy-acr.sh && bash ./deploy-acr.sh'
      }

      }
      post {
        success {
          archiveArtifacts artifacts: '.version', onlyIfSuccessful: true
        }
      }
    }
    stage('Release to Minikube') {
      environment {
        APP        = 'coffee'
        REGISTRY   = 'acrcoffeedev.azurecr.io'
        KUBECONFIG = '/var/jenkins_home/.kube/config'
      }
      steps {
        withCredentials([string(credentialsId: 'az-sp-json', variable: 'AZ_SP_JSON')]) {
          writeFile file: 'release-minikube.sh', text: '''
            #!/usr/bin/env bash
            set -euo pipefail

            export KUBECONFIG="/var/jenkins_home/.kube/config"

            # Ensure kubeconfig paths point inside Jenkins home (idempotent)
            sed -i 's|/home/[^/]*/.minikube|/var/jenkins_home/.minikube|g' "$KUBECONFIG" || true

            # Use minikube context
            kubectl config use-context minikube

            # Pull ACR creds from injected JSON
            CLIENT_ID=$(echo "$AZ_SP_JSON" | jq -r .appId)
            CLIENT_SECRET=$(echo "$AZ_SP_JSON" | jq -r .password)

            # Ensure / refresh imagePull secret (idempotent)
            kubectl create secret docker-registry acr-cred \
              --docker-server="${REGISTRY}" \
              --docker-username="${CLIENT_ID}" \
              --docker-password="${CLIENT_SECRET}" \
              --docker-email=none@example.com \
              --dry-run=client -o yaml | kubectl apply -f -

            # Apply baseline manifests from repo (idempotent)
            kubectl apply -f k8s/initdb-configmap.yaml
            kubectl apply -f k8s/postgres.yaml
            kubectl apply -f k8s/service.yaml
            kubectl apply -f k8s/deployment.yaml

            # Make sure the Deployment references the pull secret (safe to repeat)
            kubectl patch deployment/coffee-api --type=merge -p '{
              "spec": { "template": { "spec": { "imagePullSecrets": [ { "name": "acr-cred" } ] } } }
            }' || true

            # Read the image tag produced in the ACR stage
            VERSION="$(cat .version)"
            echo "Releasing ${REGISTRY}/${APP}:${VERSION}"

            # Update image and roll out
            kubectl set image deployment/coffee-api coffee="${REGISTRY}/${APP}:${VERSION}"
            kubectl rollout status deployment/coffee-postgres --timeout=120s || true
            kubectl rollout status deployment/coffee-api --timeout=180s

            # In-cluster smoke test (health + ready + menu)
            kubectl run curl --rm -i --restart=Never --image=curlimages/curl -- \
              sh -lc "set -e; \
                curl -sf http://coffee-api:9090/v1/healthz >/dev/null && \
                curl -sf http://coffee-api:9090/v1/readyz  >/dev/null && \
                curl -sf http://coffee-api:9090/v1/menu    >/dev/null"

            echo "Release to Minikube complete."
            '''
            sh 'chmod +x ./release-minikube.sh && bash ./release-minikube.sh'
        }
      }
      post {
        failure {
          sh '''
            export KUBECONFIG="/var/jenkins_home/.kube/config"
            echo "==== Debug dump ===="
            kubectl get deploy,svc,pods -o wide
            echo "---- describe deploy ----"
            kubectl describe deploy/coffee-api || true
            echo "---- recent logs ----"
            kubectl logs deploy/coffee-api --tail=200 || true
          '''
        }
      }
    }


  }
}

