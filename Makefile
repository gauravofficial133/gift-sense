# GiftSense — Developer Makefile
# Usage: make <target>
# Requires: Go 1.22+, Node 18+, npm, Git Bash (Windows)

SHELL := bash
ENV_FILE := .env

# Load .env into make environment (ignore missing file)
-include $(ENV_FILE)
export

# ── Colours ───────────────────────────────────────────────────────────────────
CYAN  := \033[0;36m
RESET := \033[0m

.PHONY: help backend frontend dev test lint build clean

# ── Default target ────────────────────────────────────────────────────────────
help:
	@echo ""
	@echo "  $(CYAN)upahaar.ai — available targets$(RESET)"
	@echo ""
	@echo "  make dev         Start backend + frontend together (two terminals)"
	@echo "  make backend     Start Go backend only  (localhost:8080)"
	@echo "  make frontend    Start Vite frontend only (localhost:5173)"
	@echo ""
	@echo "  make test        Run all backend Go tests"
	@echo "  make lint        Run go vet on the backend"
	@echo "  make build       Build backend binary + frontend production bundle"
	@echo ""
	@echo "  make install     Install frontend npm dependencies"
	@echo "  make clean       Remove build artefacts"
	@echo ""

# ── Run targets ───────────────────────────────────────────────────────────────

# Start backend and frontend in parallel (requires a terminal that shows both)
dev:
	@echo "$(CYAN)Starting backend and frontend...$(RESET)"
	@$(MAKE) -j2 backend frontend

backend:
	@echo "$(CYAN)Starting backend on :$(PORT)$(RESET)"
	cd giftsense-backend && go run ./cmd/server/

frontend:
	@echo "$(CYAN)Starting frontend on :5173$(RESET)"
	cd giftsense-frontend && npm run dev

# ── Quality targets ───────────────────────────────────────────────────────────

test:
	@echo "$(CYAN)Running backend tests...$(RESET)"
	cd giftsense-backend && go test ./... -v

lint:
	@echo "$(CYAN)Running go vet...$(RESET)"
	cd giftsense-backend && go vet ./...

# ── Build targets ─────────────────────────────────────────────────────────────

build: build-backend build-frontend

build-backend:
	@echo "$(CYAN)Building backend binary...$(RESET)"
	cd giftsense-backend && go build -o ../bin/giftsense-backend ./cmd/server/

build-frontend:
	@echo "$(CYAN)Building frontend bundle...$(RESET)"
	cd giftsense-frontend && npm run build

# ── Setup targets ─────────────────────────────────────────────────────────────

install:
	@echo "$(CYAN)Installing frontend dependencies...$(RESET)"
	cd giftsense-frontend && npm install

# ── Cleanup ───────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/ giftsense-frontend/dist/
	@echo "$(CYAN)Cleaned build artefacts.$(RESET)"
