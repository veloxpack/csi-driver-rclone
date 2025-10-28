SHELL=/bin/bash -O extglob -c
generate:
	go run gen/generate_setters/!(*_test).go -e gen/generate_setters/test_data,gen/generate_setters/test_data/another_package
