COVERALLS_TOKEN := "IVQwNa8dypGgtaLmBkFSoBcRcCl0tlqui"
GITHUB_API_TOKEN := ""
VERSION :=""

export GO111MODULE=on
export APP_NAME=servicefoundation
export SERVER_NAME=servicefoundation-1234
export DEPLOY_ENVIRONMENT=staging

cover-remote:
	go get -u golang.org/x/lint/golint
	go get -u github.com/mattn/goveralls
	go test -covermode=count -coverprofile=cover.tmp
	goveralls -service travis-ci -coverprofile cover.tmp

run-tests:
	go test -race -cover -v ./...

cover:
	go test -cover ./...

lint:
	golint ./...

vet:
	go vet ./...

clean:
	go clean

upgrade:
	go get -u

env:
	go env
