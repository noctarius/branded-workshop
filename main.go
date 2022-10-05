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
	i, err := os.Open("./data.bin")
	if err != nil {
		panic(err)
	}
	defer i.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	client, err := newPublisher()
	if err != nil {
		panic(err)
	}

	go func() {
		timestamp, typeId, value, err := read(i)
		if err != nil {
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
				fmt.Printf("Matching value: timestamp=%v, type=%d, value=%0.2f\n", timestamp, typeId, value)

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
					panic(err)
				}

				// read next record
				timestamp, typeId, value, err = read(i)
				if err != nil {
					if err == io.EOF {
						break
					}
					panic(err)
				}
			} else if comp.Before(now.Add(-time.Second)) {
				// skip record
				timestamp, typeId, value, err = read(i)
				if err != nil {
					if err == io.EOF {
						return
					}
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
	conn, err := net.Dial("tcp", "localhost")
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
