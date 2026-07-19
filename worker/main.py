"""Worker entrypoint: consume upscale tasks from RabbitMQ.

For each task the worker downloads the input image from S3, upscales it
with the FSRCNN model, and uploads the result back to S3.

Expected task message (JSON):
    {
        "task_id": "abc123",
        "input_key": "uploads/photo.png",
        "output_key": "results/photo.png"
    }
"""

import logging

from app.consumer import Consumer
from app.storage_client import Storage
from app.upscale_handler import Upscaler

log = logging.getLogger(__name__)


def make_handler(storage, upscaler):
    """Build the task handler that ties storage and the model together."""

    def handle(task):
        input_key = task["input_key"]
        output_key = task["output_key"]

        data = storage.download(input_key)
        result = upscaler.upscale_bytes(data)
        storage.upload(output_key, result)
        log.info("upscaled %s -> %s", input_key, output_key)

    return handle


def main():
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    )

    storage = Storage()
    storage.ensure_bucket()

    # Load the model once at startup, not once per task.
    upscaler = Upscaler()

    consumer = (Consumer(handler=make_handler(storage, upscaler)))
    consumer.run()


if __name__ == "__main__":
    main()
