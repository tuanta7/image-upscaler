"""Upscale an image using the FSRCNN AI model.

How it works, in one paragraph:
An image can be split into "brightness" (Y) and "color" (Cb, Cr) channels.
Our eyes are much more sensitive to brightness detail than to color detail,
so the AI model only upscales the brightness channel, and we upscale the
color channels with plain resizing. This makes the model small and fast.

Standalone usage:
    python -m app.upscaler photo.png photo_big.png --scale 3 --weights fsrcnn_x3.pth

Pretrained weights: https://github.com/yjn870/FSRCNN-pytorch
"""

import logging
import os

import cv2
import numpy as np
import torch
from torch import nn

from app.handler.base import UpscaleHandler

log = logging.getLogger(__name__)

UPSCALE_SCALE = int(os.environ.get("UPSCALE_SCALE", "3"))
WEIGHTS_PATH = os.environ.get("WEIGHTS_PATH", f"weights/fsrcnn_x{UPSCALE_SCALE}.pth")


class FSRCNN(nn.Module):
    """The FSRCNN neural network.

    It takes a small brightness image and outputs a bigger one.
    It has three stages:
      1. first_part: look at the image and extract features
      2. mid_part:   transform those features
      3. last_part:  build the enlarged image from the features
    """

    def __init__(self, scale):
        super().__init__()
        self.first_part = nn.Sequential(
            nn.Conv2d(1, 56, kernel_size=5, padding=2),
            nn.PReLU(56),
        )
        self.mid_part = nn.Sequential(
            nn.Conv2d(56, 12, kernel_size=1), nn.PReLU(12),
            nn.Conv2d(12, 12, kernel_size=3, padding=1), nn.PReLU(12),
            nn.Conv2d(12, 12, kernel_size=3, padding=1), nn.PReLU(12),
            nn.Conv2d(12, 12, kernel_size=3, padding=1), nn.PReLU(12),
            nn.Conv2d(12, 12, kernel_size=3, padding=1), nn.PReLU(12),
            nn.Conv2d(12, 56, kernel_size=1), nn.PReLU(56),
        )
        # This is the layer that actually makes the image bigger:
        # stride=scale means the output is `scale` times larger.
        self.last_part = nn.ConvTranspose2d(
            56, 1, kernel_size=9, stride=scale,
            padding=4, output_padding=scale - 1,
        )

    def forward(self, x):
        """Run the image through the three stages."""
        x = self.first_part(x)
        x = self.mid_part(x)
        return self.last_part(x)


class FSRCNNUpscaler(UpscaleHandler):
    """An FSRCNN model loaded once, ready to upscale many images."""

    def __init__(self, scale=UPSCALE_SCALE, weights=WEIGHTS_PATH):
        self.scale = scale
        # Build the model and fill it with the trained weights.
        # Without weights the model would just output noise --
        # the .pth file is the result of the training process.
        self.model = FSRCNN(scale=scale)
        self.model.load_state_dict(
            torch.load(weights, map_location="cpu", weights_only=True)
        )
        self.model.eval()  # switch to "prediction mode"

    def upscale(self, image):
        """Upscale a decoded image (numpy array) and return the result."""

        # Step 1: split the image into brightness (Y) and color (Cr, Cb).
        ycrcb = cv2.cvtColor(image, cv2.COLOR_BGR2YCrCb)

        # Step 2: upscale the whole thing with plain resizing.
        # We only keep the color channels from this; the brightness
        # channel will be replaced by the model's output below.
        height, width = ycrcb.shape[:2]
        big = cv2.resize(ycrcb, (width * self.scale, height * self.scale),
                         interpolation=cv2.INTER_CUBIC)

        # Step 3: upscale the brightness channel with the AI model.
        # The model wants numbers between 0 and 1, not 0 and 255,
        # so we divide before and multiply after.
        y = ycrcb[:, :, 0].astype(np.float32) / 255.0
        y = torch.from_numpy(y)          # numpy array -> torch tensor
        y = y.reshape(1, 1, height, width)  # shape the model expects
        with torch.no_grad():            # "just predict, don't learn"
            y_big = self.model(y)
        y_big = y_big.clamp(0, 1)        # keep values in valid range
        big[:, :, 0] = y_big.numpy().reshape(height * self.scale, width * self.scale) * 255.0

        # Step 4: merge brightness and color back into a normal image.
        return cv2.cvtColor(big, cv2.COLOR_YCrCb2BGR)


if __name__ == "__main__":
    # Standalone smoke test: upscale one file from the command line.
    import argparse

    parser = argparse.ArgumentParser(description="Upscale an image with FSRCNN")
    parser.add_argument("input", help="image to upscale")
    parser.add_argument("output", help="where to save the result")
    parser.add_argument("--scale", type=int, default=UPSCALE_SCALE, choices=[2, 3, 4])
    parser.add_argument("--weights", default=WEIGHTS_PATH, help="the trained model file (.pth)")
    args = parser.parse_args()

    upscaler = FSRCNNUpscaler(scale=args.scale, weights=args.weights)

    originalImage = cv2.imread(args.input)
    if  originalImage is None:
        raise SystemExit(f"could not read image: {args.input}")

    upscaledImage = upscaler.upscale(originalImage)
    cv2.imwrite(args.output, upscaledImage)

    print(f"saved {args.output} "
          f"({originalImage.shape[1]}x{originalImage.shape[0]} -> {upscaledImage.shape[1]}x{upscaledImage.shape[0]})")
