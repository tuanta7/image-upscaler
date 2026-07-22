"""Interface for upscale handlers. Implement `upscale` to add a new model."""

from abc import ABC, abstractmethod

import cv2
import numpy as np


class UpscaleHandler(ABC):
    """Interface for upscaling images. Implement `upscale` for a new model."""

    @abstractmethod
    def upscale(self, image):
        """Upscale a decoded image (numpy array) and return the result."""

    def upscale_bytes(self, data, ext=".png"):
        """Upscale an encoded image (raw bytes) and return encoded bytes."""
        image = cv2.imdecode(np.frombuffer(data, np.uint8), cv2.IMREAD_COLOR)
        if image is None:
            raise ValueError("could not decode image")
        result = self.upscale(image)
        ok, encoded = cv2.imencode(ext, result)
        if not ok:
            raise ValueError(f"could not encode result as {ext}")
        return encoded.tobytes()
