"""Worker entrypoint: consume upscale tasks from RabbitMQ.

For each task the worker downloads the input image from S3, upscales it
with the FSRCNN model, and uploads the result back to S3.

Expected task message (JSON):
    {
        "task_id": "abc123",
        "input_key": "uploads/photo.png",
        "output_key": "results/photo.png",
        "scale": 3
    }
"""

import logging
import os

from dotenv import load_dotenv

load_dotenv()

from app.consumer import Consumer
from app.storage_client import Storage
from app.handler.base import UpscaleHandler
from app.handler.fsrcnn import FSRCNNUpscaler, UPSCALE_SCALE

log = logging.getLogger(__name__)

SUPPORTED_SCALES = (2, 3, 4)


def make_handler(storage: Storage, upscalers: dict[int, UpscaleHandler]):
    """Build the task handler that ties storage and the model together."""

    def handle(task):
        input_key = task["input_key"]
        output_key = task["output_key"]
        scale = task.get("scale", UPSCALE_SCALE)
        upscaler = upscalers.get(scale, upscalers[UPSCALE_SCALE])

        data = storage.download(input_key)
        result = upscaler.upscale_bytes(data)
        storage.upload(output_key, result)
        log.info("upscaled %s -> %s (scale=%d)", input_key, output_key, scale)

    return handle


def main():
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    )

    log.info("starting worker (pid %d)", os.getpid())

    storage = Storage()
    storage.ensure_bucket()

    # Load one model per supported scale at startup, not once per task,
    # so a task's requested scale can be honored without a reload.
    upscalers = {
        scale: FSRCNNUpscaler(scale=scale, weights=f"weights/fsrcnn_x{scale}.pth")
        for scale in SUPPORTED_SCALES
    }
    log.info("worker ready")

    consumer = (Consumer(handler=make_handler(storage, upscalers)))
    consumer.run()


if __name__ == "__main__":
    main()
