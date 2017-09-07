COVERALLS_TOKEN := "U1GTs4phjw6ebdzs6SiQ1mzKQB2875zn5"
GITHUB_API_TOKEN := ""
VERSION :=""
APP_NAME := "servicefoundation"
SERVER_NAME := "servicefoundation-1234"
DEPLOY_ENVIRONMENT := "staging"

cover-remote:
	go get -u github.com/mattn/goveralls
	go get -u github.com/Masterminds/glide
	glide install
	goveralls -service travis-ci -coverprofile cover.tmp

run-tests:
	go get -u github.com/Masterminds/glide
	glide install
	go test -cover `go list ./... | grep -v /vendor/`

cover:
	go test -cover `go list ./... | grep -v /vendor/`

cover-old:
	#go test -coverprofile=cover.tmp `go list ./... | grep -v /vendor/` && go tool cover -html=cover.tmp `go list ./... | grep -v /vendor/`

