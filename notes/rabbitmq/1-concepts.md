# RabbitMQ

RabbitMQ is a message broker, it accepts and forwards messages.

## 1. Message Lifecycle

Reference: [Acknowledger Interface](https://github.com/rabbitmq/amqp091-go/blob/v1.10.0/delivery.go#L19)

When a message enters RabbitMQ it passes through three distinct phases

| Phase     | Internal state | Description                                                                                                 |
| --------- | -------------- | ----------------------------------------------------------------------------------------------------------- |
| Queued    | Ready          | The message resides on disk or in memory, assigned to a queue but not yet delivered.                        |
| In‑flight | Unacknowledged | The message has been sent to a consumer, yet the broker has not received an explicit acknowledgment for it. |
| Completed | Acked / Nacked | The consumer has responded, the broker then removes or re‑queues the message accordingly.                   |

Consumer acknowledgments (delivery‑side reliability)

| Method                     | Effect                                                                    | Typical use        |
| -------------------------- | ------------------------------------------------------------------------- | ------------------ |
| `basic.ack`                | Confirms successful processing; message is deleted.                       | Normal completion. |
| `basic.nack requeue=true`  | Signals failure; message is re‑queued for another attempt.                | Transient errors.  |
| `basic.nack requeue=false` | Routes the message to a dead‑letter exchange (if configured) or drops it. | Poison messages.   |
| `basic.reject`             | Single‑message variant of NACK; identical semantics.                      | Edge cases.        |

With autoAck=true the broker discards the message immediately after sending it, eliminating the unacknowledged phase but forfeiting at‑least‑once guarantees.

> [!CAUTION]
> Publisher confirms are an independent protocol feature

| Direction         | Mechanism               | Guarantees                                                              |
| ----------------- | ----------------------- | ----------------------------------------------------------------------- |
| Producer → Broker | Publisher Confirm       | Message has been persisted and routed to at least one queue.            |
| Broker → Consumer | Delivery Acknowledgment | Message has been processed (or intentionally rejected) by the consumer. |

## 2. Channels

Most applications need multiple logical connections to the broker. However, it is undesirable to keep many TCP connections open at the same time because doing so consumes system resources and makes it more difficult to configure firewalls.

- AMQP 0-9-1 connections are multiplexed with channels.
- Channels can be thought of as lightweight connections that share a single TCP connection.
- Much like connections, channels are meant to be long lived (using channel pool).
- The client cannot be configured to allow for more channels than the server configured maximum

## 3. Queues

A queue in RabbitMQ is an ordered collection of messages. Messages are enqueued and dequeued (delivered to consumers) in a FIFO manner.

Queues have properties that define how they behave. There is a set of mandatory properties and a map of optional ones

- **Name**: The name of the queue. In AMQP 0-9-1, the broker can generate a unique queue name on behalf of an app
- **Durability**: Queues can be durable or transient. Metadata of a durable queue is stored on disk, while metadata of a transient queue is stored in memory when possible.
- **Exclusive**: Queue that used by only one connection and will be deleted when that connection closes
- **Auto-delete**: Queue that has had at least one consumer is deleted when last consumer unsubscribes
- **Arguments** (optional): Used by plugins and broker-specific features such as message TTL, queue length limit, etc.

> [!NOTE]
> When an auto-delete queue is created, and there are no consumers declared on it, the queue will not be automatically deleted because the condition of "last consumer unsubscribes" for deletion is never met.

### 3.1. Quorum Queue

### 3.2. Streams

RabbitMQ Streams is a persistent replicated data structure that can complete the same tasks as queues: they buffer messages from producers that are read by consumers.

## 4. Exchanges

In AMQP 0-9-1, exchanges are the entities where publishers publish messages that are then routed to a set of queues or streams. They have several key properties that can be specified at declaration:

- **Durability**: Just like queues, exchanges can be durable or transient. However, transient exchanges are very rarely used in practice.
- **Auto-delete**: Auto-deleted exchanges are deleted when their last binding is removed.
- **Arguments** (optional): Optional exchange arguments, also known as "x-arguments" because of their field name in the AMQP 0-9-1 protocol, is a map (dictionary) of arbitrary key/value pairs that can be provided by clients when an exchange is declared.

The routing between producers and consumer queues is via **bindings**. These bindings form the logical topology of the broker.

- Exchanges routes all messages that flow through them to one or more queues, streams, or other exchanges.
- Every exchange belongs to one virtual host (logical groups of entities)

### 4.1. Exchange Types

- **direct**: exact `routing-key` match
- **fanout**: broadcast to all bound queues
- **topic**: wildcard matching with `*` (single word) and `#`(multi-word)
- **headers**: match based on message headers

### 4.2. Dead Letter Exchange

Messages from a queue can be "dead-lettered", which means these messages are republished to an exchange
