# Variables
docker_repo = ahmadibraspectrocloud/pod-service
image_tag = latest
kind_cluster = kind

.PHONY: all build push deploy kind-load kind-deploy

# Build the Docker image
build:
	docker build -t $(docker_repo):$(image_tag) .

# Push the Docker image to DockerHub
push:
	docker push $(docker_repo):$(image_tag)

# Build and push the Docker image
all: build push

# Load the Docker image into the local kind cluster
kind-load:
	kind load docker-image $(docker_repo):$(image_tag) --name $(kind_cluster)

# Deploy the application to the kind cluster
deploy:
	kubectl apply -f pod-service.yaml

port-forward:
	until kubectl port-forward service/pod-service 8000:80; do echo "Failed attempt at port-forwardig. Retrying..."; sleep 2; done

undeploy:
	kubectl delete -f pod-service.yaml

kind-deploy: build kind-load deploy port-forward
