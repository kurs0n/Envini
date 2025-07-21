DB_CONTAINER=envini-audit-postgres
DB_IMAGE=postgres:16-alpine
DB_PORT=5433
DB_USER=envini
DB_PASSWORD=envini
DB_NAME=envini

.PHONY: db-build db-run db-stop db-remove db-logs db-shell

db-build:
	docker pull $(DB_IMAGE)

db-run:
	docker run -d --name $(DB_CONTAINER) -e POSTGRES_USER=$(DB_USER) -e POSTGRES_PASSWORD=$(DB_PASSWORD) -e POSTGRES_DB=$(DB_NAME) -p $(DB_PORT):$(DB_PORT) $(DB_IMAGE)

db-stop:
	docker stop $(DB_CONTAINER)

db-remove:
	docker rm $(DB_CONTAINER)

db-logs:
	docker logs $(DB_CONTAINER)

db-shell:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)