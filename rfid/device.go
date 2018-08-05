package rfid

import (
	"errors"
	"fmt"
	"time"

	"github.com/tarm/serial"
)

// Device represents an open connection to RFID reader/writer device and
// provides methods for synchronous communication with the device.
type Device struct {
	port  *serial.Port
	debug bool
}

// ErrNoTag is returned by ReadTag() if device did not detect a tag.
var ErrNoTag = errors.New("rfid: no tag detected")

// OpenDevice establishes communication channel with RFID reader/writer device
// and returns new Device instance.
func OpenDevice(dev string, commsDebug bool) (*Device, error) {
	port, err := serial.OpenPort(&serial.Config{
		Name: dev,
		Baud: 38400,
	})
	if err != nil {
		return nil, fmt.Errorf("rfid: error opening serial port: %s", err)
	}
	return &Device{
		port:  port,
		debug: commsDebug,
	}, nil
}

// Info reads device model information.
func (d *Device) Info() (string, error) {
	status, answer, err := d.RawCommand(CommandInfo, nil)
	if err != nil {
		return "", err
	}
	if status != StatusOK {
		return "", fmt.Errorf("rfid: received unexpected status %d", status)
	}
	return string(answer), nil
}

// Beep emits a beeping signal for a given duration. This function will block
// until beep finishes. Passing 0 duration will beep forever.
func (d *Device) Beep(duration time.Duration) error {
	units := int(duration / beepUnit)
	if duration != 0 && units == 0 {
		// Round-up to the nearest beep unit, because 0 beeps forever
		units = 1
	}
	if units > 255 {
		// Round-down to the max unit value
		units = 255
	}
	status, _, err := d.RawCommand(CommandBeep, []byte{byte(units)})
	if err != nil {
		return err
	}
	if status != StatusOK {
		return fmt.Errorf("rfid: received unexpected status %d", status)
	}
	return nil
}

// ChangeLed switches device led to a given mode - off, red, green.
func (d *Device) ChangeLed(mode LedMode) error {
	status, _, err := d.RawCommand(CommandLed, []byte{byte(mode)})
	if err != nil {
		return err
	}
	if status != StatusOK {
		return fmt.Errorf("rfid: received unexpected status %d", status)
	}
	return nil
}

// ReadTag probes for an NFC tag and returns tag data if one is found. This
// method will return `ErrNoTag` if NFC tag was not detected.
func (d *Device) ReadTag() ([]byte, error) {
	status, answer, err := d.RawCommand(CommandRead, nil)
	if err != nil {
		return nil, err
	}
	switch status {
	case StatusOK:
		return answer, nil
	case StatusNoTag:
		return nil, ErrNoTag
	default:
		return nil, fmt.Errorf("rfid: received unexpected status %d, answer %v", status, answer)
	}
}

// RawCommand is a lower-level interface allowing to send custom commands to the
// device. Other higher level interface functions are using RawCommand
// internally.
func (d *Device) RawCommand(cmd Command, data []byte) (status byte, answer []byte, err error) {
	req := newRequest(cmd, data)
	if d.debug {
		fmt.Println("TX:", req)
	}
	if _, err := d.port.Write(req); err != nil {
		return 0, nil, fmt.Errorf("rfid: error sending command: %s", err)
	}
	resp, err := rx(d.port)
	if err != nil {
		return 0, nil, fmt.Errorf("rfid: error reading response: %s", err)
	}
	if d.debug {
		fmt.Println("RX:", resp)
	}

	return resp[6], resp[7 : len(resp)-1], nil

}
