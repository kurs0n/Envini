.PHONY: run stop clean build logs

run:
	./run-db.sh

stop:
	docker stop envini-secret-postgres 2>/dev/null || echo "Container not running"

clean:
	docker stop envini-secret-postgres 2>/dev/null || true
	docker rm envini-secret-postgres 2>/dev/null || true

build:
	docker build -t envini-secret-postgres .

logs:
	docker logs -f envini-secret-postgres

connect:
	docker exec -it envini-secret-postgres psql -U envini -d envini -p 5433

status:
	docker ps | grep envini-secret-postgres || echo "Database not running"
