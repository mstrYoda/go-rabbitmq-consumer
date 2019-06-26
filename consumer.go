package main

import (
	"errors"
	"log"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func declareExchange(ch *amqp.Channel, name string, kind string) {
	err := ch.ExchangeDeclare(name, kind, true, false, false, false, nil)
	failOnError(err, "can not declare exchange")
}

func declareQueue(ch *amqp.Channel, name string, deadLetterExchange string) {
	args := make(amqp.Table)

	if deadLetterExchange != "" {
		args["x-dead-letter-exchange"] = deadLetterExchange
	}

	_, err := ch.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
	failOnError(err, "can not declare q")
}

func declareDlq(ch *amqp.Channel, name string) {

	_, err := ch.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "can not declare q")
}

func queueBinding(ch *amqp.Channel, queueName string, exchangeName string) {
	err := ch.QueueBind(
		queueName,    // queue name
		"",           // routing key
		exchangeName, // exchange
		false,
		nil)
	failOnError(err, "can not bind q")
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	ch.Qos(1, 0, false)
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	declareExchange(ch, "test.exchange", amqp.ExchangeDirect)
	declareQueue(ch, "test_queue", "test.exchange.dead-letter")
	queueBinding(ch, "test_queue", "test.exchange")

	declareExchange(ch, "test.exchange.dead-letter", amqp.ExchangeFanout)
	declareDlq(ch, "test_queue.dead-letter")
	queueBinding(ch, "test_queue.dead-letter", "test.exchange.dead-letter")

	msgs, err := ch.Consume(
		"test_queue", // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			err = doWork()
			if err != nil {
				d.Reject(false)
				log.Printf("message sent to dead letter")
			} else {
				d.Ack(false)
				log.Printf("message consumed successfully")
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func doWork() error {
	return errors.New("an error occured")
}
