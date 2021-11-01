package scd30

import (
	"testing"

	"periph.io/x/conn/v3/i2c/i2ctest"
)

const addr = 0x61

func TestOpen(t *testing.T) {
	scd30, _ := Open(&i2ctest.Record{})

	if scd30.dev.Addr != 0x61 {
		t.Fatalf("Invalid addr %v", scd30.dev.Addr)
	}
}

func TestGetTemperatureOffset(t *testing.T) {

	bus := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: addr, W: []byte{0x54, 0x03}, R: nil},
			{Addr: addr, W: nil, R: []byte{0x01, 0x23, 0xa0}},
		},
	}

	scd30, _ := Open(bus)
	o, err := scd30.GetTemperatureOffset()
	assertNoError(t, err)
	if o != 0x123 {
		t.Fatalf("Got incorrect offset %v should be %v", o, 0x123)
	}
}

func TestSetTemperatureOffset(t *testing.T) {

	bus := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: addr, W: []byte{0x54, 0x03, 0x1, 0x23, 0xa0}, R: nil},
		},
	}

	scd30, _ := Open(bus)
	err := scd30.SetTemperatureOffset(0x123)
	assertNoError(t, err)
}

func TestSetAutomaticSelfCalibration(t *testing.T) {

	bus := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: addr, W: []byte{0x53, 0x06, 0x0, 0x0, 0x81}, R: nil},
		},
	}

	scd30, _ := Open(bus)
	err := scd30.SetAutomaticSelfCalibration(0)
	assertNoError(t, err)
}

func TestHasMeasurement(t *testing.T) {
	bus := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: addr, W: []byte{0x02, 0x02}, R: nil},
			{Addr: addr, W: nil, R: []byte{0x00, 0x01, 0xb0}},
		},
	}

	scd30, _ := Open(bus)
	o, err := scd30.HasMeasurement()
	assertNoError(t, err)
	if !o {
		t.Fatalf("expected true")
	}
}

func TestGetMeasurement(t *testing.T) {

	bus := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: addr, W: []byte{0x03, 0x00}, R: nil},
			{Addr: addr, W: nil, R: []byte{0x3f, 0x8c, 0xad, 0xcc, 0xcd, 0x94, 0x40, 0xc, 0x75, 0xcc, 0xcd, 0x94, 0x40, 0x53, 0x25, 0x33, 0x33, 0x88}},
		},
	}

	scd30, _ := Open(bus)
	o, err := scd30.GetMeasurement()
	assertNoError(t, err)
	assertFloat(t, 1.1, o.CO2)
	assertFloat(t, 2.2, o.Temperature)
	assertFloat(t, 3.3, o.Humidity)
}

func TestStartMeasurements(t *testing.T) {
	bus := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: addr, W: []byte{0x46, 0x00, 0x01, 0x23, 0xa0}, R: nil},
			{Addr: addr, W: []byte{0x00, 0x10, 0x00, 0x00, 0x81}, R: nil},
		},
	}

	scd30, _ := Open(bus)
	err := scd30.StartMeasurements(0x123)
	assertNoError(t, err)
}

func assertFloat(t *testing.T, expected float32, value float32) {
	if expected != value {
		t.Fatalf("Expected %f got %f", expected, value)
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Got error %v", err)
	}
}
