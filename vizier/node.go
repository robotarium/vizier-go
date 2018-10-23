package vizier

type node struct {
	mqttClient MQTTClient
	requests   map[string]Request
	links      map[string]Link
	data       map[string]string
}

type Node interface {
	Start() error
	Stop()
	Verify() bool
	Subscribe(topic string)
	SubscribeWithCallback(topic string, callback Callback)
	Get(topic string)
}

func NewNode(host string, port int) Node {
	n = &node{}
	n.mqttClient = NewMQTTClient(host, port)
	n.requests = make(map[string]Request)
	n.links = make(map[string]Link)
	n.data = make(map[string]string)

	return n
}

func (this *Node) makeRequest(method string, attempts int, timeout int) ([]byte, error) {

}

// Implementing Node

func (this *node) Start() (err error) {
	err = this.mqttClient.Start()
}

func (this *node) Stop() {
	this.mqttClient.Stop()
}

func (this *Node) Verify() bool {

}
