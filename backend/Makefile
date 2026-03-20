.PHONY: run build tidy pm2-start pm2-stop pm2-restart pm2-logs db-bootstrap db-upgrade db-verify-schema ops-retention

APP_NAME=jedug-api

run:
	go run ./cmd/api

build:
	go build -o bin/$(APP_NAME) ./cmd/api

tidy:
	go mod tidy

# PM2 process management
pm2-start: build
	pm2 start bin/$(APP_NAME) --name $(APP_NAME)

pm2-stop:
	pm2 stop $(APP_NAME)

pm2-restart: build
	pm2 restart $(APP_NAME)

pm2-logs:
	pm2 logs $(APP_NAME)

db-bootstrap:
	./scripts/bootstrap_db.sh fresh

db-upgrade:
	./scripts/bootstrap_db.sh upgrade

db-verify-schema:
	./scripts/verify_schema_governance.sh

ops-retention:
	go run ./cmd/maintenance
