OUT_DIR=../generated

all:
	mkdir -p $(OUT_DIR)
	protoc \
		--proto_path=./ \
		--go_out=$(OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) \
  		--go-grpc_opt=paths=source_relative \
		./v1/events/*.proto \
		./v1/grpc/*.proto