package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"erdi.us/chromekey/evdev/keycode"
	"erdi.us/chromekey/remap"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Customize key mapping.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	devicePath := flag.String("device", "/dev/input/event0", "Keyboard input device")
	uinputDev := flag.String("uinput", "/dev/uinput", "User input event injection device")
	timeout := flag.Duration("timeout", 0, "Exit after seconds since last event (0=disable)")
	fnEnable := flag.Bool("fnenable", true, "Enable media keys")
	fnKey := flag.Uint("fnkey", uint(keycode.KEY_F13), "Keycode of the FN key")
	verbosity := flag.Int("v", 0, "Verbose logging level")
	grab := flag.Bool("grab", true, "Grab input")
	flag.Parse()

	sigC := make(chan os.Signal, 10)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := remap.New(ctx, *devicePath, *uinputDev, *fnEnable, keycode.Key(*fnKey), *grab)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	remap.SetVerbosity(*verbosity)

	if err := s.Start(ctx, sigC, *timeout); err != nil {
		log.Fatal(err)
	}
}
