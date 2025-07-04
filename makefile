PROTO_DIR=proto
PROTO_FILE=$(PROTO_DIR)/ticket.proto

generate-proto:
	protoc \
		--go_out=paths=source_relative:. \
		--go-grpc_out=paths=source_relative:. \
		$(PROTO_FILE)