package arduino

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/goserial"
)

var log = logger.GetLogger("arduino")

// Arduino provides two-way communication between go and the arduino on the
// Ninja Block shield and the Ninja Pi Crust
type Arduino struct {
	sync.Mutex
	Incoming chan Message
	port     io.ReadWriteCloser
	acks     chan []DeviceData
}

type Message struct {
	Device []DeviceData `json:"device,omitempty"`
	ACK    []DeviceData `json:"ACK,omitempty"`
	Error  struct {
		Code int
	} `json:"Errormomitempty"`
}

type DeviceData struct {
	G  string
	V  int
	D  int
	DA interface{}
}

func Connect(path string, baudRate int) (arduino *Arduino, err error) {

	config := &serial.Config{Name: path, Baud: baudRate}
	conn, err := serial.OpenPort(config)
	if err != nil {
		return
	}

	arduino = &Arduino{
		Incoming: make(chan Message, 10),
		port:     conn,
		acks:     make(chan []DeviceData),
	}

	reader := bufio.NewReader(conn)
	go func() {
		for {
			str, err := reader.ReadString('\n')
			if err != nil {
				log.Warningf("Failed to read message from serial port: %s", err)
				continue
			}

			log.Infof("Incoming: %s", str)
			var msg Message
			err = json.Unmarshal([]byte(str), &msg)

			if err != nil {
				log.Warningf("Error parsing json: %s", err)
			}

			if msg.ACK != nil {
				select {
				case arduino.acks <- msg.ACK:
				default:
					log.Warningf("Got ack we weren't listening for")
				}
			}

			select {
			case arduino.Incoming <- msg:
			default:
				log.Warningf("Incoming channel is full. Ignoring message: %s", str)
			}

		}
	}()

	return
}

func (a *Arduino) GetVersion() (string, error) {
	ack, err := a.Write(Message{
		Device: []DeviceData{
			DeviceData{
				G:  "0",
				V:  0,
				D:  1003,
				DA: "VNO",
			},
		},
	})

	if err != nil {
		return "", err
	}

	return ack[0].DA.(string), nil
}

func (a *Arduino) WriteDeviceData(data ...DeviceData) error {
	_, err := a.Write(Message{
		Device: data,
	})

	return err
}

func (a *Arduino) Write(message Message) ([]DeviceData, error) {
	a.Lock()
	defer a.Unlock()

	j, _ := json.Marshal(message)

	log.Debugf("Outgoing: %s", j)

	a.port.Write(j)
	a.port.Write([]byte("\n"))

	select {
	case ack := <-a.acks:
		return ack, nil
	case <-time.After(time.Second * 2):
		return nil, fmt.Errorf("Arduino write timed out after 2 seconds")
	}

}
