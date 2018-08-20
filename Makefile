# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

TAG=$(shell git describe --tags --abbrev=0 2>/dev/null)
SHA=$(shell git describe --match=NeVeRmAtCh --always --abbrev=7 --dirty)

ifeq ($(TAG),)
	VERSION=$(SHA)
else
	VERSION=$(TAG)-$(SHA)
endif

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## run the agent using locally compiled binaries
	./bin/docker-operator

deps: ## fetch dependencies using dep
	dep ensure

build: ## build agent and server locally, without docker
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags '-extldflags "-static"' -o bin/docker-operator

# deploy-kubernetes:
# 	envsubst < k8s/deployment.yml | kubectl apply -f -

clean: ## remove binaries
	rm -rf bin/*