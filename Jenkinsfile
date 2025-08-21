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
  }
}

