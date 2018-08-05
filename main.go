package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fln/rfid-app/rfid"
)

var silent bool

func emitOK(d *rfid.Device)    { beepWithLed(d, 50*time.Millisecond, rfid.LedGreen) }
func emitError(d *rfid.Device) { beepWithLed(d, 200*time.Millisecond, rfid.LedRed) }

func beepWithLed(d *rfid.Device, t time.Duration, color rfid.LedMode) {
	if silent {
		return
	}
	if err := d.ChangeLed(color); err != nil {
		log.Print(err)
		return
	}
	if err := d.Beep(t); err != nil {
		log.Print(err)
		return
	}
	if err := d.ChangeLed(rfid.LedOff); err != nil {
		log.Print(err)
		return
	}
}

func readOnce(d *rfid.Device) (id []byte, err error) {
	for {
		id, err = d.ReadTag()
		if err != rfid.ErrNoTag {
			break
		}
	}
	return id, err
}

func infoMode(d *rfid.Device) {
	model, err := d.Info()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(model)
}

func readMode(d *rfid.Device) {
	id, err := readOnce(d)
	if err != nil {
		emitError(d)
		log.Fatal(err)
	}
	fmt.Println(hex.EncodeToString(id))
	emitOK(d)
}

func readLoopMode(d *rfid.Device) {
	var lastID []byte
	for {
		id, err := readOnce(d)
		if err != nil {
			// suppress read errors in loop mode
			continue
		}
		if bytes.Equal(id, lastID) {
			continue
		}
		fmt.Println(hex.EncodeToString(id))
		emitOK(d)
		lastID = id
	}
}

func main() {
	var dev string
	var mode string

	flag.StringVar(&dev, "dev", "/dev/ttyUSB0", "RFID read/writer serial interface device")
	flag.StringVar(&mode, "mode", "read", "Application mode, one of: read, read-loop, info")
	flag.BoolVar(&silent, "silent", false, "Skip beeps and LED flashes, reduces number of commands sent to the reader")
	flag.Parse()

	d, err := rfid.OpenDevice(dev, false)
	if err != nil {
		log.Fatal(err)
	}

	if !silent {
		if err := d.ChangeLed(rfid.LedOff); err != nil {
			log.Fatal("switching LED off", err)
		}
	}

	switch mode {
	case "info":
		infoMode(d)
	case "read":
		readMode(d)
	case "read-loop":
		readLoopMode(d)
	default:
		flag.PrintDefaults()
		os.Exit(2)
	}
}
