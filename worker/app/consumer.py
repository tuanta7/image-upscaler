import json
import logging
import os
import time

import pika
import pika.exceptions

log = logging.getLogger(__name__)

RABBITMQ_URL = os.environ.get("RABBITMQ_URL", "amqp://rabbitmq:password@localhost:5672/")
TASK_QUEUE = os.environ.get("TASK_QUEUE", "upscale.tasks")
RESULTS_QUEUE = os.environ.get("RESULTS_QUEUE", "upscale.results")

class Consumer:
    def __init__(self, handler, url=RABBITMQ_URL, queue=TASK_QUEUE, results_queue=RESULTS_QUEUE):
        self.handler = handler
        self.url = url
        self.queue = queue
        self.results_queue = results_queue

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

        # Log host/port only -- the full URL contains credentials.
        log.info("connecting to RabbitMQ at %s:%s", params.host, params.port)
        connection = pika.BlockingConnection(params)
        log.info("connected to RabbitMQ")
        channel = connection.channel()

        # durable=True: tasks survive a broker restart.
        channel.queue_declare(queue=self.queue, durable=True)
        channel.queue_declare(queue=self.results_queue, durable=True)

        # prefetch_count=1: don't grab a second task while we are still
        # busy with one. This is what spreads work across workers.
        channel.basic_qos(prefetch_count=1)

        channel.basic_consume(queue=self.queue, on_message_callback=self._process)
        log.info("waiting for tasks on queue %r", self.queue)
        try:
            channel.start_consuming()
        finally:
            connection.close()

    def _publish_status(self, channel, task_id, status):
        if not task_id:
            return
        channel.basic_publish(
            exchange="",
            routing_key=self.results_queue,
            body=json.dumps({"task_id": task_id, "status": status}),
            properties=pika.BasicProperties(content_type="application/json"),
        )

    def _process(self, channel, method, properties, body):
        try:
            task = json.loads(body)
        except ValueError:
            # Not JSON -- no point retrying, drop it (or let a
            # dead-letter queue catch it if one is configured).
            log.error("discarding malformed message: %r", body[:200])
            channel.basic_reject(delivery_tag=method.delivery_tag, requeue=False)
            return

        task_id = task.get("task_id")
        log.info("processing task %s", task_id or "<no id>")
        self._publish_status(channel, task_id, "processing")
        try:
            self.handler(task)
        except Exception:
            # The handler is a plugged-in model + S3 calls -- anything from
            # a corrupt upload to a transient S3 error can land here. One
            # bad task must not take the whole worker down with it.
            log.exception("task failed: %s", task_id or "<no id>")
            channel.basic_reject(delivery_tag=method.delivery_tag, requeue=False)
            self._publish_status(channel, task_id, "failed")
        else:
            channel.basic_ack(delivery_tag=method.delivery_tag)
            self._publish_status(channel, task_id, "done")


if __name__ == "__main__":
    # Standalone smoke test: just print whatever tasks arrive.
    logging.basicConfig(level=logging.INFO)
    consumer = Consumer(handler=lambda task: log.info("got task: %s", task))
    consumer.run()
