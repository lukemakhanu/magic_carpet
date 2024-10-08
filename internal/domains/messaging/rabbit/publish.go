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

type QueuePublish struct {
	url            string
	connectionName string
	name           string
	exchangeName   string
	routingKey     string

	errorChannel chan *amqp.Error
	connection   *amqp.Connection
	channel      *amqp.Channel
	closed       bool
}

type msgsConsumer func(string)

func NewQueuePublish(url, connectionName, queueName, exchangeName, routingKey string) *QueuePublish {
	q := new(QueuePublish)
	q.url = url
	q.connectionName = connectionName
	q.name = queueName
	q.exchangeName = exchangeName
	q.routingKey = routingKey

	q.connect()
	go q.reConnector()

	return q
}

func (q *QueuePublish) connect() {
	for {
		cfg := amqp.Config{
			Properties: amqp.Table{
				"connection_name": q.connectionName,
			},
		}
		conn, err := amqp.DialConfig(q.url, cfg)
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
			log.Printf("Connection to rabbitmq failed. Retrying in 2 sec... : %v", err)
		}
		time.Sleep(2000 * time.Millisecond)
	}

}

func (q *QueuePublish) declareQueue() {
	_, err := q.channel.QueueDeclare(
		q.name, // name
		true,   // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		log.Printf("Queue declaration failed : %v", err)
	}
}

func (q *QueuePublish) reConnector() {
	for {
		err := <-q.errorChannel
		if !q.closed {
			log.Printf("Reconnecting after connection closed : %v", err)

			q.connect()
		}
	}
}

func (q *QueuePublish) openChannel() {
	channel, err := q.connection.Channel()
	if err != nil {
		log.Printf("Opening channel failed : %v", err)
	}
	q.channel = channel
}

func (q *QueuePublish) Close() {
	log.Println("Closing connection ...")
	q.closed = true
	q.channel.Close()
	q.connection.Close()
}

func (q *QueuePublish) Send(message string) {
	err := q.channel.Publish(
		q.exchangeName, // exchange
		q.routingKey,   // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		log.Printf("Sending message to queue failed : %v", err)
	}
}
