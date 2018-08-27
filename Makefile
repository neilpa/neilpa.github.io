# Build, analyze and test the site by default
default: build lint test

# Build the site
build:
	go build

# Run go vet (and other potential linters)
lint:
	go vet

# Run the test suite
test:
	go test -v

# Running the site locally
local: build
	./neilpa.me

# Deploy the release artifact
deploy: default
	@echo deploy: TODO: push artifact to github
	@echo deploy: TODO: ensure your running as proper user

# Remove generated artifacts
clean:
	rm -f ./neilpa.me

# Generating the key and self signed cert
local-cert:
	openssl req \
      -x509 \
      -nodes \
      -newkey rsa:2048 \
      -keyout local.key \
      -out local.crt \
      -days 3650 \
      -subj "/C=US/ST=Washington/L=Seattle/O=Global Security/OU=IT Department/CN=*"

# Running curl against a bunch of local endpoints
local-curl:
	curl -i localhost:8080/
	curl -i localhost:8080/health
	curl -i localhost:8080/status
	curl -i localhost:8080/version
	curl -i localhost:8080/404
	curl -i localhost:8080/favicon.ico

