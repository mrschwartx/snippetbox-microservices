NAMESPACE ?= snippetbox
PLATFORM ?= linux/amd64
REPO_ROOT ?= $(abspath $(CURDIR)/..)
DOCKER_DIR := $(REPO_ROOT)/snippetbox-docker
IMAGES := snippetbox-backend-user snippetbox-backend-snippet snippetbox-frontend
KUBECTL := kubectl
ifndef IMAGE_TAG
IMAGE_TAG := $(shell git rev-parse --short HEAD 2>/dev/null || date +%Y%m%d%H%M%S)
endif
ifndef AWS_REGION
AWS_REGION := $(shell aws configure get region 2>/dev/null)
endif
ifndef AWS_ACCOUNT_ID
AWS_ACCOUNT_ID := $(shell aws sts get-caller-identity --query Account --output text 2>/dev/null)
endif
DOMAIN ?=
ACM_CERTIFICATE_ARN ?=
REGISTRY := $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
REQUIRED := require-AWS_REGION require-AWS_ACCOUNT_ID require-DOMAIN require-ACM_CERTIFICATE_ARN
SUBST := '$$AWS_ACCOUNT_ID $$AWS_REGION $$DOMAIN $$ACM_CERTIFICATE_ARN $$IMAGE_TAG'

export AWS_ACCOUNT_ID AWS_REGION DOMAIN ACM_CERTIFICATE_ARN IMAGE_TAG

.PHONY: all ecr-login ecr-repos images secrets render deploy rollout status undeploy

all: images deploy rollout

ecr-login: require-AWS_REGION require-AWS_ACCOUNT_ID
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(REGISTRY)

ecr-repos: require-AWS_REGION
	@for svc in $(IMAGES); do aws ecr describe-repositories --repository-names $$svc --region $(AWS_REGION) >/dev/null 2>&1 || aws ecr create-repository --repository-name $$svc --image-scanning-configuration scanOnPush=true --encryption-configuration encryptionType=AES256 --region $(AWS_REGION) >/dev/null || exit 1; done

require-%:
	@test -n "$($*)" || { echo "$* is required" >&2; exit 1; }

images: ecr-login ecr-repos
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
