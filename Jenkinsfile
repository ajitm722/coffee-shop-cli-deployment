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
          // same caches + also mount the Docker socket for testcontainers
          def runArgs = '-e HOME=/var/jenkins_home ' +
                        '-e GOCACHE=/var/jenkins_home/.cache/go-build ' +
                        '-e GOMODCACHE=/var/jenkins_home/go/pkg/mod ' +
                        '-v /var/run/docker.sock:/var/run/docker.sock'

          docker.image('golang:1.23').inside(runArgs) {
            sh '''
              # 1) Functional integration tests (verbose, single run)
              go test -tags=integration ./test/integration/... -count=1 -v

              # 2) Benchmarks (skip normal tests with -run '^$'; include benchmem)
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

