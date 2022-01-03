# Chromekey - simulate media keys for Chromebook

## Generate event and key names

```
go install golang.org/x/tools/cmd/stringer@latest
```

```
go generate ./...
```

or 

```
make gen
```

## Build

```
make
```

## Install

```
sudo cp chromekey /usr/local/bin/
sudo strip /usr/local/bin/chromekey
```

## Configuration

```
sudo tee /etc/udev/rules.d/99-chromekey.rules <<EOF
ACTION=="add", ATTRS{name}=="AT Translated Set 2 keyboard", TAG+="systemd", ENV{SYSTEMD_WANTS}="chromekey.service"
EOF
```

```
sudo tee /etc/systemd/system/chromekey.service <<EOF
[Service]
Type=simple
ExecStart=/usr/local/bin/chromekey
EOF
```
