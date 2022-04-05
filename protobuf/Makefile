BUF_CACHE_DIR=/tmp/.bufcache

.create-cache-dir:
	mkdir -p $(BUF_CACHE_DIR)

generate: .create-cache-dir
	docker run --rm -itv $(PWD):/workspace -v $(BUF_CACHE_DIR):/.cache -w /workspace --user=$(shell id -u) bufbuild/buf generate