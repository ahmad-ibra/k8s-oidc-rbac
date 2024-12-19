# Variables
docker_repo = ahmadibraspectrocloud/pod-service
image_tag = latest
kind_cluster = kind

.PHONY: start-dex start-dex-ui build kind-load deploy port-forward kind-deploy

# starts kind cluster, installs dex, starts port-forwarding to dex and starts the dex-ui
start-dex:
	kind create cluster
	helm install dex dex/dex --values dex-values.yaml

	until kubectl get pods --namespace default -l "app.kubernetes.io/name=dex,app.kubernetes.io/instance=dex" | grep Running; do echo "Dex is not ready yet. Retrying..."; sleep 2; done
	kubectl wait --namespace default --for=condition=ready pod -l "app.kubernetes.io/name=dex,app.kubernetes.io/instance=dex" --timeout=90s

	@echo "Starting port-forwarding 5556:5556 for Dex..."
	export POD_NAME=$$(kubectl get pods --namespace default -l "app.kubernetes.io/name=dex,app.kubernetes.io/instance=dex" -o jsonpath="{.items[0].metadata.name}"); \
	kubectl --namespace default port-forward $$POD_NAME 5556:5556

start-dex-ui:
	@echo "Starting Dex UI..."
	./bin/dex-ui

# Build the Docker image
build:
	docker build -t $(docker_repo):$(image_tag) .

# Load the Docker image into the local kind cluster
kind-load:
	kind load docker-image $(docker_repo):$(image_tag) --name $(kind_cluster)

# Deploy the application to the kind cluster
deploy:
	kubectl apply -f manifest.yaml

port-forward:
	until kubectl port-forward service/pod-service 8000:80; do echo "Failed attempt at port-forwardig. Retrying..."; sleep 2; done

kind-deploy: build kind-load deploy port-forward
