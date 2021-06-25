run:
	go run .

build:
	CGO_ENABLED=0 go build -o cm .

build-image:
	podman build -t config-manager-poc .

# export POSTGRES_URL=postgres://insights:insights@localhost:5432/config-manager?sslmode=disable (example)
migrate-up:
	migrate -database ${POSTGRES_URL} -path db/migrations/ up

migrate-down:
	migrate -database ${POSTGRES_URL} -path db/migrations/ down