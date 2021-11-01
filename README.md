# Sensirion SCD30 CO2 sensor i2c driver module for Golang

## Overview

With this module Sensirion SCD30 CO2 sensor can be accessed throug i2c bus.
Implemented:
- starting and stopping continuous measurements
- checking for ready measurement
- reading ready measurement
- getting and setting temperature compensation

## Example

```go
package main

import (
	"log"
	"time"

	"github.com/pvainio/scd30"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	dev, err := scd30.Open(bus)
	if err != nil {
		log.Fatal(err)
	}

	var interval uint16 = 5

	dev.StartMeasurements(interval)

	for {
		time.Sleep(time.Duration(interval) * time.Second)
		if hasMeasurement, err := dev.HasMeasurement(); err != nil {
			log.Fatalf("error %v", err)
		} else if !hasMeasurement {
			return
		}

		m, err := dev.GetMeasurement()
		if err != nil {
			log.Fatalf("error %v", err)
		}

		log.Printf("Got measure %f ppm %f%% %fC", m.CO2, m.Humidity, m.Temperature)
	}
}
```
