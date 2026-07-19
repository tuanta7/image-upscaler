# Worker

FSRCNN (Fast Super-Resolution Convolutional Neural Network) is a model tailored for real-time image super-resolution.

Try it standalone, all pre-trained weights are from https://github.com/yjn870/FSRCNN-pytorch

```sh
python -m app.upscale_handler photo.png photo_big.png --scale 3 --weights fsrcnn_x3.pth
```

## Concepts

- **Model/Neural Network**: A function with millions of adjustable numbers (parameters) inside. You feed it an input (a small image), it produces an output (a bigger image). In code it's the `FSRCNN` class — a pipeline of layers the input flows through.
- **Weights**: The learned values of those parameters.
- **Training**: The process of finding good values.
- **Inference**: Puts learned knowledge to work by making predictions.
- **Tensor**: PyTorch's array type, like a numpy array, images are moved between numpy (what OpenCV uses) and tensors (what the model wants)
- **Convolution** (Conv2d): A layer that slides a small filter over the image to detect local patterns (edges, textures). Almost all of FSRCNN is
  convolutions.
- **Channels**: An image can be split into brightness (Y) and color (Cb, Cr) channels a.k.a the YCrCb color space. Human eyes are much more sensitive to brightness detail than color detail, so the model only upscales the brightness channel; color is upscaled with plain (bicubic) resizing. One channel (Y) instead of three (Y, Cb, Cr) makes the model small and fast.
- **Scale**: The upscaling factor (2x, 3x, 4x). It's baked into the last layer's shape, so each scale needs its own weights file

## Python Concepts

Quick glossary of the libraries used

| Library           | What it does here                                          |
| ----------------- | ---------------------------------------------------------- |
| `torch` (PyTorch) | Defines and runs the neural network                        |
| `cv2` (OpenCV)    | Reads/writes/encodes images, color conversion, resizing    |
| `numpy`           | The array format images live in between OpenCV and PyTorch |
| `pika`            | Talks to RabbitMQ (consume task messages, ack/reject)      |
| `boto3`           | Talks to S3/MinIO (download inputs, upload results)        |

### Virtual Environments

Reference: [What are Virtual Environments?](https://fastapi.tiangolo.com/virtual-environments/#what-are-virtual-environments)

A venv is a project-local folder of installed packages, so this project's dependency versions don't clash with other projects. Activate it
before running anything.

```sh
uv venv                              # create .venv
source .venv/bin/activate            # use it in this shell
uv pip install -r requirements.txt   # install dependencies into it
uv pip freeze > requirements.txt     # snapshot exact installed versions
```

### Modules and Imports

Every `.py` file is a module.

- The `from app.storage_client import Storage` statement means get the `Storage` in the folder `app`, file `storage_client.py`
- Run a module inside a package with `python -m app.consumer` so imports resolve; plain `python main.py` works because `main.py` sits at the project root.

Checks whether your Python file is being run directly or being imported as a module into another script

```py
if __name__ == "__main__":
    print("Hello, World!")
```

Every module in `app/` uses this to ship a standalone smoke test: importing `app.storage_client` from `main.py` does nothing extra, but running
`python -m app.storage_client` executes the round-trip test at the bottom.

### Classes

A class bundles data and the functions that operate on it.

- Names with a leading underscore are private by **convention** only
- To inherit from a parent class, pass the name of the parent class inside parentheses when defining the child class

```py
from torch import nn
class FSRCNN(nn.Module):
```

```py
class Storage:
    def __init__(self, bucket):   # constructor, runs on Storage(...)
        self.bucket = bucket      # self = this instance; attributes live on it

    def download(self, key):      # method; self is passed automatically
        ...

storage = Storage("images")       # create an instance
storage.download("photo.png")     # call a method
```

### Context Managers (`with` blocks)

```py
# `switches something on for the indented block and
# guarantees it's switched back off after, even if an error occurs
with torch.no_grad():

# for files
with open(path) as f:
```
