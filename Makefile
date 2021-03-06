# Go parameters
    GOCMD=go
    GOBUILD=$(GOCMD) build
    GOCLEAN=$(GOCMD) clean
    MODCLEAN=$(GOCMD) clean -modcache
    GOTEST=$(GOCMD) test
    GOGET=$(GOCMD) get
    GOFMT=$(GOCMD) fmt ./...
    BINARY_NAME=epp
    BINARY_UNIX=$(BINARY_NAME)_unix
    NOW=`date +'%Y-%m-%d_%T'`
    DATE=`date +'%Y-%m-%d'`
    VERSIONGIT=`git rev-parse HEAD`
    VERSION=`git describe --tags`
    #REGISTRY need to be set in ENV to push lo local docker registry

    all: test build

    docker-build: fmt
		docker build --build-arg version=$(VERSIONGIT) --build-arg buildtime=$(NOW) . -t $(BINARY_NAME)

    docker-dist:
		rm -rf dist/*.xz
		rm -rf dist/*.tar
		docker save $(BINARY_NAME):latest > dist/$(BINARY_NAME)-$(DATE).tar
		xz -f -9 dist/$(BINARY_NAME)-$(DATE).tar

    build: fmt
		$(GOBUILD) -o $(BINARY_NAME) -trimpath -ldflags "-X main.sha1ver=$(VERSIONGIT) -X main.buildTime=$(NOW)" -v

    docker-push:
		docker tag $(BINARY_NAME) $(REGISTRY)/$(BINARY_NAME)
		docker tag $(BINARY_NAME) $(REGISTRY)/$(BINARY_NAME):$(VERSION)
		docker push $(REGISTRY)/$(BINARY_NAME)
		docker push $(REGISTRY)/$(BINARY_NAME):$(VERSION)

    docker: docker-build docker-push

    test:
		$(GOTEST) -v ./...

    clean:
		$(GOCLEAN)
		$(MODCLEAN)
		rm -f dist/$(BINARY_NAME)
		rm -f dist/$(BINARY_NAME).tar.xz

    run:
		$(GOBUILD) -o build/ -trimpath -ldflags "-X main.sha1ver=$(VERSIONGIT) -X main.buildTime=$(NOW)" -v ./...
		./build/$(BINARY_NAME)


    # Cross compilation
    dist:
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME) -ldflags "-X main.sha1ver=$(VERSIONGIT) -X main.buildTime=$(NOW)" -v


	# Format sources
    fmt:
		$(GOFMT)


# Update dependencies
    update:
		go get -u ./...