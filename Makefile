# Установим параметры подключения к базе данных для тестов
ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := user=user password=password dbname=db host=localhost port=5432 sslmode=disable
endif

MIGRATION_FOLDER=$(CURDIR)/migrations
LOCAL_BIN=$(GOPATH)/bin
MOCKGEN_TAG=1.6.0

.PHONY: compose-rs
compose-rs:
	make compose-rm
	make compose-up

.PHONY: compose-up
compose-up:
	docker-compose -p homework-1 up -d

.PHONY: compose-rm
compose-rm:
	docker-compose -p homework-1 rm -fvs

.PHONY: migrate-up
migrate-up:
	./migration.sh up

.PHONY: migrate-status
migrate-status:
	./migration.sh status

.PHONY: clean-db
clean-db:
	psql "$(POSTGRES_SETUP_TEST)" -c "TRUNCATE TABLE orders CASCADE;"

.PHONY: test
test:
	$(info running tests...)
	go test ./...

.PHONY: .generate-mockgen-deps
.generate-mockgen-deps:
ifeq ($(wildcard $(MOCKGEN_BIN)),)
	@GOBIN=$(LOCAL_BIN) go install github.com/golang/mock/mockgen@$(MOCKGEN_TAG)
endif

.PHONY: .generate-mockgen
generate-mockgen:
	PATH="$(LOCAL_BIN):$$PATH" go generate -x -run=mockgen ./...

.PHONY: .generate-mock
generate-mock:
	find . -name '*_mock.go' -delete
	$(MOCKGEN_BIN) mockgen -source ./storage.go -destination=./mocks/storage_mock.go -package=storage_mock
	$(MOCKGEN_BIN) mockgen -source ./moduleInterface.go -destination=./mocks/module_mock.go -package=module_mock
