PROTO_GRPC_FILES := # Disable GRPC generation
GENERATED_FILES += $(foreach f,$(PROTO_FILES:.proto=_protean.pb.go),$(if $(findstring /_,/$f),,$f))

-include .makefiles/Makefile
-include .makefiles/pkg/protobuf/v2/Makefile
-include .makefiles/pkg/go/v1/Makefile

run:
	@echo $(GENERATED_FILES)

%_protean.pb.go: %.proto $(PROTOC_COMMAND) artifacts/protobuf/bin/go.mod artifacts/protobuf/args/common artifacts/protobuf/args/go $(GO_DEBUG_DIR)/protoc-gen-go-protean
	PATH="$(GO_DEBUG_DIR):$(MF_PROJECT_ROOT)/artifacts/protobuf/bin:$$PATH" $(PROTOC_COMMAND) \
		--proto_path="$(dir $(PROTOC_COMMAND))../include" \
		--go-protean_opt=module=$$(go list -m) \
		--go-protean_out=. \
		$$(cat artifacts/protobuf/args/common artifacts/protobuf/args/go) \
		$(MF_PROJECT_ROOT)/$(@D)/*.proto

.makefiles/%:
	@curl -sfL https://makefiles.dev/v1 | bash /dev/stdin "$@"
