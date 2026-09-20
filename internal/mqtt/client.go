package mqtt

import (
	"fmt"
	"greenhouse-iot-golang/internal/config"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	Client mqtt.Client
}

func NewMQTTClient(cfg *config.Config) (*Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTTBrokerURL)
	opts.SetClientID(cfg.MQTTClientID)
	opts.SetKeepAlive(time.Duration(cfg.MQTTKeepAliveSeconds) * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		slog.Info("MQTT client connected to broker", "broker", cfg.MQTTBrokerURL)
	})

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		slog.Warn("MQTT connection lost; automatic reconnect initiated", "error", err)
	})

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.WaitTimeout(5*time.Second) && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	return &Client{Client: client}, nil
}

func (c *Client) IsConnected() bool {
	if c == nil || c.Client == nil {
		return false
	}
	return c.Client.IsConnected()
}
