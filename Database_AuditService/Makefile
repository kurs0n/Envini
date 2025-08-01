

.PHONY: run stop clean build logs

run:
	./run-db.sh

stop:
	docker stop envini-audit-postgres 2>/dev/null || echo "Container not running"

clean:
	docker stop envini-audit-postgres 2>/dev/null || true
	docker rm envini-audit-postgres 2>/dev/null || true

build:
	docker build -t envini-audit-postgres .

logs:
	docker logs -f envini-audit-postgres

connect:
	docker exec -it envini-audit-postgres psql -U envini -d envini -p 5433

status:
	docker ps | grep envini-audit-postgres || echo "Database not running"
