pipeline {
  agent any
  stages {
    stage('Checkout') {
      steps { checkout scm }
    }
    stage('Build') {
      steps {
        script {
          docker.image('golang:1.23').inside {
            sh '''
              go mod download
              make build
              ls -lh bin/
            '''
          }
        }
        archiveArtifacts artifacts: 'bin/coffee', onlyIfSuccessful: true
      }
    }
  }
}

