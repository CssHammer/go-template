include *.mk
SHELL:=/bin/bash
REGISTRY=docker.ginsp.net
GIN_DIRECTORY:=${HOME}/.gin
TAGGER_DIRECTORY:=${GIN_DIRECTORY}/tagger
TAGGER_BINARY:=${TAGGER_DIRECTORY}/tagger.py
TAGGER_REMOTE:=ssh://git@git.syneforge.com:7999/gin/tagger.git

OS :=$(shell uname -s)

.PHONY: check-werf
check-werf:
	@which werf > /dev/null 2>&1 || (echo You\'re missing werf executable; @exit 1)

.PHONY: check-kubectl
check-kubectl:
	@which kubectl > /dev/null 2>&1 || (echo You\'re missing kubectl executable; @exit 1)

.PHONY: check-python
check-python:
	@which python > /dev/null 2>&1 || (echo You\'re missing python executable; @exit 1)

.PHONY: fetch-tagger
fetch-tagger:
ifeq "$(wildcard ${TAGGER_BINARY})" ""
	@echo Tagger binary not found. Trying to download from GIN repository
ifneq (,$(wildcard ${TAGGER_DIRECTORY}))
	@echo Directory is not empty. Cleaning up
	@rm -rf ${TAGGER_DIRECTORY}
endif
	@git clone ${TAGGER_REMOTE} ${TAGGER_DIRECTORY} > /dev/null 2>&1
	@chmod +x ${TAGGER_BINARY}
endif

.PHONY: update-tagger
update-tagger:
	@rm -rf ${TAGGER_DIRECTORY}
	@git clone ${TAGGER_REMOTE} ${TAGGER_DIRECTORY} > /dev/null 2>&1
	@chmod +x ${TAGGER_BINARY}

current-tag:
ifeq "${RELEASE}" ""
	$(eval RELEASE := $(shell ${TAGGER_BINARY} --current))
endif
	@if [ "${RELEASE}" == "" ]; then echo Tagger returned non-zero exit code; exit 1; fi
	@echo Got version ${RELEASE}

.PHONY: next-tag
next-tag:
ifeq "${RELEASE}" ""
	$(eval RELEASE := $(shell ${TAGGER_BINARY}))
endif
	@if [ "${RELEASE}" == "" ]; then echo Tagger returned non-zero exit code; exit 1; fi
	@echo Preparing to deploy ${RELEASE}

.PHONY: build-werf
build-werf:
	OS=${OS} werf build ${SERVICE} --stages-storage :local

.PHONY: push-werf
push-werf:
	OS=${OS} werf publish ${SERVICE} --stages-storage :local --tag-git-tag ${RELEASE} --images-repo ${REGISTRY}

.PHONY: build
build: check next-tag build-werf push-werf

.PHONY: check
check: check-werf check-kubectl check-python fetch-tagger

define k8s-deploy
	kubectl config use-context ${1} && \
	helm3 upgrade ${SERVICE} .helm \
	  --install --timeout 10m --wait \
	  --kube-context=${1} \
	  --namespace=${2} \
	  --set "global.namespace=${2}" \
	  --set "global.image=${REGISTRY}/${SERVICE}:${RELEASE}" \
	  --values ".helm/${3}" \
	  --values .helm/dec-secrets.yaml || (rm -f .helm/dec-*.yaml && false)
	rm -f .helm/dec-*.yaml
endef

.PHONY: kube-stage-deploy
kube-stage-deploy:
	@echo Deploying release ${RELEASE} to the Google K8s Engine
	sops -d .helm/secret-values-stage.yaml > .helm/dec-secrets.yaml
	$(call k8s-deploy,${STAGE_KUBE_CONTEXT},${STAGE_NAMESPACE},values-stage.yaml)

.PHONY: deploy-stage
deploy-stage: current-tag kube-stage-deploy

.PHONY: stage
stage: build deploy-stage

.PHONY: kube-production-deploy
kube-production-deploy:
	@echo Deploying release ${RELEASE} to the Google K8s Engine
	sops -d .helm/secret-values-prod.yaml > .helm/dec-secrets.yaml
	$(call k8s-deploy,${PRODUCTION_KUBE_CONTEXT},${PRODUCTION_NAMESPACE},values-prod.yaml)

.PHONY: deploy-production
deploy-production: current-tag kube-production-deploy

.PHONY: production
production: build deploy-production

.PHONY: secrets-stage
secrets-stage:
	sops --encrypt --gcp-kms "${STAGE_GCP_KMS_KEY}" --input-type yaml --output-type yaml ./.helm/_secret-values.yaml > ./.helm/secret-values-stage.yaml

.PHONY: secrets-prod
secrets-prod:
	sops --encrypt --gcp-kms "${PRODUCTION_GCP_KMS_KEY}" --input-type yaml --output-type yaml ./.helm/_secret-values.yaml > ./.helm/secret-values-prod.yaml
