/*
* A wrapped around the paho MQTT library to provide some robustness
* and limit the MQTT-based functionality.
*
 */

package vizier

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"strconv"
	"sync"
)

type Callback func([]byte)
type fullCallback func(mqtt.Client, mqtt.Message)
type Signal int

const (
	STOP = iota
	RESTART
	DONE
)

type mqttClient struct {
	subscriptions map[string]fullCallback
	client        mqtt.Client
	mutex         sync.Mutex
	started       bool
	stopped       bool
	signals       chan Signal
	waitUntilDone chan Signal
}

type MQTTClient interface {
	Subscribe(string) chan []byte
	SubscribeWithCallback(string, Callback)
	Start() error
	Stop()
}

func NewMQTTClient(host string, port int) MQTTClient {
	// Initialize struct
	mi := &mqttClient{}
	mi.started = false
	mi.stopped = false
	mi.subscriptions = make(map[string]fullCallback)
	mi.mutex = sync.Mutex{}
	mi.signals = make(chan Signal)

	// Start goroutine for auto resubscribe
	go func() {

		fmt.Println("Entering go!")
		for {
			fmt.Println("in loop")
			signal := <-mi.signals

			fmt.Println("Got signal")

			switch signal {
			case STOP:
				fmt.Println("Stop received!")
				fmt.Println("Exiting!")
				return
				break

			case RESTART:
				fmt.Println("Restart signal received.  Resubscribing.")
				mi.mutex.Lock()
				defer mi.mutex.Unlock()
				for k, v := range mi.subscriptions {
					mh := mqtt.MessageHandler(v)
					mi.client.Subscribe(k, 0, mh)
				}
				break
			}
		}
	}()

	brokerString := "tcp://" + host + ":" + strconv.Itoa(port)
	fmt.Println(brokerString)

	ops := mqtt.NewClientOptions().SetClientID("golang").AddBroker(brokerString)
	mi.client = mqtt.NewClient(ops)

	return mi
}

func (mi *mqttClient) SubscribeWithCallback(topic string, callback Callback) {
	messageHandler := func(client mqtt.Client, message mqtt.Message) {
		callback(message.Payload())
	}

	mi.mutex.Lock()
	defer mi.mutex.Unlock()

	mi.client.Subscribe(topic, 0, messageHandler)
	mi.subscriptions[topic] = messageHandler
}

func (mi *mqttClient) Subscribe(topic string) chan []byte {
	c := make(chan []byte, 100)
	callback := func(payload []byte) {
		c <- payload
	}

	mi.SubscribeWithCallback(topic, callback)

	return c
}

func (mi *mqttClient) Start() error {
	token := mi.client.Connect()

	token.Wait()
	err := token.Error()

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (mi *mqttClient) Stop() {
	mi.client.Disconnect(1000)
	// Send stop signal to internal goroutine that handles restarts
	mi.signals <- STOP
}
