## BUILDS
HOST_PLATFORM=$$(uname -s)
HOST_ARCH=$$(uname -m)
BIN_NAME=$$(cat README.md  | grep -o '\[[a-z-]\+]' | tr -d '[]')
BUILD_FILES=main.go canvas.go renderer.go mgo-helper.go s3-uploader.go cloudflare-helper.go
## DOCKER AND TAGS
NAME=$$BIN_NAME
TAG=gcr.io/$$(gcloud config get-value project)/$(NAME)
TAG_VERSION=$$(cat README.md | grep -o '\[v[0-9]\.[0-9]\.[0-9]\]' | tr -d '[v]')

prepare:
	@echo "Verifying UPX Binery"
	@echo "HOST : " $(HOST_PLATFORM)
	@if [ -e "/usr/local/bin/upx" ]; \
	  then echo "UPX Already Installed"; \
	else \
	  echo "Installing UPX for Platform: " && \
	    if [ $(HOST_PLATFORM) == "Darwin" ]; \
	      then echo "MAC" && \
	      brew reinstall upx; \
	    else \
		   echo "Linux" && \
	      wget https://github.com/upx/upx/releases/download/v3.94/upx-3.94-amd64_linux.tar.xz	-O /tmp/upx-3.94-amd64_linux.tar.xz && \
	      tar xvf /tmp/upx-3.94-amd64_linux.tar.xz -C /tmp/ upx-3.94-amd64_linux/upx && \
	    mv /tmp/upx-3.94-amd64_linux/upx /usr/local/bin/ ; \
	    fi \
	fi

	@echo "Verifying GoDep Binary"
	@if [ -e "/usr/local/bin/dep" ]; \
	  then echo "GoDep Already Installed"; \
	else \
	  echo "Installing GoDep for Platform: " && \
	    if [ $(HOST_PLATFORM) == "Darwin" ]; \
	      then echo "MAC" && \
	      brew reinstall dep; \
	    else \
		   echo "Linux" && \
	      curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh; \
	    fi \
	fi

	@echo "Preparing GO dependencies ..."
	@if [ -e "./Gopkg.toml" ]; \
	  then echo "Skipping dep init"; \
	else \
	  dep init; \
	fi
	dep ensure
	
dev:
	go run $(BUILD_FILES)

production:
	./$(BIN_NAME)

production-osx:
	./$(BIN_NAME)-osx

build-osx:
	env GOARCH=amd64 CGO_ENABLED=0 COMPRESS_BINARY=true -ldflags="-s -w" go build -o $(BIN_NAME)-osx $(BUILD_FILES)
	@upx --lzma -9 $(BIN_NAME)-osx

##  https://blog.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/
build: prepare
	# if you want fast build remove --ultra-brute
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 COMPRESS_BINARY=true -ldflags="-s -w" go build -o $(BIN_NAME) $(BUILD_FILES)
	@upx --ultra-brute --lzma -9 $(BIN_NAME)	

# build-docker: build
# 	docker rmi $(TAG):latest -f
# 	docker build . -t $(TAG):latest
# 	docker tag $(TAG):latest $(TAG):$(TAG_VERSION)
	
# push-docker:
# 	gcloud docker -- push $(TAG):latest
# 	gcloud docker -- push $(TAG):$(TAG_VERSION)
