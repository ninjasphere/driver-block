package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/driver-block/arduino"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/support"
)

var info = ninja.LoadModuleInfo("./package.json")
var log = logger.GetLogger(info.Name)

const requiredVersion = "V12_0.46x"

var path = config.String("/dev/ttyO1", "driver-block.path")
var speed = config.Int(9600, "driver-block.baud")

type Driver struct {
	support.DriverSupport
}

func NewDriver() (*Driver, error) {

	driver := &Driver{}

	err := driver.Init(info)
	if err != nil {
		log.Fatalf("Failed to initialize driver: %s", err)
	}

	err = driver.Export(driver)
	if err != nil {
		log.Fatalf("Failed to export driver: %s", err)
	}

	return driver, nil
}

func (d *Driver) Start(_ interface{}) error {

	port, err := arduino.Connect(path, speed)

	if err != nil {
		log.Fatalf("Couldn't connect to arduino: %s", err)
	}

	version, err := port.GetVersion()

	if err != nil {
		log.Warningf("Failed to get version from arduino. Continuing anyway. #YOLO.")
	}

	if version != requiredVersion {
		log.Warningf("Unknown arduino version. Expected:%s Got: %s", requiredVersion, version)
	}

	NewLight(d, 1007, "Nina's Eyes", port)
	NewLight(d, 999, "Status Light", port)

	go func() {
		for message := range port.Incoming {
			spew.Dump("incoming", message)
		}
	}()

	return nil
}
