.PHONY: frontend backend build run

frontend:
	cd frontend && npm install && npm run build
	rm -rf backend/internal/web/dist
	cp -r frontend/dist backend/internal/web/dist

backend: frontend
	cd backend && go build -o ../dvr-manager ./cmd/server

build: backend

run: build
	./dvr-manager
