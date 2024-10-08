// Copyright 2022 lukemakhanu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rabbit

import (
	"log"
	"time"

	"github.com/streadway/amqp"
)

type QueueConsume struct {
	url            string
	name           string
	connectionName string
	consumerName   string

	errorChannel chan *amqp.Error
	connection   *amqp.Connection
	channel      *amqp.Channel
	closed       bool

	consumers []msgConsumer
}

type msgConsumer func(string)

func NewQueueConsume(url string, qName string, connectionName string, consumerName string) *QueueConsume {
	q := new(QueueConsume)
	q.url = url
	q.name = qName
	q.connectionName = connectionName
	q.consumerName = consumerName
	q.consumers = make([]msgConsumer, 0)

	q.connect()
	go q.reConnector()

	return q
}

func (q *QueueConsume) connect() {
	for {
		log.Printf("Connecting to rabbitmq on %s\n", q.url)
		cfg := amqp.Config{
			Properties: amqp.Table{
				"connection_name": q.connectionName,
			},
		}
		conn, err := amqp.DialConfig(q.url, cfg)
		//conn, err := amqp.Dial(q.url)
		if err == nil {
			q.connection = conn
			q.errorChannel = make(chan *amqp.Error)
			q.connection.NotifyClose(q.errorChannel)

			log.Println("Connection established!")

			q.openChannel()
			q.declareQueue()

			return
		}

		if err != nil {
			log.Printf("Connection to rabbit mq failed. Retrying in 2 sec... : %v ", err)
		}

		time.Sleep(2000 * time.Millisecond)
	}
}

func (q *QueueConsume) reConnector() {
	for {
		err := <-q.errorChannel
		if !q.closed {
			log.Printf("Reconnecting after connection closed : %v", err)

			q.connect()
			q.recoverConsumers()
		}
	}
}

func (q *QueueConsume) openChannel() {
	channel, err := q.connection.Channel()
	if err != nil {
		log.Printf("Opening channel failed : %v", err)
	}
	q.channel = channel
}

func (q *QueueConsume) declareQueue() {
	_, err := q.channel.QueueDeclare(
		q.name, // name
		true,   // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)

	if err != nil {
		log.Printf("QueueConsume declaration failed : %v", err)
	}
}

func (q *QueueConsume) Consume(consumer msgConsumer) {
	log.Println("Registering consumer...")
	deliveries, err := q.registerQueueConsumer()
	log.Println("Consumer registered! Processing messages...")
	q.executeMessageConsumer(err, consumer, deliveries, false)
}

func (q *QueueConsume) recoverConsumers() {
	for i := range q.consumers {
		var consumer = q.consumers[i]

		log.Println("Recovering consumer...")
		msgs, err := q.registerQueueConsumer()
		log.Println("Consumer recovered! Continuing message processing...")
		q.executeMessageConsumer(err, consumer, msgs, true)
	}
}

func (q *QueueConsume) executeMessageConsumer(err error, consumer msgConsumer, deliveries <-chan amqp.Delivery, isRecovery bool) {
	if err == nil {
		if !isRecovery {
			q.consumers = append(q.consumers, consumer)
		}
		go func() {
			for delivery := range deliveries {
				consumer(string(delivery.Body[:]))
			}
		}()
	}
}

func (q *QueueConsume) registerQueueConsumer() (<-chan amqp.Delivery, error) {
	msgs, err := q.channel.Consume(
		q.name,         // queue
		q.consumerName, // msgConsumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Printf("Consuming messages from queue failed : %v", err)
	}
	return msgs, err
}

func (q *QueueConsume) Close() {
	log.Println("Closing connection")
	q.closed = true
	q.channel.Close()
	q.connection.Close()
}
