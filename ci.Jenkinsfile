pipeline {

    agent any

    environment {
        GITHUB_CREDS_PSW = credentials("GITHUB_TOKEN")
    }

    stages {
        stage("Test with Go 1.16") {
            tools {
                go 'Go 1.16'
            }
            steps {
                withEnv(["PATH+EXTRA=${HOME}/go/bin"]) {
                  sh 'make checkfmt vet tests'
                }
            }
        }

        stage("Test with Go 1.17") {
            tools {
                go 'Go 1.17'
            }
            steps {
                withEnv(["PATH+EXTRA=${HOME}/go/bin"]) {
                  sh 'make checkfmt vet tests'
                }
            }
        }

        stage("Publish") {
            tools {
                go 'Go 1.17'
            }
            environment {
                RELEASE_TYPE = "alpha"
                VERSION_POSTFIX = "-alpha-${GIT_COMMIT.substring(0, 8)}"
                HARBOR_CREDS = credentials("projects-registry-vmware-tanzu_observability_keights_saas-robot")
                PREFIX = "projects.registry.vmware.com/tanzu_observability_keights_saas"
                DOCKER_IMAGE = "kubernetes-collector-snapshot"
            }
            steps {
                withEnv(["PATH+EXTRA=${HOME}/go/bin"]) {
                    sh './hack/jenkins/install_docker_buildx.sh'
                    sh 'make semver-cli'
                    sh 'echo $HARBOR_CREDS_PSW | docker login $PREFIX -u $HARBOR_CREDS_USR --password-stdin'
                    sh 'HARBOR_CREDS_USR=$(echo $HARBOR_CREDS_USR | sed \'s/\\$/\\$\\$/\') make publish'
                }
            }
        }

        stage("Integration Test") {
            options {
                timeout(time: 10, unit: 'MINUTES')
            }
            tools {
                go 'Go 1.17'
            }
            environment {
                GCP_CREDS = credentials("GCP_CREDS")
                GKE_CLUSTER_NAME = "k8po-jenkins-ci"
                WAVEFRONT_TOKEN = credentials("WAVEFRONT_TOKEN_NIMBA")
                VERSION_POSTFIX = "-alpha-${GIT_COMMIT.substring(0, 8)}"
                PREFIX = "projects.registry.vmware.com/tanzu_observability_keights_saas"
                DOCKER_IMAGE = "kubernetes-collector-snapshot"
            }
            steps {
                withEnv(["PATH+GO=${HOME}/go/bin", "PATH+GCLOUD=${HOME}/google-cloud-sdk/bin"]) {
                    lock("collector-integration-test") {
                        sh './hack/jenkins/setup-for-integration-test.sh'
                        sh 'make gke-connect-to-cluster'
                        sh 'VERSION_POSTFIX=$VERSION_POSTFIX make deploy-test'
                    }
                }
            }
        }
    }

    post {
        regression {
            slackSend (channel: '#tobs-k8po-team', color: '#FF0000', message: "BUILD FAILED: '<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>")
        }
        fixed {
            slackSend (channel: '#tobs-k8po-team', color: '#008000', message: "BUILD FIXED: '<${env.BUILD_URL}|${env.JOB_NAME} [${env.BUILD_NUMBER}]>")
        }
        success {
            script {
                if (env.BRANCH_NAME == 'master') {
                    sh './hack/jenkins/update_github_status.sh'
                }
            }
        }
    }
}