GITHUB_API_TOKEN := ""
VERSION :=""

run-tests:
	go get -u github.com/Masterminds/glide
	glide install
	go test -cover -v

cover:
	go test -coverprofile=cover.tmp && go tool cover -html=cover.tmp
