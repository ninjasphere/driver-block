package arduino

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
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
	Device []DeviceData `json:"device"`
	ACK    []DeviceData
	Error  struct {
		Code int
	}
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

			log.Infof("Json: %s", str)
			var msg Message
			err = json.Unmarshal([]byte(str), &msg)

			if err != nil {
				log.Warningf("Error reading serial port: %s", err)
			} else {
				spew.Dump(msg)
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

func (a *Arduino) Write(message *Message) error {
	a.Lock()
	defer a.Unlock()

	j, _ := json.Marshal(message)

	a.port.Write(j)
	a.port.Write([]byte("\n"))

	select {
	case ack := <-a.acks:
		// Check data equals what we sent?
		spew.Dump("GOT ACK", ack)
		return nil
	case <-time.After(time.Second * 2):
		return fmt.Errorf("Arduino write timed out after 2 seconds")
	}

}
