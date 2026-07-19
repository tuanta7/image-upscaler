import json
import logging
import os
import time

import pika
import pika.exceptions

log = logging.getLogger(__name__)

RABBITMQ_URL = os.environ.get("RABBITMQ_URL", "amqp://rabbitmq:password@localhost:5672/")
TASK_QUEUE = os.environ.get("TASK_QUEUE", "upscale.tasks")

class Consumer:
    def __init__(self, handler, url=RABBITMQ_URL, queue=TASK_QUEUE):
        self.handler = handler
        self.url = url
        self.queue = queue

    def run(self):
        """Consume forever, reconnecting when the broker drops us."""
        while True:
            try:
                self._consume()
            except pika.exceptions.AMQPConnectionError:
                log.warning("lost connection to RabbitMQ, retrying in 5s")
                time.sleep(5)
            except KeyboardInterrupt:
                log.info("shutting down")
                return

    def _consume(self):
        # heartbeat=600 gives long-running upscale jobs room to finish
        # without the broker deciding we are dead mid-task.
        params = pika.URLParameters(self.url)
        params.heartbeat = 600
        connection = pika.BlockingConnection(params)
        channel = connection.channel()

        # durable=True: tasks survive a broker restart.
        channel.queue_declare(queue=self.queue, durable=True)

        # prefetch_count=1: don't grab a second task while we are still
        # busy with one. This is what spreads work across workers.
        channel.basic_qos(prefetch_count=1)

        channel.basic_consume(queue=self.queue, on_message_callback=self._on_message)
        log.info("waiting for tasks on queue %r", self.queue)
        try:
            channel.start_consuming()
        finally:
            connection.close()

    def _on_message(self, channel, method, properties, body):
        try:
            task = json.loads(body)
        except ValueError:
            # Not JSON -- no point retrying, drop it (or let a
            # dead-letter queue catch it if one is configured).
            log.error("discarding malformed message: %r", body[:200])
            channel.basic_reject(delivery_tag=method.delivery_tag, requeue=False)
            return

        log.info("processing task %s", task.get("task_id", "<no id>"))
        try:
            self.handler(task)
        except Exception:
            log.exception("task failed: %s", task.get("task_id", "<no id>"))
            channel.basic_reject(delivery_tag=method.delivery_tag, requeue=False)
        else:
            channel.basic_ack(delivery_tag=method.delivery_tag)


if __name__ == "__main__":
    # Standalone smoke test: just print whatever tasks arrive.
    logging.basicConfig(level=logging.INFO)
    Consumer(handler=lambda task: log.info("got task: %s", task)).run()
