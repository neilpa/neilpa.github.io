# Generate the site by default
default: www

static:
	mkdir -p www
	cp -pR static/* www/

# Run the generator and copy over static files
blog: static
	go run main.go

www: blog
	go build -o www/run serve.go

# Remove generated artifacts
clean:
	rm -rf ./www

.PHONY: clean static
