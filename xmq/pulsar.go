package xmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

// pulsar对象
type Pulsar struct {
	Client   pulsar.Client   // pulsar client
	Topic    string          // topic
	Sub      pulsar.Consumer // consumer
	Producer pulsar.Producer // producer
	Conf     PulsarConfig    // config
}

// pulsar配置
type PulsarConfig struct {
	Url          string                  `json:"url" yaml:"url"`                     // pulsar url
	Topic        string                  `json:"topic" yaml:"topic"`                 // topic
	Namespace    string                  `json:"namespace" yaml:"namespace"`         // namespace
	Tenant       string                  `json:"tenant" yaml:"tenant"`               // tenant
	SubName      string                  `json:"sub_name" yaml:"sub_name"`           // subscription name
	SubType      pulsar.SubscriptionType `json:"sub_type" yaml:"sub_type"`           // subscription type
	ProducerName string                  `json:"producer_name" yaml:"producer_name"` // producer name
}

func NewPulsar(config PulsarConfig) *Pulsar {
	return &Pulsar{Topic: config.Topic, Conf: config}
}

func (p *Pulsar) NewClient() error {
	if p.Conf.Url == "" {
		return errors.New("url is empty")
	}
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: p.Conf.Url,
	})
	if err != nil {
		return err
	}
	p.Client = client
	return nil
}

func (p *Pulsar) NewConsumer() error {
	if p.Client == nil {
		err := p.NewClient()
		if err != nil {
			return err
		}
	}
	topic := p.Topic
	if p.Conf.Namespace != "" {
		topic = p.Conf.Namespace + "/" + p.Topic
	}
	if p.Conf.Tenant != "" {
		topic = p.Conf.Tenant + "/" + topic
	}
	consumer, err := p.Client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: p.Conf.SubName,
		Type:             p.Conf.SubType,
	})
	if err != nil {
		return err
	}
	p.Sub = consumer
	return nil
}

func (p *Pulsar) GetProducer() pulsar.Producer {
	if p.Producer == nil {
		err := p.NewProducer()
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}
	return p.Producer
}

func (p *Pulsar) NewProducer() error {
	if p.Client == nil {
		err := p.NewClient()
		if err != nil {
			return err
		}
	}
	topic := p.Topic
	if p.Conf.Namespace != "" {
		topic = p.Conf.Namespace + "/" + p.Topic
	}
	if p.Conf.Tenant != "" {
		topic = p.Conf.Tenant + "/" + topic
	}
	producer, err := p.Client.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
		Name:  p.Conf.ProducerName,
	})
	if err != nil {
		return err
	}
	p.Producer = producer
	return nil
}

func (p *Pulsar) Close() {
	if p.Sub != nil {
		p.Sub.Close()
		p.Sub = nil
		fmt.Println("pulsar sub closed")
	}
	if p.Producer != nil {
		p.Producer.Close()
		p.Producer = nil
		fmt.Println("pulsar producer closed")
	}
	if p.Client != nil {
		p.Client.Close()
		p.Client = nil
		fmt.Println("pulsar client closed")
	}
}

// 发送消息
func (p *Pulsar) Send(msg pulsar.ProducerMessage) error {
	if p.Producer == nil {
		return errors.New("producer is nil")
	}
	_, err := p.Producer.Send(context.Background(), &msg)
	return err
}

// 接收消息
func (p *Pulsar) Recv() (pulsar.Message, error) {
	if p.Sub == nil {
		return nil, errors.New("consumer is nil")
	}
	msg, err := p.Sub.Receive(context.Background())
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// 确认消息
func (p *Pulsar) Ack(msg pulsar.Message) error {
	if p.Sub == nil {
		return errors.New("consumer is nil")
	}
	return p.Sub.Ack(msg)
}

// 确认消息
func (p *Pulsar) AckID(msgID pulsar.MessageID) error {
	if p.Sub == nil {
		return errors.New("consumer is nil")
	}
	return p.Sub.AckID(msgID)
}

// 确认消息
func (p *Pulsar) AckCumulative(msg pulsar.Message) error {
	if p.Sub == nil {
		return errors.New("consumer is nil")
	}
	return p.Sub.AckCumulative(msg)
}

// 确认消息
func (p *Pulsar) AckIDCumulative(msgID pulsar.MessageID) error {
	if p.Sub == nil {
		return errors.New("consumer is nil")
	}
	return p.Sub.AckIDCumulative(msgID)
}

// 重新消费消息
func (p *Pulsar) ReconsumeLater(msg pulsar.Message, delay time.Duration) {
	if p.Sub == nil {
		return
	}
	p.Sub.ReconsumeLater(msg, delay)
}

// 发送json数据
func (p *Pulsar) SendJson(data interface{}) (pulsar.MessageID, error) {
	if p.Producer == nil {
		return nil, errors.New("producer is nil")
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return p.Producer.Send(context.Background(), &pulsar.ProducerMessage{Payload: jsonData})
}

// 异步发送json数据
func (p *Pulsar) AsyncSendJson(data interface{}, callback func(pulsar.MessageID, *pulsar.ProducerMessage, error)) {
	if p.Producer == nil {
		callback(nil, nil, errors.New("producer is nil"))
		return
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		callback(nil, nil, err)
		return
	}
	p.Producer.SendAsync(context.Background(), &pulsar.ProducerMessage{Payload: jsonData}, callback)
}
