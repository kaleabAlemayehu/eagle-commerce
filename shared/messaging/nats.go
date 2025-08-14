package messaging

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type NATSClient struct {
	conn *nats.Conn
}

func NewNATSClient(url string) (*NATSClient, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &NATSClient{conn: conn}, nil
}

func (n *NATSClient) Publish(subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return n.conn.Publish(subject, payload)
}

func (n *NATSClient) Subscribe(subject string, handler func([]byte)) (*nats.Subscription, error) {
	return n.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	})
}

func (n *NATSClient) SubscribeToRequest(subject string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	return n.conn.Subscribe(subject, handler)
}

func (n *NATSClient) Request(subject string, data any, timeout time.Duration) (*nats.Msg, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return n.conn.Request(subject, payload, timeout)
}

func (n *NATSClient) Close() {
	n.conn.Close()
}

func GenerateEventID() string {
	return uuid.NewString()
}
