CLUSTER ?= snippetbox
INGRESS_NGINX_MANIFEST ?= https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.1/deploy/static/provider/kind/deploy.yaml
NAMESPACE ?= snippetbox
IMAGE_TAG ?= local
REPO_ROOT ?= $(abspath $(CURDIR)/..)
DOCKER_DIR := $(REPO_ROOT)/snippetbox-docker
IMAGES := snippetbox-backend-user snippetbox-backend-snippet snippetbox-frontend
KUBECTL := kubectl --context kind-$(CLUSTER)

.PHONY: all cluster build load secrets render deploy status url undeploy delete

all: cluster build load deploy

cluster:
	kind create cluster --name $(CLUSTER) --config kind-config.yaml
	$(KUBECTL) apply -f $(INGRESS_NGINX_MANIFEST)
	$(KUBECTL) -n ingress-nginx rollout status deployment/ingress-nginx-controller --timeout=300s

build:
	@for svc in $(IMAGES); do docker build -t $$svc:$(IMAGE_TAG) -f $(DOCKER_DIR)/$$svc/Dockerfile $(REPO_ROOT)/$$svc || exit 1; done

load:
	@for svc in $(IMAGES); do kind load docker-image $$svc:$(IMAGE_TAG) --name $(CLUSTER) || exit 1; done

SECRETS_FILE := overlay/secrets/snippetbox.env

$(SECRETS_FILE):
	@mkdir -p $(dir $@)
	@umask 077 && printf 'MYSQL_ROOT_PASSWORD=%s\nUSER_DB_PASSWORD=%s\nSNIPPET_DB_PASSWORD=%s\nJWT_SECRET=%s\n' "$$(openssl rand -hex 24)" "$$(openssl rand -hex 24)" "$$(openssl rand -hex 24)" "$$(openssl rand -hex 32)" > $@

secrets: $(SECRETS_FILE)

render: secrets
	@$(KUBECTL) kustomize overlay

deploy: secrets
	$(KUBECTL) apply -k overlay
	$(KUBECTL) -n $(NAMESPACE) rollout status statefulset/snippetbox-mysql --timeout=300s
	$(KUBECTL) -n $(NAMESPACE) rollout status deployment/snippetbox-backend-user --timeout=300s
	$(KUBECTL) -n $(NAMESPACE) rollout status deployment/snippetbox-backend-snippet --timeout=300s
	$(KUBECTL) -n $(NAMESPACE) rollout status deployment/snippetbox-frontend --timeout=300s

status:
	$(KUBECTL) -n $(NAMESPACE) get pods,svc,ingress,pvc

url:
	@echo http://snippetbox.localtest.me

undeploy:
	$(KUBECTL) delete -k overlay --ignore-not-found

delete:
	kind delete cluster --name $(CLUSTER)
