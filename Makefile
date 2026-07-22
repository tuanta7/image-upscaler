setup:
	docker compose -f docker-compose.dev.yaml up --build -d 

start-worker:
	$(MAKE) -C worker start

start-scheduler:
	$(MAKE) -C scheduler start