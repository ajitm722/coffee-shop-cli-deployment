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
          // discover host docker.sock GID so the test container can use the socket
          def dockerGid = sh(returnStdout: true, script: "stat -c %g /var/run/docker.sock").trim()

          def runArgs = "-e HOME=/var/jenkins_home " +
                        "-e GOCACHE=/var/jenkins_home/.cache/go-build " +
                        "-e GOMODCACHE=/var/jenkins_home/go/pkg/mod " +
                        // talk to the host Docker
                        "-e DOCKER_HOST=unix:///var/run/docker.sock " +
                        "-v /var/run/docker.sock:/var/run/docker.sock " +
                        "-u 0:0 " + // run as root to avoid permission issues with the socket
                        // make the host resolvable/reachable for Ryuk healthcheck
                        "--add-host=host.docker.internal:host-gateway " +
                        "-e TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal"

          docker.image('golang:1.23').inside(runArgs) {
            sh '''
              # 1) Functional integration tests
              go test -tags=integration ./test/integration/... -count=1 -v

              # 2) Benchmarks
              go test -tags=integration ./test/integration/... -bench . -benchmem -run '^$' -count=1 | tee bench.txt
            '''
          }
        }
      }
      post {
        always {
          archiveArtifacts artifacts: 'bench.txt', allowEmptyArchive: true
        }
      }
    }

  }
}

