# Build and test the site by default
default: build test

# Build the site
build:
	go build

# Run the test suite
test:
	go test -v

# Running the site locally
local: build
	./neilpa.me

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

# Create a release artifact
release:
	@echo release: TODO: create release artifact

# Deploy the release artifact
deploy:
	@echo deploy: TODO: push artifact to github

# Remove generated artifacts
clean:
	rm -f ./neilpa.me
