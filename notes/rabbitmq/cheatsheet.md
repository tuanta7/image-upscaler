# RabbitMQ Cheatsheet

## 1. Flags

In AMQP, "flags" refer to various boolean or enumerated properties within protocol methods and message headers that control specific behaviors related to message routing, persistence, and system management. These flags determine how the broker and clients interact and handle data.

### 1.1. Declaration

| Flag            | Type | Methods                         |
| --------------- | ---- | ------------------------------- |
| **auto_delete** | bool | queue.declare, exchange.declare |

Marks the entity for automatic deletion when it is no longer in use, typically when the last consumer unsubscribes or when all bindings are removed.

| Flag        | Type | Methods                         |
| ----------- | ---- | ------------------------------- |
| **durable** | bool | queue.declare, exchange.declare |

Specifies that the declared entity should survive broker restarts. A durable queue or exchange is persisted in the broker's metadata store.

| Flag          | Type | Methods       |
| ------------- | ---- | ------------- |
| **exclusive** | bool | queue.declare |

Declares the queue as exclusive to the current connection. Exclusive queues are removed automatically when the connection closes.

| Flag         | Type | Methods          |
| ------------ | ---- | ---------------- |
| **internal** | bool | exchange.declare |

Restricts the exchange for internal routing only. Messages cannot be published directly to an internal exchange by clients. Instead, they are typically routed to the internal exchange from another exchange, often as part of a more complex routing topology.

| Flag       | Type | Methods                                                                                   |
| ---------- | ---- | ----------------------------------------------------------------------------------------- |
| **nowait** | bool | queue.declare, exchange.declare, queue.bind, queue.unbind, exchange.bind, exchange.unbind |

Indicates that no method response is expected. When set to true, the broker omits `*-ok` replies, allowing operations to proceed without synchronous confirmation.

| Flag        | Type | Methods                         |
| ----------- | ---- | ------------------------------- |
| **passive** | bool | queue.declare, exchange.declare |

Declares the entity passively, indicating that it must already exist. The broker performs a validation check and raises an exception if the entity is missing, without altering or creating it.

### 1.2. Publish

Indicates transient vs persistent storage preference.

| Flag          | Type | Methods       |
| ------------- | ---- | ------------- |
| **mandatory** | bool | basic.publish |

Requires the broker to return unroutable messages instead of silently dropping them. Returned messages are emitted through the `basic.return` channel.

| Flag              | Type | Methods                                                                                              |
| ----------------- | ---- | ---------------------------------------------------------------------------------------------------- |
| **delivery_mode** | bool | basic.publish (in [message properties](https://www.rabbitmq.com/docs/publishers#message-properties)) |

Specifies that the message should be persisted to disk, allowing it to survive broker restarts, provided the target queue is durable.

| Flag            | Type | Methods                     |
| --------------- | ---- | --------------------------- |
| **redelivered** | bool | basic.deliver, basic.get-ok |

Indicates whether the message has previously been delivered and returned to the queue.

### 1.3. Consume

| Flag         | Type | Methods       |
| ------------ | ---- | ------------- |
| **auto_ack** | bool | basic.consume |

Indicates that messages are delivered without requiring explicit acknowledgments.

| Flag             | Type   | Methods       |
| ---------------- | ------ | ------------- |
| **consumer_tag** | string | basic.consume |

Provides a consumer identifier that can be used to manage or cancel the consumer.

| Flag         | Type | Methods       |
| ------------ | ---- | ------------- |
| **no_local** | bool | basic.consume |

Prevent a consumer from receiving messages that were published by the same connection that the consumer is using. This flag is rarely used and not fully supported in RabbitMQ.

| Flag               | Type | Methods   |
| ------------------ | ---- | --------- |
| **prefetch_count** | bool | basic.qos |

Defines the maximum number of unacknowledged messages that can be delivered to the consumer. This is used to regulate load distribution and flow control.

> [!NOTE]
> When `auto_ack` is active, no unacknowledged messages exist, causing QoS settings such as `prefetch_count` to be bypassed.

## 2. Arguments

Arguments can be any key-value pair and are used for feature extensions. Some properties are mandatory while others are optional.
