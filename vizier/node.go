package vizier

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type node struct {
	mqttClient MQTTClient
	requests   map[string]Request
	links      map[string]Link
	data       map[string]string

	subscribable map[string]bool
	publishable  map[string]bool
	gettable     map[string]bool
	puttable     map[string]bool
}

type Node interface {
	Start() error
	Stop()
	Verify() bool
	Subscribe(string) chan []byte
	SubscribeWithCallback(string, Callback)
	Get(string) (string, bool)
}

func NewNode(host string, port int) Node {
	this := &node{}
	this.mqttClient = NewMQTTClient(host, port)
	this.requests = make(map[string]Request)
	this.links = make(map[string]Link)
	this.data = make(map[string]string)
	this.subscribable = make(map[string]bool)
	this.publishable = make(map[string]bool)
	this.gettable = make(map[string]bool)
	this.puttable = make(map[string]bool)

	return this
}

func insertAfter(channel chan<- bool, timeout int) {
	time.Sleep(time.Duration(timeout) * time.Second)
	channel <- true
}

func (this *node) makeRequest(link string, method string, attempts int, timeout int) (response VizierResponse, ok bool) {

	response = VizierResponse{}
	ok = false

	tokens := strings.Split(link, "/")
	if len(tokens) == 0 {
		return response, false
	}
	remoteEndpoint := tokens[0]

	id := MessageID(20)
	req := VizierRequest{id, link, "GET", ""}
	reqMap := make(map[string]interface{})
	reqMap["id"] = id
	reqMap["link"] = link
	reqMap["method"] = "GET"
	reqMap["body"] = ""
	fmt.Println(req)
	mReq, err := json.Marshal(req)
	fmt.Println(string(mReq))

	if err != nil {
		fmt.Println(err)
		return response, false
	}

	responseLink := remoteEndpoint + "/responses/" + id
	requestLink := remoteEndpoint + "/requests"
	responseChan := this.mqttClient.Subscribe(responseLink)

	for i := 0; i < attempts; i++ {

		this.mqttClient.Publish(requestLink, mReq)
		fmt.Println("Attempting to get response")

		timeoutChan := make(chan bool)
		go insertAfter(timeoutChan, timeout)

		select {
		case <-timeoutChan:
			break
		case responseBytes := <-responseChan:
			response_ := VizierResponse{}
			if err := json.Unmarshal(responseBytes, &response_); err == nil {
				response = response_
			} else {
				fmt.Println("uh oh!")
			}
			ok = true
			// Set i to attempts to break out of the loop
			i = attempts
		}
	}

	fmt.Println(response)
	this.mqttClient.Unsubscribe(responseLink)

	return response, ok
}

func (this *node) Get(link string) (string, bool) {
	if body, ok := this.makeRequest(link, "GET", 1, 1); ok {
		return body.Body, ok
	}

	return "", false
}

// Implementing Node
func (this *node) Subscribe(topic string) chan []byte {
	return this.mqttClient.Subscribe(topic)
}

func (this *node) SubscribeWithCallback(topic string, callback Callback) {
	this.mqttClient.SubscribeWithCallback(topic, callback)
}

func (this *node) Unsubscribe(topic string) {
	this.mqttClient.Unsubscribe(topic)
}

func (this *node) Verify() bool {
	return true
}

func (this *node) Start() (err error) {
	err = this.mqttClient.Start()
	return err
}

func (this *node) Stop() {
	this.mqttClient.Stop()
}
