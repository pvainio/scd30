package scd30

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"

	"periph.io/x/conn/v3/i2c"

	"github.com/sigurn/crc8"
)

type Measurement struct {
	CO2         float32
	Temperature float32
	Humidity    float32
}

type SCD30 struct {
	dev *i2c.Dev
}

var mutex sync.Mutex

func Open(bus i2c.Bus) (*SCD30, error) {
	mutex.Lock()
	defer mutex.Unlock()

	dev := &i2c.Dev{Addr: 0x61, Bus: bus}

	return &SCD30{dev: dev}, nil
}

// StartMeasurements starts continous measerements at given interval seconds
func (dev SCD30) StartMeasurements(interval uint16) error {
	mutex.Lock()
	defer mutex.Unlock()

	// Set measurement interval
	if err := dev.sendCommandArg(0x4600, interval); err != nil {
		return err
	}

	// Start continuous measurements
	if err := dev.sendCommandArg(0x0010, 0); err != nil {
		return err
	}

	return nil
}

// StopMeasurements stops continuous measurements
func (dev SCD30) StopMeasurements() error {
	mutex.Lock()
	defer mutex.Unlock()
	return dev.sendCommand(0x0104)
}

// GetTemperatureOffset gets temperature offset to compensate internal
// heating.  Value is 1/100C
func (dev SCD30) GetTemperatureOffset() (uint16, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if err := dev.sendCommand(0x5403); err != nil {
		return 0, err
	}

	data, err := dev.readData(3)

	if err != nil {
		return 0, err
	}
	data, err = readValid16(bytes.NewBuffer(data))
	if err != nil {
		return 0, err
	} else {
		return binary.BigEndian.Uint16(data), nil
	}
}

// SetTemperatureOffset sets temperature offset to compensate internal
// heating.  Value is 1/100C
func (dev SCD30) SetTemperatureOffset(offset uint16) error {
	mutex.Lock()
	defer mutex.Unlock()
	return dev.sendCommandArg(0x5403, offset)
}

// GetMeasurement returns ready measurement.  HasMeasurement should be
// used first to check if there is one.
func (dev SCD30) GetMeasurement() (*Measurement, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if err := dev.sendCommand(0x0300); err != nil {
		return nil, err
	}

	if data, err := dev.readData(18); err != nil {
		return nil, err
	} else {
		buf := bytes.NewBuffer(data)
		co2, err := readValidFloat32(buf)
		if err != nil {
			return nil, err
		}
		temp, err := readValidFloat32(buf)
		if err != nil {
			return nil, err
		}
		hum, err := readValidFloat32(buf)
		if err != nil {
			return nil, err
		}
		return &Measurement{CO2: co2, Temperature: temp, Humidity: hum}, nil
	}
}

// HasMeasurement checks if there is ready measurement
func (dev SCD30) HasMeasurement() (bool, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if err := dev.sendCommand(0x0202); err != nil {
		return false, err
	}

	if data, err := dev.readData(3); err != nil {
		return false, err
	} else {
		if data[2] != crc(data[:2]) {
			return false, fmt.Errorf("crc error, expected %x got %x", crc(data[:2]), data[2])
		}
		return data[1] == 1, nil
	}
}

// SetAutomaticSelfCalibration, 1 on, 0 off
func (dev SCD30) SetAutomaticSelfCalibration(value uint16) error {
	mutex.Lock()
	defer mutex.Unlock()
	return dev.sendCommandArg(0x5306, value)
}

// SetForcedCalibration, co2 ppm
func (dev SCD30) SetForcedCalibration(value uint16) error {
	mutex.Lock()
	defer mutex.Unlock()
	return dev.sendCommandArg(0x5204, value)
}

func (dev SCD30) readData(len int) ([]byte, error) {

	data := make([]byte, len)

	if err := dev.dev.Tx(nil, data); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func (dev SCD30) sendCommand(cmd uint16) error {
	cmdData := make([]byte, 2)
	binary.BigEndian.PutUint16(cmdData, cmd)
	return dev.writeAndWait(cmdData)
}

func (dev SCD30) sendCommandArg(cmd uint16, arg uint16) error {
	cmdData := make([]byte, 2)
	argData := make([]byte, 2)
	binary.BigEndian.PutUint16(cmdData, cmd)
	binary.BigEndian.PutUint16(argData, arg)
	write := []byte{cmdData[0], cmdData[1], argData[0], argData[1], crc(argData)}
	return dev.writeAndWait(write)
}

func (dev SCD30) writeAndWait(data []byte) error {

	if err := dev.dev.Tx(data, nil); err != nil {
		return err
	}

	time.Sleep(4 * time.Millisecond)

	return nil
}

func crc(data []byte) byte {
	return crc8.Checksum(data, crcTable)
}

func readValid16(buf *bytes.Buffer) ([]byte, error) {
	data := buf.Next(2)
	crc8, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	if crc(data) != crc8 {
		return nil, fmt.Errorf("crc error, expected %x got %x", crc(data), crc8)
	}
	return data, nil
}

func readValidFloat32(buf *bytes.Buffer) (float32, error) {
	var out bytes.Buffer

	for i := 0; i < 2; i++ {
		if data, err := readValid16(buf); err != nil {
			return float32(math.NaN()), err
		} else {
			out.Write(data)
		}
	}
	uint := binary.BigEndian.Uint32(out.Bytes())
	return math.Float32frombits(uint), nil
}

var crcTable *crc8.Table

func init() {
	crcTable = crc8.MakeTable(crc8.Params{Poly: 0x31, Init: 0xff, RefIn: false, RefOut: false, XorOut: 0x00, Check: 0xff, Name: "Sensirion"})
}
