// Binary remaps function keys to Chromebook media keys on Linux.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/erdichen/chromekey/evdev"
	"github.com/erdichen/chromekey/evdev/keycode"
	"github.com/erdichen/chromekey/log"
	"github.com/erdichen/chromekey/remap"
	"github.com/erdichen/chromekey/remap/config"
	"google.golang.org/protobuf/encoding/prototext"
)

const description = `Emulates a FN key to convert functions key to media keys. Choose any valid keycode as the FN key.

  1. Press FN key once to toggle between media key and function key modes.
  2. Press FN+Shift+key to use third level key mapping.
  3. Run '%s led' to list LED names.
  4. Run '%s key' to list key names.

`

func checkCmd() bool {
	switch flag.Arg(0) {
	case "key":
		for i := int32(0); i < int32(keycode.Code_KEY_MAX); i++ {
			n := keycode.Code_name[i]
			if n != "" {
				fmt.Printf("  %s\n", n)
			}
		}
	case "led":
		for i := int32(0); i < int32(keycode.LED_MAX); i++ {
			n := keycode.LED_name[i]
			if n != "" {
				fmt.Printf("  %s\n", n)
			}
		}
	default:
		return false
	}
	return true
}

var verbosity = flag.Int("v", 0, "Verbose logging level")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), description, os.Args[0], os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
	}
	devicePath := flag.String("input_device", "", "Keyboard input device")
	inputDevDir := flag.String("evdev_dir", "/dev/input", "Keyboard input device directory")
	uinputDev := flag.String("uinput", "/dev/uinput", "User input event injection device")
	timeout := flag.Duration("timeout", 0, "Exit after seconds since last event (0=disable)")
	grab := flag.Bool("grab", true, "Grab evdev input device")
	cfgFile := flag.String("config_file", "", "Configuration file")
	dumpConfig := flag.Bool("dump_config", false, "Dump configuration file")
	useDefault := flag.Bool("use_default", true, "Use default configuration if config_file is not set")
	showKey := flag.Bool("show_key", false, "Show keycodes only and don't remap or forward the keys")
	fnKey := keycode.Code_KEY_RESERVED
	flag.Func("fnkey", "Keycode of the FN key (default KEY_FN13)", func(value string) error {
		key, ok := keycode.Code_value[value]
		if !ok || key >= int32(keycode.Code_KEY_MAX) {
			return fmt.Errorf("invalid key: %v", value)
		}
		fnKey = keycode.Code(key)
		return nil
	})
	useLED := keycode.LED_CNT
	flag.Func("use_led", "Use LED as FN key indicator", func(value string) error {
		led, ok := keycode.LED_value[value]
		if !ok || led >= int32(keycode.LED_MAX) {
			return fmt.Errorf("invalid LED: %v", value)
		}
		useLED = keycode.LED(led)
		return nil
	})
	flag.Parse()

	if checkCmd() {
		return
	}

	if *showKey {
		*verbosity += 3
		*grab = false
	}

	remap.SetVerbosity(*verbosity)

	sigC := make(chan os.Signal, 10)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM, syscall.SIGTSTP, syscall.SIGCONT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cfg config.RunConfig

	// Loads configuration from the given file if the flag is valid.
	if *cfgFile != "" {
		b, err := ioutil.ReadFile(*cfgFile)
		if err != nil {
			log.Fatalf("failed to open configuration file: %v", err)
		}
		var pb config.KeymapConfig
		if err := prototext.Unmarshal(b, &pb); err != nil {
			log.Fatalf("failed to marshal configuration proto: %v", err)
		}
		cfg = config.FromPBConfig(&pb)
	} else if *useDefault {
		cfg = config.DefaultRunConfig()
	}

	if useLED != keycode.LED_CNT {
		cfg.UseLED = useLED
	}

	if fnKey != keycode.Code_KEY_RESERVED {
		// Overrides FN key in config from flag value.
		cfg.FnKey = fnKey
	}

	// Dump the configuration and exit. Use this flag to create new default configuration file.
	if *dumpConfig {
		pb := config.ToPBConfig(cfg)
		if _, err := os.Stdout.WriteString((prototext.MarshalOptions{Indent: "  "}).Format(pb)); err != nil {
			log.Fatalf("failed to dump configuration file: %v", err)
		}
		return
	}

	// Opens an evdev device if the devicePath flag is valid.
	var in *evdev.Device
	if *devicePath != "" {
		d, err := evdev.OpenDevice(*devicePath)
		if err != nil {
			log.Fatalf("failed to create open evdev device: %v", err)
		}
		if d.IsKeyboard() {
			in = d
		}
	}
	// If devicePath does not specify a valid device, try to open an input device in the inputDevDir directory.
	if in == nil {
		d, err := openInputDevice(*inputDevDir)
		if err != nil {
			log.Fatalf("failed to create open evdev device: %v", err)
		}
		in = d
	}

	if *showKey {
		readAndPrintKeys(ctx, in, sigC)
		return
	}

	// Create new remapper instance.
	s, err := remap.New(ctx, in, *uinputDev, cfg, *grab)
	if err != nil {
		log.Fatalf("failed to create key remapper: %v", err)
	}
	defer s.Close()

	evC := remap.StartReadEventsLoop(ctx, in)

	// Start the remapper event loop.
	if err := s.Start(ctx, sigC, evC, *timeout); err != nil {
		log.Fatalf("key remapper stopped: %v", err)
	}
}

// readAndPrintKeys prints keycodes to help with writing the configuration file.
func readAndPrintKeys(ctx context.Context, in *evdev.Device, sigC chan os.Signal) {
	defer in.Close()
	evC := remap.StartReadEventsLoop(ctx, in)
	done := false
	for !done {
		select {
		case sig := <-sigC:
			switch sig {
			case syscall.SIGTSTP:
				fmt.Printf("·Suspend breaks keyboard input. Press Ctrl-C to exit!\n")
			default:
				done = true
			}
		case events, ok := <-evC:
			if !ok {
				done = true
				break
			}
			for _, v := range events {
				fmt.Printf("%s\n", v.String())
			}
		}
	}
}

// openInputDevice opens event devices in a directory to find the first keyboard device.
func openInputDevice(devDir string) (*evdev.Device, error) {
	files, err := ioutil.ReadDir(devDir)
	if err != nil {
		return nil, err
	}

	for _, v := range files {
		if !strings.HasPrefix(v.Name(), "event") {
			continue
		}
		file := filepath.Join(devDir, v.Name())
		dev, e := evdev.OpenDevice(file)
		if e != nil {
			err = e
			continue
		}
		if dev.IsKeyboard() {
			if *verbosity > 1 {
				log.Infof("Opened keyboard input device: %v", file)
			}
			return dev, nil
		}
		if *verbosity > 1 {
			log.Infof("Skipped non-keyboard input device: %v", file)
		}
		if err := dev.Close(); err != nil {
			log.Errorf("failed to close an evdev device: %v", err)
		}
	}

	if err == nil {
		err = errors.New("found no input device")
	}

	return nil, err
}
