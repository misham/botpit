package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/misham/botpi/brightness"
	"github.com/misham/botpi/display"
	"github.com/misham/botpi/driver"
	"github.com/misham/botpi/face"

	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func main() {
	brightnessFlag := flag.String("brightness", "auto", "brightness level: auto, bright, normal, dark")
	lat := flag.Float64("lat", 0, "latitude for auto brightness (required with -brightness auto)")
	lon := flag.Float64("lon", 0, "longitude for auto brightness (required with -brightness auto)")
	flag.Parse()

	var bright display.Brightness
	dynamicBrightness := *brightnessFlag == "auto"

	if !dynamicBrightness {
		switch *brightnessFlag {
		case "bright":
			bright = display.BrightnessBright
		case "normal":
			bright = display.BrightnessNormal
		case "dark":
			bright = display.BrightnessDark
		default:
			fmt.Fprintf(os.Stderr, "unknown brightness %q, use: auto, bright, normal, dark\n", *brightnessFlag)
			os.Exit(1)
		}
	} else {
		if *lat == 0 && *lon == 0 {
			fmt.Fprintln(os.Stderr, "auto brightness requires -lat and -lon flags")
			os.Exit(1)
		}
		bright = display.BrightnessNormal // initial value until controller updates
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
	disp := display.New(dev, bright)
	words := loadWordsFile()
	anim := face.NewAnimator(disp, words)

	// Handle graceful shutdown
	stop := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigs
		close(stop)
	}()

	// Start dynamic brightness controller if in auto mode
	if dynamicBrightness {
		loc := brightness.Location{Lat: *lat, Lon: *lon}
		ctrl := brightness.NewController(disp, loc, brightness.DefaultWeatherURL)
		go ctrl.Run(stop)
	}

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

// loadWordsFile reads words.json from next to the executable.
// Returns nil if the file doesn't exist or is invalid.
func loadWordsFile() []string {
	exe, err := os.Executable()
	if err != nil {
		return nil
	}
	path := filepath.Join(filepath.Dir(exe), "words.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var wl struct {
		Words []string `json:"words"`
	}
	if err := json.Unmarshal(data, &wl); err != nil {
		log.Printf("invalid words.json: %v", err)
		return nil
	}
	return wl.Words
}
