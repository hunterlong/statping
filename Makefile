VERSION=$(shell cat version.txt)
COMMIT=$(shell git rev-parse HEAD)
SIGN_KEY=B76D61FAA6DB759466E83D9964B9C6AAE2D55278
BINARY_NAME=statping
GOBUILD=go build -a
GOVERSION=1.14.0
NODE_VERSION=12.18.2
XGO=xgo -go $(GOVERSION) --dest=build
BUILDVERSION=-ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT}"
TRVIS_SECRET=O3/2KTOV8krv+yZ1EB/7D1RQRe6NdpFUEJNJkMS/ollYqmz3x2mCO7yIgIJKCKguLXZxjM6CxJcjlCrvUwibL+8BBp7xJe4XFIOrjkPvbbVPry4HkFZCf2GfcUK6o4AByQ+RYqsW2F17Fp9KLQ1rL3OT3eLTwCAGKx3tlY8y+an43zkmo5dN64V6sawx26fh6XTfww590ey+ltgQTjf8UPNup2wZmGvMo9Hwvh/bYR/47bR6PlBh6vhlKWyotKf2Fz1Bevbu0zc35pee5YlsrHR+oSF+/nNd/dOij34BhtqQikUR+zQVy9yty8SlmneVwD3yOENvlF+8roeKIXb6P6eZnSMHvelhWpAFTwDXq2N3d/FIgrQtLxsAFTI3nTHvZgs6OoTd6dA0wkhuIGLxaL3FOeztCdxP5J/CQ9GUcTvifh5ArGGwYxRxQU6rTgtebJcNtXFISP9CEUR6rwRtb6ax7h6f1SbjUGAdxt+r2LbEVEk4ZlwHvdJ2DtzJHT5DQtLrqq/CTUgJ8SJFMkrJMp/pPznKhzN4qvd8oQJXygSXX/gz92MvoX0xgpNeLsUdAn+PL9KketfR+QYosBz04d8k05E+aTqGaU7FUCHPTLwlOFvLD8Gbv0zsC/PWgSLXTBlcqLEz5PHwPVHTcVzspKj/IyYimXpCSbvu1YOIjyc=
PUBLISH_BODY='{ "request": { "branch": "master", "message": "Homebrew update version v${VERSION}", "config": { "env": { "VERSION": "${VERSION}", "COMMIT": "$(TRAVIS_COMMIT)" } } } }'
TRAVIS_BUILD_CMD='{ "request": { "branch": "master", "message": "Compile master for Statping v${VERSION}", "config": { "merge_mode": "replace", "language": "go", "go": 1.14, "install": true, "sudo": "required", "services": ["docker"], "env": { "secure": "${TRVIS_SECRET}" }, "before_deploy": ["git config --local user.name \"hunterlong\"", "git config --local user.email \"info@socialeck.com\"", "git tag v$(VERSION) --force"], "deploy": [{ "provider": "releases", "api_key": "$$GITHUB_TOKEN", "file_glob": true, "file": "build/*", "skip_cleanup": true, "on": { "branch": "master" } }], "before_script": ["rm -rf ~/.nvm && git clone https://github.com/creationix/nvm.git ~/.nvm && (cd ~/.nvm && git checkout `git describe --abbrev=0 --tags`) && source ~/.nvm/nvm.sh && nvm install stable", "nvm install 10.17.0", "nvm use 10.17.0 --default", "npm install -g sass yarn cross-env", "pip install --user awscli"], "script": ["make release"], "after_success": [], "after_deploy": ["make post-release"] } } }'
TEST_DIR=$(GOPATH)/src/github.com/statping/statping
PATH:=$(GOPATH)/bin:$(PATH)
OS = freebsd linux openbsd
ARCHS = 386 arm amd64 arm64

all: build-deps compile install test build

test: clean compile
	go test -v -p=1 -ldflags="-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT}" -coverprofile=coverage.out ./...

build: clean
	CGO_ENABLED=1 go build -a -ldflags "-s -w -X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT}" -o statping --tags "netgo osusergo" ./cmd

go-build: clean
	rm -rf source/dist
	rm -rf source/rice-box.go
	wget https://assets.statping.com/source.tar.gz
	tar -xvf source.tar.gz
	go build -a -ldflags "-s -w -extldflags -static -X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT}" -o statping --tags "netgo osusergo" ./cmd

lint:
	go fmt ./...
	golint ./...
	impi --local github.com/statping/statping/ --scheme stdLocalThirdParty ./...
	goimports ./...

up:
	docker-compose -f docker-compose.yml -f dev/docker-compose.full.yml up -d --remove-orphans
	make print_details

local: clean frontend-build build
	export STATPING_DIR=./docker/statping/sqlite && ./statping --port 8080

down:
	docker-compose -f docker-compose.yml -f dev/docker-compose.full.yml down --volumes --remove-orphans

lite: clean
	docker build -t statping/statping:dev -f dev/Dockerfile.dev .
	docker-compose -f dev/docker-compose.lite.yml down
	docker-compose -f dev/docker-compose.lite.yml up --remove-orphans

reup: down clean compose-build-full up

# build all arch's and release Statping
release: test-deps
	wget -O statping.gpg $(SIGN_URL)
	gpg --import statping.gpg
	make build-all

test-ci: clean compile test-deps
	DB_CONN=sqlite go test -v -covermode=count -coverprofile=coverage.out -p=1 ./...
	goveralls -coverprofile=coverage.out -service=travis-ci -repotoken ${COVERALLS}

cypress: clean
	echo "Statping Bin: "`which statping`
	echo "Statping Version: "`statping version`
	cd frontend && yarn test
	killall statping

test-api:
	DB_CONN=sqlite DB_HOST=localhost DB_DATABASE=sqlite DB_PASS=none DB_USER=none statping &
	sleep 5000 && newman run dev/postman.json -e dev/postman_environment.json --delay-request 500

test-deps:
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls
	go get github.com/GeertJohan/go.rice/rice
	go get github.com/mattn/go-sqlite3
	go install github.com/mattn/go-sqlite3
	go install github.com/wellington/go-libsass

deps:
	go get -d -v -t ./...

protoc:
	cd types/proto && protoc --gofast_out=plugins=grpc:. statping.proto

yarn-serve:
	cd frontend && yarn serve

yarn-install:
	cd frontend && rm -rf node_modules && yarn

go-run:
	go run ./cmd

start:
	docker-compose -f docker-compose.yml -f dev/docker-compose.full.yml start

stop:
	docker-compose -f docker-compose.yml -f dev/docker-compose.full.yml stop

logs:
	docker logs statping --follow

db-up:
	docker-compose -f dev/docker-compose.db.yml up -d --remove-orphans

db-down:
	docker-compose -f dev/docker-compose.db.yml down --volumes --remove-orphans

console:
	docker exec -t -i statping /bin/sh

compose-build-full: docker-base
	docker-compose -f docker-compose.yml -f dev/docker-compose.full.yml build --parallel --build-arg VERSION=${VERSION}

docker-base:
	docker build -t statping/statping:base -f Dockerfile.base --build-arg VERSION=${VERSION} .

docker-latest: docker-base
	docker build -t statping/statping:latest --build-arg VERSION=${VERSION} .

docker-vue:
	docker build -t statping/statping:vue --build-arg VERSION=${VERSION} .

docker-test:
	docker-compose -f docker-compose.test.yml up --remove-orphans

push-base: clean compile docker-base
	docker push statping/statping:base

push-vue: clean compile docker-base docker-vue
	docker push statping/statping:base
	docker push statping/statping:vue

modd:
	modd -f ./dev/modd.conf

top:
	docker-compose -f docker-compose.yml -f dev/docker-compose.full.yml top

frontend-build:
	@echo "Removing old frontend distributions..."
	@rm -rf source/dist && rm -rf frontend/dist
	@echo "yarn install and build static frontend"
	cd frontend && yarn && yarn build
	@cp -r frontend/dist source/
	@cp -r frontend/src/assets/scss source/dist/
	@cp frontend/public/robots.txt source/dist/
	@echo "Frontend build complete at ./source/dist"

yarn:
	rm -rf source/dist && rm -rf frontend/dist
	cd frontend && yarn

# compile assets using SASS and Rice. compiles scss -> css, and run rice embed-go
compile: frontend-build
	rm -f source/rice-box.go
	cd source && rice embed-go
	make generate

embed:
	cd source && rice embed-go

install: build
	mv $(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

install-local: build
	mv $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

install-darwin:
	go build -a -ldflags "-X main.VERSION=${VERSION}" -o statping --tags "netgo darwin" ./cmd
	mv $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

generate:
	go generate ./...

build-all: clean compile build-folders build-linux build-linux-arm build-darwin build-win compress-folders

build-deps:
	apt install -y libc6-armel-cross libc6-dev-armel-cross binutils-arm-linux-gnueabi \
	libncurses5-dev build-essential bison flex libssl-dev bc gcc-arm-linux-gnueabihf g++-arm-linux-gnueabihf \
	gcc-arm-linux-gnueabi g++-arm-linux-gnueabi libsqlite3-dev gcc-mingw-w64 gcc-mingw-w64-x86-64

build-darwin:
	GO111MODULE="on" GOOS=darwin GOARCH=amd64 \
		go build -a -ldflags "-s -w -X main.VERSION=${VERSION}" -o releases/statping-darwin-amd64/statping --tags "netgo darwin" ./cmd

build-win:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GO111MODULE="on" GOOS=windows GOARCH=amd64 \
		go build -a -ldflags "-s -w -extldflags -static -X main.VERSION=${VERSION}" -o releases/statping-windows-amd64/statping.exe ./cmd
	CGO_ENABLED=1 CC=i686-w64-mingw32-gcc CXX=i686-w64-mingw32-g++ GO111MODULE="on" GOOS=windows GOARCH=386 \
		go build -a -ldflags "-s -w -extldflags -static -X main.VERSION=${VERSION}" -o releases/statping-windows-386/statping.exe ./cmd

build-linux:
	CGO_ENABLED=1 GO111MODULE="on" GOOS=linux GOARCH=amd64 \
		go build -a -ldflags "-s -w -extldflags -static -X main.VERSION=${VERSION}" -o releases/statping-linux-amd64/statping --tags "netgo linux" ./cmd
	CGO_ENABLED=1 GO111MODULE="on" GOOS=linux GOARCH=386 \
		go build -a -ldflags "-s -w -extldflags -static -X main.VERSION=${VERSION}" -o releases/statping-linux-386/statping --tags "netgo linux" ./cmd

build-linux-arm:
	CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc-6 CXX=arm-linux-gnueabihf-g++-6 GO111MODULE="on" GOOS=linux GOARCH=arm GOARM=6 \
		go build -a -ldflags "-s -w -extldflags -static -X main.VERSION=${VERSION}" -o releases/statping-linux-arm6/statping --tags "netgo" ./cmd
	CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc-6 CXX=arm-linux-gnueabihf-g++-6 GO111MODULE="on" GOOS=linux GOARCH=arm GOARM=7 \
		go build -a -ldflags "-s -w -extldflags -static -X main.VERSION=${VERSION}" -o releases/statping-linux-arm7/statping --tags "netgo" ./cmd
	CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc-6 CXX=aarch64-linux-gnu-g++-6 GO111MODULE="on" GOOS=linux GOARCH=arm64 \
		go build -a -ldflags "-s -w -extldflags -static -X main.VERSION=${VERSION}" -o releases/statping-linux-arm64/statping --tags "netgo" ./cmd

build-folders:
	mkdir build || true
	for os in windows darwin linux;\
    do \
        for arch in 386 amd64 arm6 arm7 arm64;\
        do \
            mkdir -p releases/statping-$$os-$$arch/; \
        done \
    done

compress-folders:
	mkdir build || true
	for os in darwin linux;\
    do \
        for arch in 386 amd64 arm6 arm7 arm64;\
		do \
			chmod +x releases/statping-$$os-$$arch/statping || true; \
			tar -czf releases/statping-$$os-$$arch.tar.gz -C releases/statping-$$os-$$arch statping || true; \
		done \
	done
	chmod +x releases/statping-windows-386/statping.exe || true
	chmod +x releases/statping-windows-amd64/statping.exe || true
	chmod +x releases/statping-windows-arm/statping.exe || true
	zip -j releases/statping-windows-386.zip releases/statping-windows-386/statping.exe || true; \
	zip -j releases/statping-windows-amd64.zip releases/statping-windows-amd64/statping.exe || true; \
	zip -j releases/statping-windows-arm.zip releases/statping-windows-arm/statping.exe || true; \
	find ./releases/ -name "*.tar.gz" -type f -size +1M -exec mv "{}" build/ \;
	find ./releases/ -name "*.zip" -type f -size +1M -exec mv "{}" build/ \;

# remove files for a clean compile/build
clean:
	@echo "Cleaning temporary and build folders..."
	@rm -rf ./{logs,assets,plugins,*.db,config.yml,.sass-cache,config.yml,statping,build,.sass-cache,index.html,vendor}
	@rm -rf cmd/{logs,assets,plugins,*.db,config.yml,.sass-cache,*.log,*.html,*.json}
	@rm -rf core/{logs,assets,plugins,*.db,config.yml,.sass-cache,*.log}
	@rm -rf types/notifications/{logs,assets,plugins,*.db,config.yml,.sass-cache,*.log}
	@rm -rf handlers/{logs,assets,plugins,*.db,config.yml,.sass-cache,*.log}
	@rm -rf notifiers/{logs,assets,plugins,*.db,config.yml,.sass-cache,*.log}
	@rm -rf source/{logs,assets,plugins,*.db,config.yml,.sass-cache,*.log}
	@rm -rf types/{logs,assets,plugins,*.db,config.yml,.sass-cache,*.log}
	@rm -rf utils/{logs,assets,plugins,*.db,config.yml,.sass-cache,*.log}
	@rm -rf frontend/{logs,plugins,*.db,config.yml,.sass-cache,*.log}
	@rm -rf dev/{logs,assets,plugins,*.db,config.yml,.sass-cache,*.log,test/app,plugin/*.so}
	@rm -rf frontend/cypress/videos
	@rm -f coverage.* sass
	@rm -rf **/*.db-journal
	@find . -name "*.out" -type f -delete
	@find . -name "*.cpu" -type f -delete
	@find . -name "*.mem" -type f -delete
	@rm -rf {build,releases,tmp,source/build,snap,parts,prime,snap,stage}
	@echo "Finished removing temporary and build folders"

print_details:
	@echo \==== Statping Development Instance ====
	@echo \Statping Vue Frontend:     http://localhost:8888
	@echo \Statping Backend API:      http://localhost:8585
	@echo \==== Statping Instances ====
	@echo \Statping SQLite:     http://localhost:4000
	@echo \Statping MySQL:      http://localhost:4005
	@echo \Statping Postgres:   http://localhost:4010
	@echo \==== Databases ====
	@echo \PHPMyAdmin:          http://localhost:6000  \(MySQL database management\)
	@echo \SQLite Web:          http://localhost:6050  \(SQLite database management\)
	@echo \PGAdmin:             http://localhost:7000  \(Postgres database management \| email: admin@admin.com password: admin\)
	@echo \Prometheus:          http://localhost:7050  \(Prometheus Web UI\)
	@echo \==== Monitoring and IDE ====
	@echo \Grafana:             http://localhost:3000  \(username: admin, password: admin\)

coverage: test-deps
	$(GOPATH)/bin/goveralls -coverprofile=coverage.out -service=travis -repotoken $(COVERALLS)

# build Statping using a travis ci trigger
travis-build:
	curl -s -X POST -H "Content-Type: application/json" -H "Accept: application/json" -H "Travis-API-Version: 3" -H "Authorization: token $(TRAVIS_API)" -d $(TRAVIS_BUILD_CMD) https://api.travis-ci.com/repo/statping%2Fstatping/requests

download-key:
	wget -O statping.gpg $(SIGN_URL)
	gpg --import statping.gpg

# push the :dev docker tag using curl
dockerhub-dev:
	docker build --build-arg VERSION=${VERSION} -t statping/statping:dev --no-cache -f Dockerfile.base .
	docker push statping/statping:dev

dockerhub:
	docker build --build-arg VERSION=${VERSION} -t statping/statping:base --no-cache -f Dockerfile.base .
	docker build --build-arg VERSION=${VERSION} -t statping/statping:latest --no-cache -f Dockerfile .
	docker tag statping/statping statping/statping:v${VERSION}
	docker push statping/statping:base
	docker push statping/statping:v${VERSION}
	docker push statping/statping

docker-build-dev:
	docker build --build-arg VERSION=${VERSION} -t statping/statping:latest --no-cache -f Dockerfile .
	docker tag statping/statping:latest statping/statping:dev-v${VERSION}

post-release: frontend-build upload_to_s3 publish-homebrew dockerhub

# update the homebrew application to latest for mac
publish-homebrew:
	curl -s -X POST -H "Content-Type: application/json" -H "Accept: application/json" -H "Travis-API-Version: 3" -H "Authorization: token $(TRAVIS_API)" -d $(PUBLISH_BODY) https://api.travis-ci.com/repo/statping%2Fhomebrew-statping/requests

upload_to_s3:
	tar -czvf source.tar.gz source/
	aws s3 cp source.tar.gz s3://assets.statping.com/
	rm -rf source.tar.gz
	aws s3 cp source/dist/css/ s3://assets.statping.com/css/ --recursive --exclude "*" --include "*.css"
	aws s3 cp source/dist/js/ s3://assets.statping.com/js/ --recursive --exclude "*" --include "*.js"
	aws s3 cp source/dist/scss/ s3://assets.statping.com/scss/ --recursive --exclude "*" --include "*.scss"
	aws s3 cp install.sh s3://assets.statping.com/

travis_s3_creds:
	mkdir -p ~/.aws
	echo "[default]\naws_access_key_id = ${AWS_ACCESS_KEY_ID}\naws_secret_access_key = ${AWS_SECRET_ACCESS_KEY}" > ~/.aws/credentials

sign-all:
	gpg --default-key $SIGN_KEY --detach-sign --armor statpinger

valid-sign:
	gpg --verify statping.asc

sentry-release:
	sentry-cli releases --org statping --project backend new v${VERSION}
	sentry-cli releases --org statping --project backend set-commits v${VERSION} --auto
	sentry-cli releases --org statping --project backend finalize v${VERSION}
	sentry-cli releases --org statping --project frontend new v${VERSION}
	sentry-cli releases --org statping --project frontend set-commits v${VERSION} --auto
	sentry-cli releases --org statping --project frontend finalize v${VERSION}

download-bins: clean
	mkdir build || true
	wget "https://github.com/statping/statping/releases/download/v${VERSION}/statping-linux-386.tar.gz"
	wget "https://github.com/statping/statping/releases/download/v${VERSION}/statping-linux-amd64.tar.gz"
	wget "https://github.com/statping/statping/releases/download/v${VERSION}/statping-linux-arm.tar.gz"
	wget "https://github.com/statping/statping/releases/download/v${VERSION}/statping-linux-arm64.tar.gz"
	mv statping-linux-386.tar.gz build/
	mv statping-linux-amd64.tar.gz build/
	mv statping-linux-arm.tar.gz build/
	mv statping-linux-arm64.tar.gz build/

snapcraft: download-bins
	mkdir snap
	mv snapcraft.yaml snap/
	PWD=$(shell pwd)
	snapcraft clean statping
	docker run --rm -v ${PWD}/build/statping-linux-amd64.tar.gz:/build/statping-linux.tar.gz -w /build --env VERSION=${VERSION} snapcore/snapcraft bash -c "apt update && snapcraft --target-arch=amd64"
	snapcraft clean statping
	docker run --rm -v ${PWD}/build/statping-linux-386.tar.gz:/build/statping-linux.tar.gz -w /build --env VERSION=${VERSION} snapcore/snapcraft bash -c "apt update && snapcraft --target-arch=i386"
	snapcraft clean statping
	docker run --rm -v ${PWD}/build/statping-linux-arm64.tar.gz:/build/statping-linux.tar.gz -w /build --env VERSION=${VERSION} snapcore/snapcraft bash -c "apt update && snapcraft --target-arch=arm64"
	snapcraft clean statping
	docker run --rm -v ${PWD}/build/statping-linux-arm.tar.gz:/build/statping-linux.tar.gz -w /build --env VERSION=${VERSION} snapcore/snapcraft bash -c "apt update && snapcraft --target-arch=arm"
	snapcraft push statping_${VERSION}_amd64.snap --release stable
	snapcraft push statping_${VERSION}_arm64.snap --release stable
	snapcraft push statping_${VERSION}_i386.snap --release stable
	snapcraft push statping_${VERSION}_arm.snap --release stable

postman: clean compile
	API_SECRET=demosecret123 statping --port=8080 > /dev/null &
	sleep 3
	newman run -e dev/postman_environment.json dev/postman.json
	killall statping

certs:
	openssl req -newkey rsa:2048 \
	  -new -nodes -x509 \
	  -days 3650 \
	  -out cert.pem \
	  -keyout key.pem \
	  -subj "/C=US/ST=California/L=Santa Monica/O=Statping/OU=Development/CN=localhost"

xgo-latest:
	xgo --go $(GOVERSION) --targets=linux/amd64,linux/386,linux/arm-7,linux/arm-6,linux/arm64,windows/386,windows/amd64,darwin/386,darwin/amd64 --out='statping' --pkg='cmd' --dest=build --tags 'netgo' --ldflags='-X main.VERSION=${VERSION} -X main.COMMIT=$(COMMIT) -linkmode external -extldflags "-static"' .

buildx-latest: multiarch
	docker buildx create --name statping-latest
	docker buildx inspect --builder statping-latest --bootstrap
	docker buildx build --builder statping-latest --cache-from "type=local,src=/tmp/.buildx-cache" --cache-to "type=local,dest=/tmp/.buildx-cache" --pull --push --platform linux/amd64,linux/arm64,linux/arm/v7,linux/arm/v6 -f Dockerfile -t statping/statping:latest -t statping/statping:v${VERSION} --build-arg=VERSION=${VERSION} --build-arg=COMMIT=${COMMIT} .
	docker buildx rm statping-latest

buildx-dev: multiarch
	docker buildx create --name statping-dev
	docker buildx inspect --builder statping-dev --bootstrap
	docker buildx build --builder statping-dev --cache-from "type=local,src=/tmp/.buildx-cache" --cache-to "type=local,dest=/tmp/.buildx-cache" --pull --push --platform linux/amd64,linux/arm64,linux/arm/v7,linux/arm/v6 -f Dockerfile -t statping/statping:dev --build-arg=VERSION=${VERSION} --build-arg=COMMIT=${COMMIT} .
	docker buildx rm statping-dev

buildx-base: multiarch
	docker buildx create --name statping-base
	docker buildx inspect --builder statping-base --bootstrap
	docker buildx build --builder statping-base --cache-from "type=local,src=/tmp/.buildx-cache" --cache-to "type=local,dest=/tmp/.buildx-cache" --pull --push --platform linux/amd64,linux/arm64,linux/arm/v7,linux/arm/v6 -f Dockerfile.base -t statping/statping:base --build-arg=VERSION=${VERSION} --build-arg=COMMIT=${COMMIT} .
	docker buildx rm statping-base

multiarch:
	mkdir /tmp/.buildx-cache || true
	docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

delve:
	go build -gcflags "all=-N -l" -o statping ./cmd
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./statping

check:
	@echo "Checking the programs required for the build are installed..."
	@echo "go:     $(shell go version) - $(shell which go)" && go version >/dev/null 2>&1 || (echo "ERROR: go 1.14 is required."; exit 1)
	@echo "node:   $(shell node --version) - $(shell which node)" && node --version >/dev/null 2>&1 || (echo "ERROR: node 12.x is required."; exit 1)
	@echo "yarn:   $(shell yarn --version) - $(shell which yarn)" && yarn --version >/dev/null 2>&1 || (echo "ERROR: yarn is required."; exit 1)
	@echo "All required programs are installed!"

#sentry-release:
#	sentry-cli releases new -p $SENTRY_PROJECT $VERSION
#	sentry-cli releases set-commits --auto $VERSION
#	sentry-cli releases files $VERSION upload-sourcemaps dist

gen_help:
	for file in ./statping.wiki/*.md
	  do
		# convert each file to html and place it in the html directory
		# --gfm == use github flavoured markdown
		marked -o html/$file.html $file --gfm
	done

.PHONY: all check build certs multiarch install-darwin go-build build-all buildx-base buildx-dev buildx-latest build-alpine test-all test test-api docker frontend up down print_details lite sentry-release snapcraft build-linux build-mac build-win build-all postman
.SILENT: travis_s3_creds
