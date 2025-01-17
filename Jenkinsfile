node {
    // Install the desired Go version
    def root = tool name: 'go 1.14.3', type: 'go'

    env.GOROOT="${root}"
    env.GOBIN="${WORKSPACE}/bin"
    env.PATH="${root}/bin:${env.GOBIN}:${env.PATH}"
    env.UNIDOC_EXTRACT_FORCETEST="1"
    env.UNIDOC_E2E_FORCE_TESTS="1"
    env.UNIDOC_RENDERTEST_FORMFILL_FORCETEST="1"
    env.UNIDOC_EXTRACT_TESTDATA="/home/jenkins/corpus/unidoc-extractor-testdata"
    env.UNIDOC_RENDERTEST_BASELINE_PATH="/home/jenkins/corpus/unidoc-creator-render-testdata-upd2"
    env.UNIDOC_PASSTHROUGH_TESTDATA="/home/jenkins/corpus/unidoc-e2e-testdata"
    env.UNIDOC_ALLOBJECTS_TESTDATA="/home/jenkins/corpus/unidoc-e2e-testdata"
    env.UNIDOC_SPLIT_TESTDATA="/home/jenkins/corpus/unidoc-e2e-split-testdata"
    env.UNIDOC_EXTRACT_IMAGES_TESTDATA="/home/jenkins/corpus/unidoc-e2e-extract-images-testdata"
    env.UNIDOC_JBIG2_TESTDATA="/home/jenkins/corpus/jbig2-testdata"
    env.UNIDOC_FDFMERGE_TESTDATA="/home/jenkins/corpus/fdfmerge-testdata"
    env.UNIDOC_RENDERTEST_FORMFILL_TESTDATA="/home/jenkins/corpus/unidoc-form-fill-render-testdata"
    env.UNIDOC_RENDERTEST_FORMFILL_BASELINE="/home/jenkins/corpus/unidoc-form-fill-render-baseline"
    env.UNIDOC_GS_BIN_PATH="/usr/bin/gs"
    env.CGO_ENABLED="0"

    env.TMPDIR="${WORKSPACE}/temp"
    sh "mkdir -p ${env.GOBIN}"
    sh "mkdir -p ${env.TMPDIR}"

    dir("${WORKSPACE}/unipdf") {
        sh 'go version'

        stage('Checkout') {
            echo "Pulling unipdf on branch ${env.BRANCH_NAME}"
            checkout scm
        }

        stage('Prepare') {
            // Get linter and other build tools.
            sh 'go get golang.org/x/lint/golint'
            sh 'go get github.com/tebeka/go2xunit'
            sh 'go get github.com/t-yuki/gocover-cobertura'
        }

        stage('Linting') {
            // Go vet - List issues
            sh '(go vet ./... >govet.txt 2>&1) || true'

            // Go lint - List issues
            sh '(golint ./... >golint.txt 2>&1) || true'
        }

        stage('Testing') {
            // Go test - No tolerance.
            sh "rm -f ${env.TMPDIR}/*.pdf"
            sh '2>&1 go test -count=1 -v ./... | tee gotest.txt'
        }

        stage('Check generated PDFs') {
            // Check the created output pdf files.
            sh "find ${env.TMPDIR} -maxdepth 1 -name \"*.pdf\" -print0 | xargs -t -n 1 -0 gs -dNOPAUSE -dBATCH -sDEVICE=nullpage -sPDFPassword=password -dPDFSTOPONERROR -dPDFSTOPONWARNING"
        }

        stage('Test coverage') {
            sh 'go test -count=1 -coverprofile=coverage.out -covermode=atomic -coverpkg=./... ./...'
            sh '/home/jenkins/codecov.sh'
            sh 'gocover-cobertura < coverage.out > coverage.xml'
            step([$class: 'CoberturaPublisher', coberturaReportFile: 'coverage.xml'])
        }

        stage('Post') {
            // Assemble vet and lint info.
            warnings parserConfigurations: [
                [pattern: 'govet.txt', parserName: 'Go Vet'],
                [pattern: 'golint.txt', parserName: 'Go Lint']
            ]

            sh 'go2xunit -fail -input gotest.txt -output gotest.xml'
            junit "gotest.xml"
        }
    }

    dir("${WORKSPACE}/unipdf-examples") {
        stage('Build examples') {
            // Output environment variables (useful for debugging).
            sh("printenv")

            // Pull unipdf-examples from connected branch, or master otherwise.
            def examplesBranch = "development"

            // Check if connected branch is defined explicitly.
            def safeName = env.BRANCH_NAME.replaceAll(/[\/\.]/, '')
            def fpath = "/home/jenkins/exbranch/" + safeName
            if (fileExists(fpath)) {
                examplesBranch = readFile(fpath).trim()
            }

            echo "Pulling unipdf-examples on branch ${examplesBranch}"
            git url: 'https://github.com/unidoc/unidoc-examples.git', branch: examplesBranch

            // Use replace directive to use disk version of unipdf.
            sh 'echo "replace github.com/mgmeyers/unipdf/v3 => ../unipdf" >>go.mod'

            // Dependencies for examples.
            sh './build_examples.sh'
        }

        stage('Passthrough benchmark pdfdb_small') {
            sh './bin/pdf_passthrough_bench /home/jenkins/corpus/pdfdb_small/* | grep -v "Testing " | grep -v "copy of" | grep -v "To get " | grep -v " - pass"'
        }
    }
}
