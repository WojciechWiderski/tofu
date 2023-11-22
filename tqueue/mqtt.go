package tqueue

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/WojciechWiderski/tofu/tconfig"
)

type MQTT struct {
	config      tconfig.MQTT
	Client      mqtt.Client
	Publishers  []PubFn
	Subscribers []SubFn
}

type SubFn struct {
	Topic string
	Fn    mqtt.MessageHandler
}

type PubFn struct {
	Topic string
	Fn    func() (interface{}, error)
}

func (m *MQTT) AddSubscribe(topic string, fn func(in interface{})) {
	var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		fn(msg.Payload())
	}
	m.Subscribers = append(m.Subscribers, SubFn{
		Topic: topic,
		Fn:    messagePubHandler,
	})
}

func (m *MQTT) AddPublisher(topic string, fn func() (interface{}, error)) {
	m.Publishers = append(m.Publishers, PubFn{
		Topic: topic,
		Fn:    fn,
	})
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func NewMqtt(config tconfig.MQTT) *MQTT {
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

func (m *MQTT) Publish(topic string, fn func() (interface{}, error)) {
	out, err := fn()
	if err == nil {
		token := m.Client.Publish(topic, 0, false, out)
		token.Wait()
	}
}

func (m *MQTT) Subscribe(topic string, fn mqtt.MessageHandler) {
	token := m.Client.Subscribe(topic, 1, fn)
	token.Wait()
	fmt.Println("Subscribed to topic: ", topic)
}

func (m *MQTT) Disconnect() {
	fmt.Println("Disconnecting...")
	m.Client.Disconnect(250)
}
