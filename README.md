# chromekey

Remap function keys to Chromebook media keys on Linux. This program uses `evdev` and `uinput` to perform the key mapping at close to kernel level. It is more flexible than HWDB scancode mapping. Unlike XKB mappings, this program work in the Linux console and the GDM login screen.

Features:

1. Default mapping of function key to the original Chromebook media keys

2. Support third-level key mapping with FN+Shift+key combinations

    1. FN+Shift+brightness up/down control keyboard backlights

    2. FN+Shift+Search toggles Cap Lock

3. Optionally take over an keyboard LED as FN key lock LED.

## Build the binary

### Install the Go SDK

Download the Go SDK from the [Downloads](https://go.dev/dl/) page.

Following the [install instructions](https://go.dev/doc/install).

### Download the Go modules

```
go mod download
```

### Build with cgo

```
go build github.com/erdichen/chromekey/cmd/chromekey
```

### Build without cgo

In case you don't have the C headers for cgo, you can build the pure Go version.

```
CGO_ENABLED=0 go build github.com/erdichen/chromekey/cmd/chromekey
```

## Configuration

First pick a key as the `FN` key. I don't use the `Lock` key (at the upper left corner) on my Chromebook, so I turned it into the `FN` key.

### FN key configuration snippet

```
fn_key:  KEY_F13
```

### Create a default configuration file with the `--dump_config` flag.

```
./chromekey --dump_config > chromekey.config
```

### Use the `--show_key` flag to find key names

```
./chromekey --show_key
```

## Installation

### Copy the binary to `/usr/local/bin`

```
sudo cp ./chromekey /usr/local/bin
```

### Add a simple systemd service unit file

```
sudo tee /etc/systemd/system/chromekey.service <<EOF
[Service]
Type=simple
ExecStart=/usr/local/bin/chromekey
EOF
```

#### Don't forget to add the `--config_file` in case you use a configuration file.

```
ExecStart=/usr/local/bin/chromekey --config_file=/usr/local/etc/chromekey.config
```

#### Optional: Use NUM_LOCK LED as FN Lock LED

NOTE: Don't use this option if you have an external USB keyboard with a numpad.

```
ExecStart=/usr/local/bin/chromekey --use_led=NUML
```

### Add a udev rule to trigger the systemd service

```
sudo tee /etc/udev/rules.d/99-chromekey.rules <<EOF
ACTION=="add", ATTRS{name}=="AT Translated Set 2 keyboard", TAG+="systemd", ENV{SYSTEMD_WANTS}="chromekey.service"
EOF
```

### Nice Addition

Install a [Lock Keys](https://extensions.gnome.org/extension/36/lock-keys/) Gnome Shell extension to show LED status on screen.
