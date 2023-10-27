package tofu

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTT struct {
	config      MQTTConfig
	Client      mqtt.Client
	publishers  []PubFn
	subscribers []SubFn
}

type SubFn struct {
	Topic string
	fn    mqtt.MessageHandler
}

type PubFn struct {
	Topic string
	fn    func() (interface{}, error)
}

func (m *MQTT) AddSubscribe(topic string, fn func(in interface{})) {
	var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		fn(msg.Payload())
	}
	m.subscribers = append(m.subscribers, SubFn{
		Topic: topic,
		fn:    messagePubHandler,
	})
}

func (m *MQTT) AddPublisher(topic string, fn func() (interface{}, error)) {
	m.publishers = append(m.publishers, PubFn{
		Topic: topic,
		fn:    fn,
	})
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func NewMQTT(config MQTTConfig) *MQTT {
	m := &MQTT{
		config: config,
	}
	client, err := m.connectToBroker()
	if err != nil {
		panic(err)
	}
	m.Client = client
	return m
}

func (m *MQTT) connectToBroker() (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", m.config.Broker, m.config.Port))
	opts.SetClientID(m.config.ClientID)
	opts.SetUsername(m.config.Username)
	opts.SetPassword(m.config.Password)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

func (m *MQTT) publish(topic string, fn func() (interface{}, error)) {
	out, err := fn()
	if err == nil {
		token := m.Client.Publish(topic, 0, false, out)
		token.Wait()
	}
}

func (m *MQTT) subscribe(topic string, fn mqtt.MessageHandler) {
	token := m.Client.Subscribe(topic, 1, fn)
	token.Wait()
	fmt.Println("Subscribed to topic: ", topic)
}

func (m *MQTT) disconnect() {
	fmt.Println("Disconnecting...")
	m.Client.Disconnect(250)
}
