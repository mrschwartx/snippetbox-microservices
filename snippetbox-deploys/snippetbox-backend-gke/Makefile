NAMESPACE ?= snippetbox
PLATFORM ?= linux/amd64
REPO_ROOT ?= $(abspath $(CURDIR)/..)
DOCKER_DIR := $(REPO_ROOT)/snippetbox-docker
IMAGES := snippetbox-msa-backend-user snippetbox-msa-backend-snippet snippetbox-msa-frontend
KUBECTL := kubectl
ifndef IMAGE_TAG
IMAGE_TAG := $(shell git rev-parse --short HEAD 2>/dev/null || date +%Y%m%d%H%M%S)
endif
ifndef GCP_PROJECT
GCP_PROJECT := $(shell gcloud config get-value project 2>/dev/null)
endif
ifndef GCP_REGION
GCP_REGION := $(shell gcloud config get-value compute/region 2>/dev/null)
endif
AR_REPOSITORY ?= snippetbox
STATIC_IP_NAME ?= snippetbox-ip
DOMAIN ?=
REGISTRY := $(GCP_REGION)-docker.pkg.dev/$(GCP_PROJECT)/$(AR_REPOSITORY)
REQUIRED := require-GCP_PROJECT require-GCP_REGION require-DOMAIN
SUBST := '$$GCP_PROJECT $$GCP_REGION $$AR_REPOSITORY $$STATIC_IP_NAME $$DOMAIN $$IMAGE_TAG'

export GCP_PROJECT GCP_REGION AR_REPOSITORY STATIC_IP_NAME DOMAIN IMAGE_TAG

.PHONY: all registry docker-auth static-ip images secrets render deploy rollout status undeploy

all: images deploy rollout

registry: require-GCP_PROJECT require-GCP_REGION
	gcloud artifacts repositories describe $(AR_REPOSITORY) --location=$(GCP_REGION) --project=$(GCP_PROJECT) >/dev/null 2>&1 || gcloud artifacts repositories create $(AR_REPOSITORY) --repository-format=docker --location=$(GCP_REGION) --project=$(GCP_PROJECT)

docker-auth: require-GCP_REGION
	gcloud auth configure-docker $(GCP_REGION)-docker.pkg.dev --quiet

static-ip: require-GCP_PROJECT
	gcloud compute addresses describe $(STATIC_IP_NAME) --global --project=$(GCP_PROJECT) >/dev/null 2>&1 || gcloud compute addresses create $(STATIC_IP_NAME) --global --project=$(GCP_PROJECT)
	gcloud compute addresses describe $(STATIC_IP_NAME) --global --project=$(GCP_PROJECT) --format='value(address)'

require-%:
	@test -n "$($*)" || { echo "$* is required" >&2; exit 1; }

images: docker-auth registry
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
	$(KUBECTL) -n $(NAMESPACE) rollout status deployment/snippetbox-msa-backend-user --timeout=300s
	$(KUBECTL) -n $(NAMESPACE) rollout status deployment/snippetbox-msa-backend-snippet --timeout=300s
	$(KUBECTL) -n $(NAMESPACE) rollout status deployment/snippetbox-msa-frontend --timeout=300s

status:
	$(KUBECTL) -n $(NAMESPACE) get pods,svc,ingress,hpa,pdb,pvc

undeploy:
	$(KUBECTL) kustomize overlay | envsubst '$(SUBST)' | $(KUBECTL) delete --ignore-not-found -f -
