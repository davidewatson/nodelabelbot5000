GOCMD=go
GOBUILD=$(GOCMD) build
GOGET=$(GOCMD) get
GOTEST=$(GOCMD) test
GOINSTALL=$(GOCMD) install
DEPCMD=dep

# Variables for binary version information
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=$(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Associating version information variables with data
VERSION_PACKAGE=github.com/samsung-cnct/nodelabelbot5000/pkg/version
LD_DATA_GIT_COMMIT=-X ${VERSION_PACKAGE}.gitCommit=${GIT_COMMIT}
LD_DATA_GIT_TAG=-X ${VERSION_PACKAGE}.gitVersion=${GIT_TAG}
LD_DATA_GIT_DIRTY=-X ${VERSION_PACKAGE}.gitTreeState=${GIT_DIRTY}
LD_DATA_BUILD_DATE=-X ${VERSION_PACKAGE}.buildDate=${BUILD_DATE}
LD_DATA=${LD_DATA_GIT_COMMIT} ${LD_DATA_GIT_TAG} ${LD_DATA_GIT_DIRTY} ${LD_DATA_BUILD_DATE}

# Creating static binary
STATIC_FLAGS=-w -extldflags "-static"

# Generating ldflag string
LDFLAGS = '${STATIC_FLAGS} ${LD_DATA}'

# Location of
NLB5K_BINARY_NAME=nodelabelbot5000
NLB5K_SOURCE_FILE=cmd/nodelabelbot5000/main.go

# build binary command
BUILD_NLB5K_BINARY_CMD=${GOBUILD} -a -tags netgo -ldflags ${LDFLAGS} -o ${NLB5K_BINARY_NAME} ${NLB5K_SOURCE_FILE}

# By default we will most likely want to build a linux build of the application
all: new-dep nlb5k-bin-linux-amd64

container-build: install-dep dep-ensure nlb5k-bin-linux-amd64

# Clean will clean up dep and remove any binaries
clean: remove-binary

# Installs the required tools
install-tools: install-dep

# Install dep will install dep in $GOROOT/bin
install-dep:
	go get -u github.com/golang/dep/cmd/dep

# New Dep will just blow away Gopkg, vendor and do a dep init
new-dep:
	rm -rf vendor Gopkg.toml Gopkg.lock
	$(DEPCMD) init

# For builds where we're just going to go along with whatever Gopkg is saying
dep-ensure:
	$(DEPCMD) ensure

# Build a darwin (mac) binary
nlb5k-bin-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 ${BUILD_NLB5K_BINARY_CMD}

# Build a linux amd64 binary
nlb5k-bin-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ${BUILD_NLB5K_BINARY_CMD}

# Delete an existing binary / cleanup
remove-binary:
	rm ${NLB5K_BINARY_NAME}


