"""Upscale an image using the FSRCNN AI model.

How it works, in one paragraph:
An image can be split into "brightness" (Y) and "color" (Cb, Cr) channels.
Our eyes are much more sensitive to brightness detail than to color detail,
so the AI model only upscales the brightness channel, and we upscale the
color channels with plain resizing. This makes the model small and fast.

Usage:
    python main.py photo.png photo_big.png --scale 3 --weights fsrcnn_x3.pth

Pretrained weights: https://github.com/yjn870/FSRCNN-pytorch
"""

import argparse

import cv2
import numpy as np
import torch
from torch import nn


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


def upscale(image, model, scale):
    """Upscale one image (a numpy array) and return the result."""

    # Step 1: split the image into brightness (Y) and color (Cr, Cb).
    ycrcb = cv2.cvtColor(image, cv2.COLOR_BGR2YCrCb)

    # Step 2: upscale the whole thing with plain resizing.
    # We only keep the color channels from this; the brightness
    # channel will be replaced by the model's output below.
    height, width = ycrcb.shape[:2]
    big = cv2.resize(ycrcb, (width * scale, height * scale), interpolation=cv2.INTER_CUBIC)

    # Step 3: upscale the brightness channel with the AI model.
    # The model wants numbers between 0 and 1, not 0 and 255,
    # so we divide before and multiply after.
    y = ycrcb[:, :, 0].astype(np.float32) / 255.0
    y = torch.from_numpy(y)          # numpy array -> torch tensor
    y = y.reshape(1, 1, height, width)  # shape the model expects
    with torch.no_grad():            # "just predict, don't learn"
        y_big = model(y)
    y_big = y_big.clamp(0, 1)        # keep values in valid range
    big[:, :, 0] = y_big.numpy().reshape(height * scale, width * scale) * 255.0

    # Step 4: merge brightness and color back into a normal image.
    return cv2.cvtColor(big, cv2.COLOR_YCrCb2BGR)


def main():
    # Read the command-line arguments.
    parser = argparse.ArgumentParser(description="Upscale an image with FSRCNN")
    parser.add_argument("input", help="image to upscale")
    parser.add_argument("output", help="where to save the result")
    parser.add_argument("--scale", type=int, default=3, choices=[2, 3, 4])
    parser.add_argument("--weights", default="fsrcnn_x3.pth",
                        help="the trained model file (.pth)")
    args = parser.parse_args()

    # Build the model and fill it with the trained weights.
    # Without weights the model would just output noise --
    # the .pth file is the result of the training process.
    model = FSRCNN(scale=args.scale)
    model.load_state_dict(torch.load(args.weights, weights_only=True))
    model.eval()  # switch to "prediction mode"

    # Load the image, upscale it, save the result.
    image = cv2.imread(args.input)
    if image is None:
        raise SystemExit(f"could not read image: {args.input}")

    result = upscale(image, model, args.scale)
    cv2.imwrite(args.output, result)
    print(f"saved {args.output} "
          f"({image.shape[1]}x{image.shape[0]} -> {result.shape[1]}x{result.shape[0]})")


if __name__ == "__main__":
    main()
