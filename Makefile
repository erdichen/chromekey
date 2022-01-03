all:
	go build erdi.us/chromekey/cmd/chromekey

gen:
	go generate ./...
