all:
	go build erdi.us/chromekey/cmd/chromekey

cgo:
	CGO_ENABLED=1 GOARCH=amd64 go build -o chromekey64_cgo erdi.us/chromekey/cmd/chromekey
	CGO_ENABLED=1 GOARCH=386 go build -o chromekey32_cgo erdi.us/chromekey/cmd/chromekey
	CGO_ENABLED=1 GOARCH=amd64 go build -o keytest64_cgo erdi.us/chromekey/cmd/keytest
	CGO_ENABLED=1 GOARCH=386 go build -o keytest32_cgo erdi.us/chromekey/cmd/keytest

nocgo:
	CGO_ENABLED=0 GOARCH=amd64 go build -o chromekey64_nocgo erdi.us/chromekey/cmd/chromekey
	CGO_ENABLED=0 GOARCH=386 go build -o chromekey32_nocgo erdi.us/chromekey/cmd/chromekey
	CGO_ENABLED=0 GOARCH=amd64 go build -o keytest64_nocgo erdi.us/chromekey/cmd/keytest
	CGO_ENABLED=0 GOARCH=386 go build -o keytest32_nocgo erdi.us/chromekey/cmd/keytest

gen:
	go generate ./...

install:
	strip chromekey
	cp -f chromekey /usr/local/bin
	cp -f config/chromekey.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl start chromekey.service
