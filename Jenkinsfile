@Library('pipeline-library@master')

pipeline {
    agent {
        node{
        "dev_mesos_v2"
        }
    }

    environment {
        GENESYSCLOUD_OAUTHCLIENT_ID = credentials('GENESYSCLOUD_OAUTHCLIENT_ID')
        GENESYSCLOUD_OAUTHCLIENT_SECRET = credentials('GENESYSCLOUD_OAUTHCLIENT_SECRET')
        GOPATH = "${HOME}/${LANGUAGE}"
		//TF_ACC = "1"
        //TF_LOG = "DEBUG"
        //TF_LOG_PATH = "../test.log"
		GENESYSCLOUD_REGION = "us-east-1"
        GENESYSCLOUD_SDK_DEBUG =  "true"
        GENESYSCLOUD_TOKEN_POOL_SIZE =  20
    }
    tools {
        go 'go 1.21.0'
        terraform 'Terraform 1.4.7'
    }

    stages {
      
       stage('Install Dependencies') {
            steps {
                echo 'Installing dependencies'
                sh 'go version'
                sh 'go mod download'
                sh 'go build -v .'
            }
	   }

       stage('Terraform Check') {
            steps {
                echo 'Check Terraform Installation'
				//sh 'terraform init'
                sh 'terraform -version'

            }
    }

       stage('Tests') {
            steps {
                echo 'Running Tests'
                sh 'go test -timeout 80m -v -cover ./genesyscloud/... -parallel 20 -coverprofile=coverage.out'

            }
    }
  }

}


}
