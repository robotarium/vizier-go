/*
* A wrapped around the paho MQTT library to provide some robustness
* and limit the MQTT-based functionality.
*
 */

package main

import (
	json "encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"strconv"
	"sync"
)

type Callback func([]byte)

type mqttInterface struct {
	callbacks map[string]Callback
	client    mqtt.Client
	mutex     sync.Mutex
	started   bool
	stopped   bool
}

type MqttInterface interface {
	Subscribe(string) chan []byte
	SubscribeWithCallback(string, Callback)
	Start() error
	Stop()
}

func NewMqttInterface(host string, port int) MqttInterface {
	mi := &mqttInterface{}
	mi.callbacks = make(map[string]Callback)
	mi.mutex = sync.Mutex{}
	mi.started = false
	mi.stopped = false

	brokerString := "tcp://" + host + ":" + strconv.Itoa(port)
	fmt.Println(brokerString)

	ops := mqtt.NewClientOptions().SetClientID("golang").AddBroker(brokerString)
	mi.client = mqtt.NewClient(ops)

	return mi
}

func (mi *mqttInterface) SubscribeWithCallback(topic string, callback Callback) {
	messageHandler := func(client mqtt.Client, message mqtt.Message) {
		callback(message.Payload())
	}

	mi.mutex.Lock()
	defer mi.mutex.Unlock()

	mi.client.Subscribe(topic, 0, messageHandler)
	mi.callbacks[topic] = callback
}

func (mi *mqttInterface) Subscribe(topic string) chan []byte {
	c := make(chan []byte, 100)
	callback := func(payload []byte) {
		c <- payload
	}

	mi.SubscribeWithCallback(topic, callback)

	return c
}

func (mi *mqttInterface) Start() error {
	token := mi.client.Connect()

	token.Wait()
	err := token.Error()

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (mi *mqttInterface) Stop() {
	mi.client.Disconnect(1000)
}

func main() {
	mi := NewMqttInterface("192.168.1.8", 1884)
	err := mi.Start()

	if err != nil {
		fmt.Println(err)
		return
	}

	channel := mi.Subscribe("matlab_api/1")

	callback := func(payload []byte) {
		fmt.Println("Callback:", string(payload))

		var js interface{}
		err := json.Unmarshal(payload, &js)

		if err != nil {
			fmt.Println("error!")
			fmt.Println(err)
			return
		}

		fmt.Println(js)
	}
	mi.SubscribeWithCallback("matlab_api/2", callback)

	for i := 0; i < 1000; i++ {
		receive := <-channel
		fmt.Println(string(receive))
	}

	mi.Stop()
}
