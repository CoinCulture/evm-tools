PACKAGES=$(shell go list ./... | grep -v '/vendor/')

all: install 

install: 
	@go install $(PACKAGES)

.PHONY: install 
