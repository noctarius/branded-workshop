package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.golang/paho"
	"io"
	"math"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type message struct {
	Timestamp time.Time `json:"timestamp"`
	TypeId    int8      `json:"type"`
	Value     float64   `json:"value"`
}

func main() {
	fmt.Printf("Opening data.bin...\n")
	i, err := os.Open("./data.bin")
	if err != nil {
		panic(err)
	}
	defer i.Close()

	fmt.Printf("Setting up OS Signals...\n")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	fmt.Printf("Creating MQTT publisher...\n")
	client, err := newPublisher()
	if err != nil {
		panic(err)
	}

	go func() {
		fmt.Printf("Streaming data.bin...\n")

		timestamp, typeId, value, err := read(i)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			panic(err)
		}

		for {
			select {
			case <-sigs:
				done <- true
				break
			default:
			}

			now := time.Now()
			comp := time.Date(now.Year(), now.Month(), now.Day(), timestamp.Hour(), timestamp.Minute(), timestamp.Second(), 0, now.Location())
			if now.After(comp) && now.Add(-time.Second).Before(comp) {
				fmt.Printf("Matching record: timestamp=%v, type=%d, value=%0.2f\n", timestamp, typeId, value)

				// send mqtt
				msg := message{
					Timestamp: timestamp,
					TypeId:    typeId,
					Value:     value,
				}
				data, err := json.Marshal(msg)
				if err != nil {
					panic(err)
				}
				if err := client.publish(data); err != nil {
					fmt.Printf("Error: %v\n", err)
					panic(err)
				}

				// read next record
				timestamp, typeId, value, err = read(i)
				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Printf("Error: %v\n", err)
					panic(err)
				}
			} else if comp.Before(now.Add(-time.Second)) || comp.After(now.Add(time.Minute)) {
				// skip record
				timestamp, typeId, value, err = read(i)
				if err != nil {
					if err == io.EOF {
						return
					}
					fmt.Printf("Error: %v\n", err)
					panic(err)
				}
			}
		}
		done <- true
	}()

	<-done
}

func read(reader io.Reader) (timestamp time.Time, typeId int8, value float64, err error) {
	buffer := make([]byte, 17, 17)
	if _, err = reader.Read(buffer); err != nil {
		return
	}

	v := binary.LittleEndian.Uint64(buffer[0:])
	millis := int64(v)
	timestamp = time.UnixMilli(millis)
	typeId = int8(buffer[8])
	v = binary.LittleEndian.Uint64(buffer[9:])
	value = math.Float64frombits(v)
	return
}

type pahoPublisher struct {
	client *mqtt.Client
}

func newPublisher() (*pahoPublisher, error) {
	pub := &pahoPublisher{}
	pub.reconnect()

	return pub, nil
}

func (p *pahoPublisher) reconnect() {
	conn, err := net.Dial("tcp", "10.96.1.25:1883")
	if err != nil {
		panic(err)
	}

	p.client = mqtt.NewClient(mqtt.ClientConfig{})
	p.client.Conn = conn

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connack, err := p.client.Connect(ctx, &mqtt.Connect{
		CleanStart:   true,
		PasswordFlag: false,
		UsernameFlag: false,
		KeepAlive:    20,
	})

	if err != nil {
		panic(err)
	}

	if connack.ReasonCode != 0 {
		panic(err)
	}
}

func (p *pahoPublisher) publish(message []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := p.client.Publish(ctx, &mqtt.Publish{
		Topic:   "/metrics",
		QoS:     1,
		Retain:  true,
		Payload: message,
	}); err != nil {
		return err
	}
	return nil
}

func (p *pahoPublisher) close() {
	p.client.Disconnect(&mqtt.Disconnect{})
}
