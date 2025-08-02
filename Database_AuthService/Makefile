DB_CONTAINER=envini-auth-postgres
DB_IMAGE=postgres:16-alpine
DB_PORT=5432
DB_USER=envini
DB_PASSWORD=envini
DB_NAME=envini

.PHONY: build run stop remove logs shell

build:
	docker pull $(DB_IMAGE)

run:
	docker run -d --name $(DB_CONTAINER) -e POSTGRES_USER=$(DB_USER) -e POSTGRES_PASSWORD=$(DB_PASSWORD) -e POSTGRES_DB=$(DB_NAME) -p $(DB_PORT):5432 $(DB_IMAGE)

stop:
	docker stop $(DB_CONTAINER)

remove:
	docker rm $(DB_CONTAINER)

logs:
	docker logs $(DB_CONTAINER)

shell:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)