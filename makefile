test: docker_build_test
	docker compose down
	docker compose up -d
	docker compose exec -T gophermart go test  -v -tags="integration unit" ./...
	docker compose down

unit_test:
	go test ./...

docker_build:
	docker build -t service -f gophermart.Dockerfile .
	docker build -t accrual -f accrual.Dockerfile .

docker_build_test:
	docker build -t service_test  -f gophermart.Dockerfile . --target=build

docker_run:
	docker run --publish 8080:8080 service


test-db-up:
	docker compose -f docker-compose.yml up --build db

test-db-down:
	docker compose -f docker-compose.yml down --volumes db