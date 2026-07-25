# RabbitMQ Pub/Sub

To deliver a message to multiple consumers, RabbitMQ supports some useful features:

- **Fanout Exchanges**: Route a copy of every message published to them to every queue, stream or exchange bound to it
- **Topic Exchange**: Route messages based on a pattern
- **Exclusivity** (static configuration): For Classic Queues only, when registering a consumer with an AMQP 0-9-1 client, the exclusive flag of the basic.consume method can be set to true to request the consumer to be the only one on the target queue.
- **Single Active Consumer** (dynamic feature): Single active consumer allows to have only one consumer at a time consuming from a queue and to **fail over** to another registered consumer in case the active one is cancelled or dies.

## 1. Topic Exchange

 The logic behind the `topic` exchange is similar to a `direct` one - a message sent with a particular routing key will be delivered to all the queues that are bound with a matching binding key. However there are two important special cases for binding keys:

- `*` can substitute for exactly one word.
- `#` can substitute for zero or more words.

## 2. Temporary Queue 

There are three ways to make queue deleted automatically

- **Exclusive Queues** (covered below)
- **Auto-delete Queues**: Queue that has had at least one consumer is deleted when last consumer unsubscribes
- [TTL](https://www.rabbitmq.com/docs/ttl): Both queues and messages can have a TTL

An exclusive queue can only be used (consumed from, purged, deleted, etc) by its declaring connection. Since such queues cannot be shared between N consumers, consider using server-generated names for exclusive queues

- Queue with empty string name will be given a generated one. 
- Exclusive queues are deleted when their declaring connection is closed or gone (e.g. due to underlying TCP connection loss).

```go
// This queue will be deleted when the connection that declared it closes because it is declared as exclusive.
q, err := ch.QueueDeclare(
  "",    // name
  false, // durable
  false, // auto-delete
  true,  // exclusive
  false, // no-wait
  nil,   // arguments
)
```
