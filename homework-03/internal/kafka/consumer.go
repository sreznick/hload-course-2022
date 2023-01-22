package kafka

import (
	"github.com/Shopify/sarama"
)

type Consumer interface {
	sarama.ConsumerGroupHandler
	GetMessagesChan() chan []byte
}

type consumer struct {
	channel chan []byte
}

func NewConsumer() Consumer {
	return &consumer{
		channel: make(chan []byte),
	}
}

func (c *consumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-session.Context().Done():
			close(c.channel)
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				close(c.channel)
				return nil
			}

			c.channel <- msg.Value
			session.MarkMessage(msg, "")
			session.Commit()
		}
	}
}

func (c *consumer) GetMessagesChan() chan []byte {
	return c.channel
}
