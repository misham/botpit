package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/misham/botpi/display"
	"github.com/misham/botpi/driver"
	"github.com/misham/botpi/face"

	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func main() {
	brightnessFlag := flag.String("brightness", "normal", "brightness level: bright, normal, dark")
	flag.Parse()

	var brightness display.Brightness
	switch *brightnessFlag {
	case "bright":
		brightness = display.BrightnessBright
	case "normal":
		brightness = display.BrightnessNormal
	case "dark":
		brightness = display.BrightnessDark
	default:
		fmt.Fprintf(os.Stderr, "unknown brightness %q, use: bright, normal, dark\n", *brightnessFlag)
		os.Exit(1)
	}

	// Initialize periph.io host
	if _, err := host.Init(); err != nil {
		log.Fatalf("periph init: %v", err)
	}

	// Open I2C bus
	bus, err := i2creg.Open("1")
	if err != nil {
		log.Fatalf("i2c open: %v", err)
	}
	defer bus.Close()

	// Initialize IS31FL3731
	dev := driver.NewDevice(bus, display.I2CAddr)
	if err := dev.Init(); err != nil {
		log.Fatalf("driver init: %v", err)
	}

	// Create display and animator
	disp := display.New(dev, brightness)
	anim := face.NewAnimator(disp)

	// Handle graceful shutdown
	stop := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigs
		close(stop)
	}()

	log.Printf("botpi started (brightness=%s)", *brightnessFlag)

	// Run animation loop (blocks until stop)
	if err := anim.Run(stop); err != nil {
		log.Printf("animation error: %v", err)
	}

	// Graceful shutdown: clear display
	log.Println("shutting down...")
	if err := dev.Shutdown(); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
