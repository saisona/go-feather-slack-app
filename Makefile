# File              : Makefile
# Author            : Alexandre Saison <alexandre.saison@inarix.com>
# Date              : 23.02.2021
# Last Modified Date: 23.02.2021
# Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>

init:
	go mod download
	go mod vendor
	go build .

test: init
	go test -v -cover -coverprofile=c.out ./...
	
coverage: test
	go tool cover -html=c.out -o coverage.html

.PHONY: all coverage test 
