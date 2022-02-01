# chromekey

Remap function keys to Chromebook media keys on Linux. This program uses `evdev` and `uinput` to perform the key mapping at close to kernel level. It is more flexible than HWDB scancode mapping. Unlike XKB mappings, this program work in the Linux console and the GDM login screen.

Features:

1. Default mapping of function key to the original Chromebook media keys

    1. Press the FN key to toggle the FN locked mode

    2. Press FN+key to select the second level alternate key

2. Support third-level key mapping with FN+Shift+key combinations

    1. FN+Shift+brightness up/down control keyboard backlights

3. Optionally take over an keyboard LED as the FN key lock LED.

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
go build
```

### Build without cgo

In case you don't have the C headers for cgo, you can build the pure Go version.

```
CGO_ENABLED=0 go build
```

## Configuration

First pick a key as the `FN` key. I don't use the `Lock` key (at the upper left corner) on my Chromebook, so I turned it into the `FN` key.

### Create a default configuration file with the `-dump_config` flag.

```
./chromekey -dump_config > chromekey.config
```

### Use the `-show_key` flag to find key names

Stop any running instance to release the grab on the keyboard device first.

```
./chromekey -show_key
```

### FN key configuration snippet

```
fn_key:  KEY_F13
```

### Set the FN lock state on start up

```
fn_enabled: true
```

### Set third-level keys for `FN+Shift+` key maps

```
third_level_key:  KEY_LEFTSHIFT
third_level_key:  KEY_RIGHTSHIFT
```

### Map third-level keys for keyboard backlight

Use `FN+Shift+F6/F7` for keyboard backlight adjustment:

```
third_level_key_map:  {
  from:  KEY_F6
  to:  KEY_KBDILLUMDOWN
}
third_level_key_map:  {
  from:  KEY_F7
  to:  KEY_KBDILLUMUP
}
```

### Add additional key map that uses the FN key as a modifier

For example, press FN+backspace to send the DELETE key:

```
mod_key_map:  {
  from:  KEY_BACKSPACE
  to:  KEY_DELETE
}
```

#### Optional: Use the Num Lock LED as the FN Lock LED

NOTE: Don't use this option if you have an external USB keyboard with a numpad.

```
use_led: NUML
```

## Installation

### Copy the binary to `/usr/local/bin`

```
sudo cp ./chromekey /usr/local/bin
```

### Copy the configuration file to `/usr/local/etc`

```
sudo cp chromekey.config /usr/local/etc/
```

### Add a simple systemd service unit file

```
sudo tee /etc/systemd/system/chromekey.service <<EOF
[Service]
Type=simple
ExecStart=/usr/local/bin/chromekey -config_file=/usr/local/etc/chromekey.config
EOF
```

### Add a udev rule to trigger the systemd service

```
sudo tee /etc/udev/rules.d/99-chromekey.rules <<EOF
ACTION=="add", ATTRS{name}=="AT Translated Set 2 keyboard", TAG+="systemd", ENV{SYSTEMD_WANTS}="chromekey.service"
EOF
```

### Optional: Install the bash completion file

```
sudo cp config/chromekey.completion /etc/bash_completion.d/chromekey
```

### Nice Addition

Install a [Lock Keys](https://extensions.gnome.org/extension/36/lock-keys/) Gnome Shell extension to show LED status on screen.

