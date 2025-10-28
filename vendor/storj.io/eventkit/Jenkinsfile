pipeline {

    agent {
       label 'ondemand'
    }

    options {
          timeout(time: 26, unit: 'MINUTES')
    }

    environment {
        GOTRACEBACK = 'all'
    }

    stages {
        stage('Gerrit status') {
            steps {
                withCredentials([sshUserPrivateKey(credentialsId: 'gerrit-trigger-ssh', keyFileVariable: 'SSH_KEY', usernameVariable: 'SSH_USER')]) {
                    sh './scripts/gerrit-status.sh verify start 0'
                }
            }
        }
        stage('Checkout') {
            steps {
               checkout scm
            }
        }
        stage('Lint') {
            steps {
                sh "earthly --ci +lint"
            }
        }
        stage('Test') {
            steps {
                sh "earthly --ci +test"
            }
        }
    }
    post {
        success {
            withCredentials([sshUserPrivateKey(credentialsId: 'gerrit-trigger-ssh', keyFileVariable: 'SSH_KEY', usernameVariable: 'SSH_USER')]) {
                sh './scripts/gerrit-status.sh verify success +1'
            }
        }
        failure {
            withCredentials([sshUserPrivateKey(credentialsId: 'gerrit-trigger-ssh', keyFileVariable: 'SSH_KEY', usernameVariable: 'SSH_USER')]) {
                sh './scripts/gerrit-status.sh verify failure -1'
            }
        }
        aborted {
            withCredentials([sshUserPrivateKey(credentialsId: 'gerrit-trigger-ssh', keyFileVariable: 'SSH_KEY', usernameVariable: 'SSH_USER')]) {
                sh './scripts/gerrit-status.sh verify failure -1'
            }
        }
    }
}