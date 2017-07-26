GITHUB_API_TOKEN := ""
VERSION :=""

run-tests:
	go get -u github.com/Masterminds/glide
	glide install
	go test -cover `go list ./... | grep -v /vendor/`

cover:
	go test -cover `go list ./... | grep -v /vendor/`

cover-old:
	#go test -coverprofile=cover.tmp `go list ./... | grep -v /vendor/` && go tool cover -html=cover.tmp `go list ./... | grep -v /vendor/`
