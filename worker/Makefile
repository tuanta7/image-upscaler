setup:
	source .venv/bin/activate
	uv pip install -r requirements.txt

start:
	python main.py

try-upscale:
	python -m app.handler.fsrcnn images/original.jpg images/upscaled_x4.jpg --scale 4 --weights weights/fsrcnn_x4.pth