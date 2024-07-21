ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := "postgres://user:password@localhost:5432/db?sslmode=disable"
endif

MIGRATION_FOLDER=$(CURDIR)/migrations
LOCAL_BIN=$(GOPATH)/bin
MOCKGEN_TAG=1.6.0

ORDERS_PROTO_PATH=api/proto/orders_grpc/v1

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
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" up

.PHONY: migrate-status
migrate-status:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" status

.PHONY: clean-db
clean-db:
	psql "$(POSTGRES_SETUP_TEST)" -c "TRUNCATE TABLE orders CASCADE;"

.PHONY: test
test:
	$(info running tests...)
	go test ./...

.PHONY: integration-tests
integration-tests:
	$(info running integration tests...)
	go test -tags=integration ./...

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
	$(MOCKGEN_BIN) mockgen -source ./module_interface.go -destination=./mocks/module_mock.go -package=module_mock

.PHONY: .bin-deps
.bin-deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: .generate-grpc-stubs
generate-grpc-stubs: .bin-deps
	mkdir -p ./pkg/$(ORDERS_PROTO_PATH)

	protoc -I ./api/proto \
		--go_out=./pkg/$(ORDERS_PROTO_PATH) --go_opt=paths=source_relative \
		--go-grpc_out=./pkg/$(ORDERS_PROTO_PATH) --go-grpc_opt=paths=source_relative \
		$(ORDERS_PROTO_PATH)/orders.proto

.PHONY: run-prometheus
run-prometheus:
	prometheus --config.file=config/prometheus.yml

.PHONY: vendor-proto
vendor-proto: vendor-proto/google/protobuf vendor-proto/google/api vendor-proto/protoc-gen-openapiv2/options vendor-proto/validate

# Устанавливаем proto описания protoc-gen-openapiv2/options
vendor-proto/protoc-gen-openapiv2/options:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/grpc-ecosystem/grpc-gateway vendor.proto/grpc-ecosystem && \
 	cd vendor.proto/grpc-ecosystem && \
	git sparse-checkout set --no-cone protoc-gen-openapiv2/options && \
	git checkout
	mkdir -p vendor.proto/protoc-gen-openapiv2
	mv vendor.proto/grpc-ecosystem/protoc-gen-openapiv2/options vendor.proto/protoc-gen-openapiv2
	rm -rf vendor.proto/grpc-ecosystem


# Устанавливаем proto описания google/protobuf
vendor-proto/google/protobuf:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/protocolbuffers/protobuf vendor.proto/protobuf &&\
	cd vendor.proto/protobuf &&\
	git sparse-checkout set --no-cone src/google/protobuf &&\
	git checkout
	mkdir -p vendor.proto/google
	mv vendor.proto/protobuf/src/google/protobuf vendor.proto/google
	rm -rf vendor.proto/protobuf

vendor-proto/google/api:
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/googleapis/googleapis vendor.proto/googleapis && \
 	cd vendor.proto/googleapis && \
	git sparse-checkout set --no-cone google/api && \
	git checkout
	mkdir -p  vendor.proto/google
	mv vendor.proto/googleapis/google/api vendor.proto/google
	rm -rf vendor.proto/googleapis

vendor-proto/validate:
	git clone -b main --single-branch --depth=2 --filter=tree:0 \
		https://github.com/bufbuild/protoc-gen-validate vendor.proto/tmp && \
		cd vendor.proto/tmp && \
		git sparse-checkout set --no-cone validate &&\
		git checkout
		mkdir -p vendor.proto/validate
		mv vendor.proto/tmp/validate vendor.proto/
		rm -rf vendor.proto/tmp