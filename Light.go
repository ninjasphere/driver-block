package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/driver-block/arduino"
	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/devices"
	"github.com/ninjasphere/go-ninja/model"
)

func NewLight(driver *Driver, D int, name string, port *arduino.Arduino) error {
	light, err := devices.CreateLightDevice(driver, &model.Device{
		NaturalID:     fmt.Sprintf("%s-%d", config.Serial(), D),
		NaturalIDType: "block-arduino",
		Name:          &name,
		Signatures: &map[string]string{
			"ninja:manufacturer": "Ninja Blocks Inc",
			"ninja:productType":  "Light",
			"ninja:thingType":    "light",
		},
	}, driver.Conn)

	if err != nil {
		log.FatalError(err, "Could not create light device")
	}

	if err := light.EnableOnOffChannel(); err != nil {
		log.FatalError(err, "Could not enable hue on-off channel")
	}

	if err := light.EnableBrightnessChannel(); err != nil {
		log.FatalError(err, "Could not enable hue brightness channel")
	}

	if err := light.EnableColorChannel("temperature", "hue"); err != nil {
		log.FatalError(err, "Could not enable color channel")
	}

	port.OnDeviceData(func(data arduino.DeviceData) {
		if data.D == D {
			spew.Dump("Light Data!", data)
		}
	})

	return nil
}
