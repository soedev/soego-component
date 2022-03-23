package erabbitmq

import "github.com/streadway/amqp"

type MessageHandle func(<-chan amqp.Delivery)
type AckHandle func(<-chan amqp.Delivery, bool)
