# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# We need all the Make variables exported as env vars.
# Note that the ?= operator works regardless.

# Enable Go modules.
export GO111MODULE=on

VERSION ?= v0.0.1
DOCKER_PLATFORM ?= linux/amd64

IMG ?= kaasops/echo-server:${VERSION}

# Print the help menu.
.PHONY: help
help:
	@grep -hE '^[ a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'

# Build the binary
.PHONY: build
build: vet;$(info $(M)...Build the binary.)  @ ## Build the binary.
	go build -o echo-server .

.PHONY: docker-build
docker-build: ## Build docker image
	docker build --platform ${DOCKER_PLATFORM} -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image
	docker push ${IMG}