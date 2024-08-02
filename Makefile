# type 'make help' to visualize the help information

##@ Utility
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ CI/CD
build: ## Build the application to the output folder (default: ./buil/main)
	@echo "Building..."	
	@go build -race -o build/main cmd/api/main.go

docker-build: ## Build a container image and add the version and latest tag
	@if command -v docker > /dev/null; then \
		echo "Building..."; \
		docker buildx build -t jsfelipearaujo/ms-product-catalog:latest -t jsfelipearaujo/ms-product-catalog:$$(git rev-parse --short HEAD) .; \
	else \
		read -p "Docker Buildx is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			brew install docker-buildx; \
			echo "Building..."; \
			docker buildx build -t jsfelipearaujo/ms-product-catalog:latest -t jsfelipearaujo/ms-product-catalog:$$(git rev-parse --short HEAD) .; \
		else \
			echo "You chose not to install Docker Buildx. Exiting..."; \
			exit 1; \
		fi; \
	fi

docker-push: ## Push the container image to the registry
	echo "Pushing..."; \
	docker push jsfelipearaujo/ms-product-catalog:latest && \
	docker push jsfelipearaujo/ms-product-catalog:$$(git rev-parse --short HEAD)

clean: ## Clean the binary
	@echo "Cleaning..."
	@rm -f build/main

sec: ## Security checker with gosec
	@if command -v gosec > /dev/null; then \
		echo "Analyzing..."; \
		gosec ./...; \
	else \
		read -p "gosec is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/securego/gosec/v2/cmd/gosec@latest; \
			echo "Analyzing..."; \
			gosec ./...; \
		else \
			echo "You chose not to intall gosec. Exiting..."; \
			exit 1; \
		fi; \
	fi

scan: ## Scan for code vulnerabilities with govulncheck
	@if command -v govulncheck > /dev/null; then \
		echo "Scanning..."; \
		govulncheck -test -format text -show color ./...; \
	else \
		read -p "Go 'govulncheck' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install golang.org/x/vuln/cmd/govulncheck@latest; \
			echo "Scanning..."; \
			govulncheck -test -format text -show color ./...; \
		else \
			echo "You chose not to intall govulncheck. Exiting..."; \
			exit 1; \
		fi; \
	fi

k8s-attach: ## Attach to the application running in Kubernetes
	aws eks update-kubeconfig --name fastfood --region us-east-1

k8s-copy-config: ## Copy the application configuration to Kubernetes
	cat /home/jfelipearaujo/.kube/config > k8s/kubeconfig

k8s-deploy: ## Deploy the application to Kubernetes
	@if [ "$(id)" != "" ]; then \
		echo "Generating sensitive data..."; \
		cat k8s/service-account.yaml | sed "s/{{AWS_ACCOUNT_ID}}/$(id)/g" > k8s/service-account-sensitive.yaml; \
		echo "Deploying..."; \
		kubectl apply -f k8s/namespace.yaml; \
		kubectl apply -f k8s/configmap.yaml; \
		kubectl apply -f k8s/service-account-sensitive.yaml; \
		kubectl apply -f k8s/secret.yaml; \
		kubectl apply -f k8s/deployment.yaml; \
		kubectl apply -f k8s/service.yaml; \
		kubectl apply -f k8s/hpa.yaml; \
		kubectl apply -f k8s/ingres.yaml; \
		rm k8s/service-account-sensitive.yaml; \
		echo "Deployed!"; \
	else \
		read -p "please, inform the AWS Account ID to be used: " id; \
		if [ "$$id" != "" ]; then \
			echo "Generating sensitive data..."; \
			cat k8s/service-account.yaml | sed "s/{{AWS_ACCOUNT_ID}}/$$id/g" > k8s/service-account-sensitive.yaml; \
			echo "Deploying..."; \
			kubectl apply -f k8s/namespace.yaml; \
			kubectl apply -f k8s/configmap.yaml; \
			kubectl apply -f k8s/service-account-sensitive.yaml; \
			kubectl apply -f k8s/secret.yaml; \
			kubectl apply -f k8s/deployment.yaml; \
			kubectl apply -f k8s/service.yaml; \
			kubectl apply -f k8s/hpa.yaml; \
			kubectl apply -f k8s/ingres.yaml; \
			rm k8s/service-account-sensitive.yaml; \
			echo "Deployed!"; \
		else \
			echo "You must inform the AWS Account ID to be used. Exiting..."; \
			exit 1; \
		fi; \
	fi

k8s-destroy: ## Destroy the application from Kubernetes
	@if [ "$(id)" != "" ]; then \
		echo "Destroying..."; \
		kubectl delete -f k8s/ingres.yaml; \
		kubectl delete -f k8s/hpa.yaml; \
		kubectl delete -f k8s/service.yaml; \
		kubectl delete -f k8s/deployment.yaml; \
		kubectl delete -f k8s/secret.yaml; \
		kubectl delete -f k8s/service-account-sensitive.yaml; \
		kubectl delete -f k8s/configmap.yaml; \
		kubectl delete -f k8s/namespace.yaml; \
		rm k8s/service-account-sensitive.yaml; \
		echo "Destroyed!"; \
	else \
		read -p "please, inform the AWS Account ID to be used: " id; \
		if [ "$$id" != "" ]; then \
			echo "Generating sensitive data..."; \
			cat k8s/service-account.yaml | sed "s/{{AWS_ACCOUNT_ID}}/$$id/g" > k8s/service-account-sensitive.yaml; \
			echo "Destroying..."; \
			kubectl delete -f k8s/ingres.yaml; \
			kubectl delete -f k8s/hpa.yaml; \
			kubectl delete -f k8s/service.yaml; \
			kubectl delete -f k8s/deployment.yaml; \
			kubectl delete -f k8s/secret.yaml; \
			kubectl delete -f k8s/service-account-sensitive.yaml; \
			kubectl delete -f k8s/configmap.yaml; \
			kubectl delete -f k8s/namespace.yaml; \
			rm k8s/service-account-sensitive.yaml; \
			echo "Destroyed!"; \
		else \
			echo "You must inform the AWS Account ID to be used. Exiting..."; \
			exit 1; \
		fi; \
	fi

lint: ## Go Linter
	@if command -v golangci-lint > /dev/null; then \
		echo "Analyzing..."; \
		golangci-lint run ./...; \
	else \
		read -p "golangci-lint is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.2; \
			echo "Analyzing..."; \
			golangci-lint run ./...; \
		else \
			echo "You chose not to intall golangci-lint. Exiting..."; \
			exit 1; \
		fi; \
	fi

##@ Runner
run: ## Run the application
	@if test ! -f .env; then \
		make env; \
	fi
	@go run cmd/api/main.go local;

##@ Testing
test: ## Test the application
	@echo "Testing..."
	@go test -race -count=1 ./internal/... -coverprofile=coverage.out

test-bdd: ## Run BDD tests
	@echo "Running BDD tests..."
	@go test -count=1 ./tests/... -test.v -test.run ^TestFeatures$

cover: ## View the coverage
	@echo "Analyzing coverage..."
	@go tool cover -func=coverage.out

test-cover: ## Run the tests and view the coverage
	make test && make cover

bench: ## Run benchmarks
	@echo "Benchmarking..."
	@go test -cpu=1,2,4,6,8,16 -benchmem -bench=. ./internal/features/transactions/handlers/create_transaction -run=^Benchmark

load: ## Run the load test using k6
	@if command -v k6 > /dev/null; then \
		echo "Generating..."; \
		k6 run tests/index.js; \
	else \
		read -p "k6 is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			brew install k6; \
			echo "Generating..."; \
			k6 run tests/index.js; \
		else \
			echo "You chose not to intall k6. Exiting..."; \
			exit 1; \
		fi; \
	fi

##@ Developing
install-go: ## Install a desired version of GoLang
	@desiredVersion="go1.22.5"; \
	currentVersion=$$(go version | grep -oE 'go[0-9]+\.[0-9]+\.[0-9]+'); \
	if [ "$$currentVersion" != "$$desiredVersion" ]; then \
		cat /etc/os-release; \
		uname -r; \
		wget --quiet https://go.dev/dl/$$desiredVersion.linux-amd64.tar.gz; \
		sudo rm -rf /usr/local/go && tar -C /usr/local -xzf $$desiredVersion.linux-amd64.tar.gz; \
		export PATH=$HOME/go/bin:/usr/local/go/bin:$PATH; \
		rm -rf $$desiredVersion.linux-amd64.tar.gz; \
		go version; \
	else \
		echo "GoLang version matchs the desired version: $$desiredVersion"; \
	fi

env: ## Create the .env file based on example
	@echo "Generating..."
	@cp .env.example .env

docker-up: ## Run the containers
	@if command -v docker compose > /dev/null; then \
		docker compose up -d; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up -d; \
	fi

docker-down: ## Shutdown containers
	@if command -v docker compose > /dev/null; then \
		docker compose down; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

watch: ## Live reload using air
	@if test ! -f .env; then \
		make env; \
	fi

	@if command -v air > /dev/null; then \
		air; \
		echo "Watching...";\
	else \
		read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/cosmtrek/air@latest; \
			air; \
			echo "Watching...";\
		else \
			echo "You chose not to install air. Exiting..."; \
			exit 1; \
		fi; \
	fi

TAG := $(shell git describe --tags --abbrev=0 2>/dev/null)
VERSION := $(shell echo $(TAG) | sed 's/v//')

tag: ## Create or bump the version tag
	@if [ -z "$(TAG)" ]; then \
        echo "No previous version found. Creating v1.0 tag..."; \
        git tag v1.0; \
    else \
        echo "Previous version found: $(VERSION)"; \
        read -p "Bump major version (M/m) or release version (R/r)? " choice; \
        if [ "$$choice" = "M" ] || [ "$$choice" = "m" ]; then \
            echo "Bumping major version..."; \
			major=$$(echo $(VERSION) | cut -d'.' -f1); \
            major=$$(expr $$major + 1); \
            new_version=$$major.0; \
        elif [ "$$choice" = "R" ] || [ "$$choice" = "r" ]; then \
            echo "Bumping release version..."; \
			release=$$(echo $(VERSION) | cut -d'.' -f2); \
            release=$$(expr $$release + 1); \
            new_version=$$(echo $(VERSION) | cut -d'.' -f1).$$release; \
        else \
            echo "Invalid choice. Aborting."; \
            exit 1; \
        fi; \
        echo "Creating tag for version v$$new_version..."; \
        git tag v$$new_version; \
    fi

##@ Auto generated files
gen-mocks: ## Gen mock files using mockery
	@if command -v mockery > /dev/null; then \
		echo "Generating..."; \
		mockery; \
	else \
		read -p "Go 'mockery' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/vektra/mockery/v2@latest; \
			echo "Generating..."; \
			mockery; \
		else \
			echo "You chose not to intall mockery. Exiting..."; \
			exit 1; \
		fi; \
	fi

gen-docs: ## Gen Swagger docs using swag
	@if command -v swag > /dev/null; then \
		swag init --parseDependency -g main.go -d cmd/api,internal; \
	else \
		read -p "Go 'swag' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/swaggo/swag/cmd/swag@latest; \
			swag init --parseDependency -g main.go -d cmd/api,internal; \
		else \
			echo "You chose not to intall swag. Exiting..."; \
			exit 1; \
		fi; \
	fi

gen-pkg-docs: ## Gen Package docs using gomarkdoc
	@if command -v gomarkdoc > /dev/null; then \
		gomarkdoc --output '{{.Dir}}/README.md' ./...; \
	else \
		read -p "Go 'gomarkdoc' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest; \
			gomarkdoc --output '{{.Dir}}/README.md' ./...; \
		else \
			echo "You chose not to intall gomarkdoc. Exiting..."; \
			exit 1; \
		fi; \
	fi

fmt-docs: ## Format generated Swagger docs using swag
	@if command -v swag > /dev/null; then \
		swag fmt -d internal -g cmd/api/main.go; \
	else \
		read -p "Go 'swag' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/swaggo/swag/cmd/swag@latest; \
			swag fmt -d internal -g cmd/api/main.go; \
		else \
			echo "You chose not to intall swag. Exiting..."; \
			exit 1; \
		fi; \
	fi

gen-scaffold-bdd: ## Gen BDD scaffold using godog
	@if command -v godog > /dev/null; then \
		echo "Generating BDD scaffold..."; \
		godog ./tests/features; \
	else \
		read -p "Go 'godog' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/cucumber/godog/cmd/godog@latest; \
			echo "Generating BDD scaffold..."; \
			godog ./tests/features; \
		else \
			echo "You chose not to intall godog. Exiting..."; \
			exit 1; \
		fi; \
	fi

.PHONY: build run test clean
