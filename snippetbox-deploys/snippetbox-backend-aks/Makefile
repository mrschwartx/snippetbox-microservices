NAMESPACE ?= snippetbox
PLATFORM ?= linux/amd64
REPO_ROOT ?= $(abspath $(CURDIR)/..)
DOCKER_DIR := $(REPO_ROOT)/snippetbox-docker
IMAGES := snippetbox-backend-user snippetbox-backend-snippet snippetbox-frontend
KUBECTL := kubectl
ifndef IMAGE_TAG
IMAGE_TAG := $(shell git rev-parse --short HEAD 2>/dev/null || date +%Y%m%d%H%M%S)
endif
ACR_NAME ?=
RESOURCE_GROUP ?=
AKS_NAME ?=
DOMAIN ?=
KEYVAULT_CERT_URI ?=
REGISTRY := $(ACR_NAME).azurecr.io
REQUIRED := require-ACR_NAME require-DOMAIN require-KEYVAULT_CERT_URI
SUBST := '$$ACR_NAME $$DOMAIN $$KEYVAULT_CERT_URI $$IMAGE_TAG'

export ACR_NAME DOMAIN KEYVAULT_CERT_URI IMAGE_TAG

.PHONY: all acr-login attach-acr approuting images secrets render deploy rollout status undeploy

all: images deploy rollout

acr-login: require-ACR_NAME
	az acr login --name $(ACR_NAME)

attach-acr: require-ACR_NAME require-RESOURCE_GROUP require-AKS_NAME
	az aks update --resource-group $(RESOURCE_GROUP) --name $(AKS_NAME) --attach-acr $(ACR_NAME)

approuting: require-RESOURCE_GROUP require-AKS_NAME
	az aks approuting enable --resource-group $(RESOURCE_GROUP) --name $(AKS_NAME)

require-%:
	@test -n "$($*)" || { echo "$* is required" >&2; exit 1; }

images: acr-login
	@for svc in $(IMAGES); do docker buildx build --platform $(PLATFORM) -t $(REGISTRY)/$$svc:$(IMAGE_TAG) -f $(DOCKER_DIR)/$$svc/Dockerfile $(REPO_ROOT)/$$svc --push || exit 1; done

SECRETS_FILE := overlay/secrets/snippetbox.env

$(SECRETS_FILE):
	@mkdir -p $(dir $@)
	@umask 077 && printf 'MYSQL_ROOT_PASSWORD=%s\nUSER_DB_PASSWORD=%s\nSNIPPET_DB_PASSWORD=%s\nJWT_SECRET=%s\n' "$$(openssl rand -hex 24)" "$$(openssl rand -hex 24)" "$$(openssl rand -hex 24)" "$$(openssl rand -hex 32)" > $@

secrets: $(SECRETS_FILE)

render: secrets $(REQUIRED)
	@$(KUBECTL) kustomize overlay | envsubst '$(SUBST)'

deploy: secrets $(REQUIRED)
	$(KUBECTL) kustomize overlay | envsubst '$(SUBST)' | $(KUBECTL) apply -f -

rollout:
	$(KUBECTL) -n $(NAMESPACE) rollout status statefulset/snippetbox-mysql --timeout=300s
	$(KUBECTL) -n $(NAMESPACE) rollout status deployment/snippetbox-backend-user --timeout=300s
	$(KUBECTL) -n $(NAMESPACE) rollout status deployment/snippetbox-backend-snippet --timeout=300s
	$(KUBECTL) -n $(NAMESPACE) rollout status deployment/snippetbox-frontend --timeout=300s

status:
	$(KUBECTL) -n $(NAMESPACE) get pods,svc,ingress,hpa,pdb,pvc

undeploy:
	$(KUBECTL) kustomize overlay | envsubst '$(SUBST)' | $(KUBECTL) delete --ignore-not-found -f -
