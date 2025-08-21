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

  }
}

