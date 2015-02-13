package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/driver-block/arduino"
	"github.com/ninjasphere/go-ninja/logger"
)

const requiredVersion = "V12_0.46x"
const path = "/dev/ttyO1"
const speed = 9600

var log = logger.GetLogger("driver-block")

func main() {

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

	go func() {
		for message := range port.Incoming {
			spew.Dump("incoming", message)
		}
	}()

	col := arduino.Message{
		Device: []arduino.DeviceData{
			arduino.DeviceData{
				G:  "0",
				V:  0,
				D:  1007, // 999 = status
				DA: "FF00FF",
			},
		},
	}

	col2 := arduino.Message{
		Device: []arduino.DeviceData{
			arduino.DeviceData{
				G:  "0",
				V:  0,
				D:  1007, // 999 = status
				DA: "00FF00",
			},
		},
	}

	col3 := arduino.Message{
		Device: []arduino.DeviceData{
			arduino.DeviceData{
				G:  "0",
				V:  0,
				D:  1007, // 999 = status
				DA: "0000FF",
			},
		},
	}

	for {
		port.Write(col)
		port.Write(col2)
		port.Write(col3)
	}

	select {}
}
