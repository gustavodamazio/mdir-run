# Makefile do mdir-run
# Compatível com GNU Make 3.81 (padrão do macOS).
# Linux/Windows são compilados via fyne-cross (requer Docker em execução).

APP           ?= mdir-run
APP_ID        ?= com.gustavodamazio.mdir-run
BIN_DIR       ?= bin
DIST_DIR      ?= dist
FYNE_CROSS_DIR ?= fyne-cross
FYNE_DIST     ?= $(FYNE_CROSS_DIR)/dist

DARWIN_ARCHS  ?= amd64 arm64
LINUX_ARCHS   ?= amd64 arm64
WINDOWS_ARCHS ?= amd64 arm64

# As imagens do fyne-cross trazem uma versão de Go mais antiga que a do go.mod
# e usam GOTOOLCHAIN=local. Com "auto", o Go do container baixa a versão exigida
# pelo go.mod (fica no cache do fyne-cross, então baixa só uma vez).
FYNE_CROSS_FLAGS ?= -env GOTOOLCHAIN=auto

# Flags extras para `make run` / `make run-cli`, ex.: make run-cli ARGS='-dir . -commands "ls"'
ARGS ?=

DARWIN_TARGETS  := $(addprefix build-darwin-,$(DARWIN_ARCHS))
LINUX_TARGETS   := $(addprefix build-linux-,$(LINUX_ARCHS))
WINDOWS_TARGETS := $(addprefix build-windows-,$(WINDOWS_ARCHS))

.DEFAULT_GOAL := help

.PHONY: help run run-cli build-local build build-darwin build-linux build-windows \
	$(DARWIN_TARGETS) $(LINUX_TARGETS) $(WINDOWS_TARGETS) \
	package release install deps tidy fmt vet test check clean

help: ## Lista os alvos disponíveis
	@echo "Uso: make <alvo> [VAR=valor]"
	@echo ""
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_%-]+:.*## / {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "  Alvos por arquitetura:"
	@echo "    $(DARWIN_TARGETS)"
	@echo "    $(LINUX_TARGETS)"
	@echo "    $(WINDOWS_TARGETS)"

## ---------- Desenvolvimento ----------

run: ## Executa o app em modo GUI (go run .)
	go run . $(ARGS)

run-cli: ## Executa o app em modo CLI (go run . -cli)
	go run . -cli $(ARGS)

deps: ## Baixa as dependências (go mod download)
	go mod download

tidy: ## Organiza go.mod/go.sum (go mod tidy)
	go mod tidy

fmt: ## Formata o código (go fmt)
	go fmt ./...

vet: ## Análise estática (go vet)
	go vet ./...

test: ## Roda os testes (go test)
	go test ./...

check: fmt vet test ## fmt + vet + test

install: ## Instala o binário no GOPATH/bin (go install)
	go install .

## ---------- Build ----------

build-local: ## Compila para a máquina atual em bin/mdir-run
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP) .

build: build-darwin build-linux build-windows ## Compila para todas as plataformas (equivale ao build.sh)
	@echo ""
	@echo "Build concluído. Binários em $(BIN_DIR)/:"
	@ls -la $(BIN_DIR)

build-darwin: $(DARWIN_TARGETS) ## Compila para macOS (amd64 + arm64, CGO nativo)

build-linux: $(LINUX_TARGETS) ## Compila para Linux via fyne-cross (amd64 + arm64)

build-windows: $(WINDOWS_TARGETS) ## Compila para Windows via fyne-cross (amd64 + arm64)

$(DARWIN_TARGETS): build-darwin-%:
	@mkdir -p $(BIN_DIR)
	@echo "==== Compilando darwin/$* ===="
	CGO_ENABLED=1 GOOS=darwin GOARCH=$* go build -o $(BIN_DIR)/$(APP)-darwin-$* .

$(LINUX_TARGETS): build-linux-%:
	@echo "==== Compilando linux/$* via fyne-cross ===="
	fyne-cross linux -arch=$* -app-id $(APP_ID) $(FYNE_CROSS_FLAGS)
	@set -e; \
	src="$(FYNE_DIST)/linux-$*/$(APP).tar.xz"; \
	if [ ! -f "$$src" ]; then echo "❌ Erro: $$src não encontrado"; exit 1; fi; \
	tmp="$(FYNE_CROSS_DIR)/tmp-extract-linux-$*"; \
	rm -rf "$$tmp"; mkdir -p "$$tmp" $(BIN_DIR); \
	tar -xf "$$src" -C "$$tmp"; \
	cp -f "$$tmp/usr/local/bin/$(APP)" $(BIN_DIR)/$(APP)-linux-$*; \
	rm -rf "$$tmp"; \
	echo "✅ linux/$* copiado para $(BIN_DIR)/$(APP)-linux-$*"

$(WINDOWS_TARGETS): build-windows-%:
	@echo "==== Compilando windows/$* via fyne-cross ===="
	fyne-cross windows -arch=$* -app-id $(APP_ID) $(FYNE_CROSS_FLAGS)
	@set -e; \
	src="$(FYNE_DIST)/windows-$*/$(APP).exe.zip"; \
	if [ ! -f "$$src" ]; then echo "❌ Erro: $$src não encontrado"; exit 1; fi; \
	tmp="$(FYNE_CROSS_DIR)/tmp-extract-windows-$*"; \
	rm -rf "$$tmp"; mkdir -p "$$tmp" $(BIN_DIR); \
	unzip -o -q "$$src" -d "$$tmp"; \
	cp -f "$$tmp/$(APP).exe" $(BIN_DIR)/$(APP)-windows-$*.exe; \
	rm -rf "$$tmp"; \
	echo "✅ windows/$* copiado para $(BIN_DIR)/$(APP)-windows-$*.exe"

## ---------- Release ----------

package: ## Empacota bin/mdir-run-* em dist/ (.tar.gz / .zip; equivale ao package.sh)
	@set -e; \
	mkdir -p $(DIST_DIR); \
	found=0; \
	for file in $(BIN_DIR)/$(APP)-*; do \
		[ -f "$$file" ] || continue; \
		found=1; \
		name=$$(basename "$$file"); \
		case "$$name" in \
			*.exe) zip -j -q "$(DIST_DIR)/$${name%.exe}.zip" "$$file" ;; \
			*)     tar -czf "$(DIST_DIR)/$$name.tar.gz" -C $(BIN_DIR) "$$name" ;; \
		esac; \
		echo "📦 $$name"; \
	done; \
	if [ $$found -eq 0 ]; then echo "❌ Nenhum binário $(APP)-* em $(BIN_DIR)/. Rode 'make build' antes."; exit 1; fi; \
	echo "Empacotamento concluído. Arquivos em $(DIST_DIR)/"

release: clean build package ## clean + build + package

clean: ## Remove bin/, dist/ e fyne-cross/
	rm -rf $(BIN_DIR) $(DIST_DIR) $(FYNE_CROSS_DIR)
