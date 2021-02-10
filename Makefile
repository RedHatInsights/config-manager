run:
	go run .

build:
	CGO_ENABLED=0 go build -o config_manager main.go

build-image:
	docker build -t config-manager-poc .

migrate-up:
	migrate -database ${POSTGRES_URL} -path db/migrations/ up

migrate-down:
	migrate -database ${POSTGRES_URL} -path db/migrations/ down