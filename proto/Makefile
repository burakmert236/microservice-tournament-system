OUT_DIR=../generated

all:
	mkdir -p $(OUT_DIR)
	protoc \
		--proto_path=. \
		--go_out=$(OUT_DIR) \
		--go_opt=paths=source_relative \
		./events/*.proto \
		./grpc/*.proto