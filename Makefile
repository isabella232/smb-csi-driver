THIS_FILE := $(lastword $(MAKEFILE_LIST))

build:
	go build .

running	:=	"$(shell docker inspect -f '{{.State.Running}}' "kind-registry" 2>/dev/null || true)"
image-local-registry: SHELL:=/bin/bash
image-local-registry:
	[ $(running) != "true" ] && docker run \
	    -d --restart=always -p "5000:5000" --name "kind-registry" \
	    registry:2 || true
	docker build --no-cache -t cfpersi/smb-plugin .
	docker tag cfpersi/smb-plugin localhost:5000/cfpersi/smb-plugin:local-test
	docker push localhost:5000/cfpersi/smb-plugin:local-test

start-docker:
	start-docker &
	echo 'until docker info; do sleep 5; done' >/usr/local/bin/wait_for_docker
	chmod +x /usr/local/bin/wait_for_docker
	timeout 300 wait_for_docker

kill-docker:
	pkill dockerd

test:
	@$(MAKE) -f $(THIS_FILE) image-local-registry
	go get github.com/onsi/ginkgo/ginkgo
	~/go/bin/ginkgo -flakeAttempts 3 -race -focus "(Identity|Node) Service"

e2e: SHELL:=/bin/bash
e2e:
	@$(MAKE) -f $(THIS_FILE) image-local-registry
	go get github.com/onsi/ginkgo/ginkgo
	cd test && pwd && ~/go/bin/ginkgo -r -focus "CSI Volumes"

fly:
	fly -t persi execute -p -c ~/workspace/smb-volume-k8s-release/smb-csi-driver/ci/integration-tests.yml -i smb-volume-k8s-release=$$HOME/workspace/smb-volume-k8s-release

fly-e2e:
	fly -t persi execute -p -c ~/workspace/smb-volume-k8s-release/smb-csi-driver/ci/e2e-tests.yml -i smb-volume-k8s-release=$$HOME/workspace/smb-volume-k8s-release

kustomize:
	kubectl apply --kustomize ./overlays/deploy

kustomize-delete:
	kubectl delete --kustomize ./overlays/deploy