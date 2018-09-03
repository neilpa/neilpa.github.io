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

# Publish new/updated content to the site
publish:
	@echo todo: push new content to the site w/out restarting

# Running the site locally
local: build
	./neilpa.me -local

# Deploy the release artifact
deploy: default
	GOOS=linux go build -o neilpa.me-linux
	scp neilpa.me-linux neilpa.me:~/
	scp static/ neilpa.me:~/static/
	ssh neilpa.me "pkill neilpa.me || true && mv neilpa.me-linux neilpa.me"
	#ssh neilpa.me "nohup ./neilpa.me > /dev/null 2>&1 &"
	@echo todo: figure out how to restart cleanly, manual for now

# Remove generated artifacts
clean:
	rm -f ./neilpa.me*

# Running curl against a bunch of local endpoints
local-curl:
	curl -i localhost:8080/
	curl -i localhost:8080/health
	curl -i localhost:8080/status
	curl -i localhost:8080/version
	curl -i localhost:8080/404
	curl -i localhost:8080/favicon.ico

# Generating the key and self signed cert
# TODO: Look at the cert generation code in the go stdlib
local-cert:
	openssl req \
      -x509 \
      -nodes \
      -newkey rsa:2048 \
      -keyout local.key \
      -out local.crt \
      -days 3650 \
      -subj "/C=US/ST=Washington/L=Seattle/O=Global Security/OU=IT Department/CN=*"

