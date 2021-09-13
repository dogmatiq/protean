PROTO_GRPC_FILES := # Disable GRPC generation

# Note, this file depends on the protoc-gen-go-protean binary from this repo to
# be build. It has been added to .gitignore so that it excluded from the
# GO_SOURCE_FILES variable, as otherwise it would create a circular dependency.
GO_TEST_REQ += internal/testservice/service_protean.pb.go

-include .makefiles/Makefile
-include .makefiles/pkg/protobuf/v2/Makefile
-include .makefiles/pkg/go/v1/Makefile

%_protean.pb.go: %.proto $(PROTOC_COMMAND) artifacts/protobuf/bin/go.mod artifacts/protobuf/args/common artifacts/protobuf/args/go $(GO_DEBUG_DIR)/protoc-gen-go-protean
	PATH="$(GO_DEBUG_DIR):$(MF_PROJECT_ROOT)/artifacts/protobuf/bin:$$PATH" $(PROTOC_COMMAND) \
		--proto_path="$(dir $(PROTOC_COMMAND))../include" \
		--go-protean_opt=module=$$(go list -m) \
		--go-protean_out=. \
		$$(cat artifacts/protobuf/args/common artifacts/protobuf/args/go) \
		$(MF_PROJECT_ROOT)/$(@D)/*.proto

.makefiles/%:
	@curl -sfL https://makefiles.dev/v1 | bash /dev/stdin "$@"
