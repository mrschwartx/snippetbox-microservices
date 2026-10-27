PROFILE ?= snippetbox
CPUS ?= 4
MEMORY ?= 6144
NAMESPACE ?= snippetbox
IMAGE_TAG ?= local
REPO_ROOT ?= $(abspath $(CURDIR)/..)
DOCKER_DIR := $(REPO_ROOT)/snippetbox-docker
IMAGES := snippetbox-msa-backend-user snippetbox-msa-backend-snippet snippetbox-msa-frontend
KUBECTL := kubectl --context $(PROFILE)

.PHONY: all cluster build load secrets render deploy status url tunnel undeploy delete

all: cluster build load deploy

cluster:
	minikube start -p $(PROFILE) --driver=docker --cpus=$(CPUS) --memory=$(MEMORY)
	minikube -p $(PROFILE) addons enable ingress
	until $(KUBECTL) -n ingress-nginx get deployment ingress-nginx-controller >/dev/null 2>&1; do sleep 2; done
	$(KUBECTL) -n ingress-nginx rollout status deployment/ingress-nginx-controller --timeout=300s

build:
	@for svc in $(IMAGES); do docker build -t $$svc:$(IMAGE_TAG) -f $(DOCKER_DIR)/$$svc/Dockerfile $(REPO_ROOT)/$$svc || exit 1; done

load:
	@for svc in $(IMAGES); do minikube -p $(PROFILE) image load $$svc:$(IMAGE_TAG) || exit 1; done

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
	$(KUBECTL) -n $(NAMESPACE) rollout status deployment/snippetbox-msa-backend-user --timeout=300s
	$(KUBECTL) -n $(NAMESPACE) rollout status deployment/snippetbox-msa-backend-snippet --timeout=300s
	$(KUBECTL) -n $(NAMESPACE) rollout status deployment/snippetbox-msa-frontend --timeout=300s

status:
	$(KUBECTL) -n $(NAMESPACE) get pods,svc,ingress,pvc

url:
	@echo http://$$(minikube -p $(PROFILE) ip)

tunnel:
	minikube -p $(PROFILE) tunnel

undeploy:
	$(KUBECTL) delete -k overlay --ignore-not-found

delete:
	minikube delete -p $(PROFILE)
