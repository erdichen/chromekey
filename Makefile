all:
	go build

cgo:
	CGO_ENABLED=1 GOARCH=amd64 go build -o chromekey64_cgo
	CGO_ENABLED=1 GOARCH=386 go build -o chromekey32_cg

nocgo:
	CGO_ENABLED=0 GOARCH=amd64 go build -o chromekey64_nocg
	CGO_ENABLED=0 GOARCH=386 go build -o chromekey32_nocg

gen:
	go generate ./...

keytest:
	CGO_ENABLED=1 GOARCH=amd64 go run ./cmd/keytest
	CGO_ENABLED=1 GOARCH=386 go run  ./cmd/keytest
	CGO_ENABLED=0 GOARCH=amd64 go run  ./cmd/keytest
	CGO_ENABLED=0 GOARCH=386 go run  ./cmd/keytest

install:
	strip chromekey
	cp -f chromekey /usr/local/bin
	cp -f config/chromekey.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl start chromekey.service
