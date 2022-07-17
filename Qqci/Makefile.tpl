SHELL=/bin/bash

APPNAME         ?= {{.Appname}}
PKGNAME         ?= $(APPNAME).tar.gz
BUILD_DIR        = $(shell pwd)
TEMP_OUTPUT_DIR  = $(shell pwd)/_output/$(APPNAME)


tar: build package
	@echo -e "======\033[44m all done \033[0m"

build:
	go mod tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APPNAME) ./
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APPNAME).exe ./
package:
	@-rm -rf $(TEMP_OUTPUT_DIR)                    >/dev/null 2>&1
	mkdir -p $(TEMP_OUTPUT_DIR)                    >/dev/null 2>&1
	cp -rL $(BUILD_DIR)/$(APPNAME)      $(TEMP_OUTPUT_DIR)/
	cp -rL $(BUILD_DIR)/$(APPNAME).exe      $(TEMP_OUTPUT_DIR)/
	cp -rL $(BUILD_DIR)/README.md       $(TEMP_OUTPUT_DIR)/
	cd $(TEMP_OUTPUT_DIR)/.. && tar -czf $(PKGNAME) $(APPNAME)

