setup:
	source .venv/bin/activate
	uv pip install -r requirements.txt

try-upscale:
	python -m app.upscale_handler ./images/original.jpg ./images/upscaled.jpg --scale 3 --weights weights/fsrcnn_x3.pth