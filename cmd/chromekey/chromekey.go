package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"erdi.us/chromekey/evdev/keycode"
	"erdi.us/chromekey/remap"
	"erdi.us/chromekey/remap/config"
	"google.golang.org/protobuf/encoding/prototext"
)

func main() {
	var fnKey keycode.Code
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Customize key mapping.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	devicePath := flag.String("device", "/dev/input/event0", "Keyboard input device")
	uinputDev := flag.String("uinput", "/dev/uinput", "User input event injection device")
	timeout := flag.Duration("timeout", 0, "Exit after seconds since last event (0=disable)")
	verbosity := flag.Int("v", 0, "Verbose logging level")
	grab := flag.Bool("grab", true, "Grab input")
	cfgFile := flag.String("config_file", "", "Configuration file")
	dumpConfig := flag.Bool("dump_config", false, "Dump configuration file")
	flag.Func("fnkey", "Keycode of the FN key", func(value string) error {
		key, ok := keycode.Code_value[value]
		if !ok {
			return fmt.Errorf("invalid key: %v", value)
		}
		fnKey = keycode.Code(key)
		return nil
	})
	flag.Parse()

	sigC := make(chan os.Signal, 10)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := remap.New(ctx, *devicePath, *uinputDev, fnKey, *grab)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	remap.SetVerbosity(*verbosity)

	if *cfgFile != "" {
		b, err := ioutil.ReadFile(*cfgFile)
		if err != nil {
			log.Fatal(err)
		}
		var pb config.KeymapConfig
		if err := prototext.Unmarshal(b, &pb); err != nil {
			log.Fatal(err)
		}
		var cfg = remap.KeymapConfig{
			FnENabled:   pb.FnEnabled,
			FnKey:       keycode.Code(pb.FnKey),
			KeyMap:      config.FromPBKeymap(pb.KeyMap),
			ShiftKeyMap: config.FromPBKeymap(pb.ShiftKeyMap),
		}
		if fnKey != keycode.Code_KEY_RESERVED {
			// Overrides FN key in config from flag value.
			cfg.FnKey = fnKey
		}
		s.SetConfig(cfg)
	}

	if *dumpConfig {
		cfg := s.Config()
		var pb = config.KeymapConfig{
			FnEnabled: cfg.FnENabled,
			FnKey:     keycode.Code_KEY_FN,
		}
		pb.KeyMap = config.ToPBKeymap(cfg.KeyMap)
		pb.ShiftKeyMap = config.ToPBKeymap(cfg.ShiftKeyMap)

		if _, err := os.Stdout.WriteString((prototext.MarshalOptions{Indent: "  "}).Format(&pb)); err != nil {
			log.Fatal(err)
		}
	}

	if err := s.Start(ctx, sigC, *timeout); err != nil {
		log.Fatal(err)
	}
}
