COVERALLS_TOKEN := "IVQwNa8dypGgtaLmBkFSoBcRcCl0tlqui"
GITHUB_API_TOKEN := ""
VERSION :=""
APP_NAME := "servicefoundation"
SERVER_NAME := "servicefoundation-1234"
DEPLOY_ENVIRONMENT := "staging"

cover-remote:
	go get -u github.com/golang/lint/golint
	go get -u github.com/mattn/goveralls
	go get -u github.com/golang/dep/cmd/dep
	dep ensure
	go test -covermode=count -coverprofile=cover.tmp
	goveralls -service travis-ci -coverprofile cover.tmp

run-tests:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure
	go test -cover `go list ./... | grep -v /vendor/`

cover:
	go test -cover `go list ./... | grep -v /vendor/`

lint:
	golint `go list ./... | grep -v /vendor/`

vet:
	go vet `go list ./... | grep -v /vendor/`

cover-old:
	#go test -coverprofile=cover.tmp `go list ./... | grep -v /vendor/` && go tool cover -html=cover.tmp `go list ./... | grep -v /vendor/`

