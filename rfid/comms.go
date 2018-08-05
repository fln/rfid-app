package rfid

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// List of know status values. Status value is returned when executing device
// commands in raw mode, with RawCommand() method.
const (
	StatusOK    byte = 0x00
	StatusNoTag      = 0x01
)

// Command is an enum type for encoding different commands that can be sent to
// device.
type Command uint16

// List of supported commands.
const (
	CommandInfo   Command = 0x0102
	CommandBeep           = 0x0103
	CommandLed            = 0x0104
	CommandRead           = 0x010C
	CommandWrite2         = 0x020C
	CommandWrite3         = 0x030C
)

// LedMode encodes device led status - off, red or green.
type LedMode byte

// List of LED modes accepted by LED change command.
const (
	LedOff   LedMode = 0x00
	LedRed           = 0x01
	LedGreen         = 0x02
)

// beepUnit is a minimum beep duration supported by this device. When sending
// beep command to the device beep duration is sent as a number of beep units.
const beepUnit = time.Second / 255

// msgPrefix is a static header added for all messages (sent and received) when
// communicating with device.
const msgPrefix uint16 = 0xAADD

func xorChecksum(buf []byte) byte {
	var csum byte
	for _, b := range buf {
		csum ^= b
	}
	return csum
}

func copyUint16(buf []byte, n uint16) int {
	binary.BigEndian.PutUint16(buf, n)
	return 2
}

func newRequest(cmd Command, data []byte) []byte {
	// header+length+command+data+checksum
	buf := make([]byte, 2+2+2+len(data)+1)

	pos := 0
	pos += copyUint16(buf[pos:], msgPrefix)             // header
	pos += copyUint16(buf[pos:], uint16(2+len(data)+1)) // length
	pos += copyUint16(buf[pos:], uint16(cmd))           // command
	pos += copy(buf[pos:], data)                        // data
	buf[pos] = xorChecksum(buf[4:pos])                  // checksum

	return buf
}

func rx(r io.Reader) ([]byte, error) {
	header := make([]byte, 4)
	pos, err := io.ReadAtLeast(r, header, len(header))
	if err != nil {
		return nil, fmt.Errorf("header read error: %s", err)
	}
	if binary.BigEndian.Uint16(header) != msgPrefix {
		return nil, errors.New("bad response header")
	}
	length := binary.BigEndian.Uint16(header[2:])
	if length < 4 {
		return nil, errors.New("invalid payload length value")
	}
	buf := make([]byte, len(header)+int(length))
	copy(buf, header)
	_, err = io.ReadAtLeast(r, buf[pos:], int(length))
	if err != nil {
		return nil, fmt.Errorf("payload read error: %s", err)
	}
	if buf[len(buf)-1] != xorChecksum(buf[4:len(buf)-1]) {
		return nil, errors.New("response have invalid checksum")
	}
	return buf, nil
}
