all:
	go build erdi.us/chromekey/cmd/chromekey

gen:
	go generate ./...

install:
	strip chromekey
	cp -f chromekey /usr/local/bin
	cp -f config/chromekey.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl start chromekey.service
