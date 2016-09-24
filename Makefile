GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
DOCKER_PREFIX=sudo
IMAGE_NAME=unixvoid/cryodns
GIT_HASH=$(shell git rev-parse HEAD | head -c 10)

cryodns:
	$(GOC) cryodns.go

dependencies:
	go get github.com/gorilla/mux
	go get gopkg.in/gcfg.v1
	go get github.com/unixvoid/glogger
	go get github.com/miekg/dns
	go get gopkg.in/redis.v4
	go get golang.org/x/crypto/sha3

run:
	go run \
		cryodns/cryodns.go \
		cryodns/api_listener.go \
		cryodns/api_list_dns.go \
		cryodns/api_add_dns.go \
		cryodns/api_remove_dns.go \
		cryodns/bootstrap.go \
		cryodns/register.go \
		cryodns/rotate.go

docker:
	$(MAKE) stat
	mkdir stage.tmp/
	cp bin/cryodns* stage.tmp/cryodns
	cp deps/Dockerfile stage.tmp/
	cp deps/run.sh stage.tmp/
	cp deps/rootfs.tar.gz stage.tmp/
	cp config.gcfg stage.tmp/
	sed -i "s/<GIT_HASH>/$(GIT_HASH)/g" stage.tmp/Dockerfile
	cd stage.tmp/ && \
		$(DOCKER_PREFIX) docker build -t $(IMAGE_NAME) .

dockerrun:
	$(DOCKER_PREFIX) docker run \
		-d \
		-p 9053:9053 \
		-p 9080:9080 \
		-v /tmp/:/redisbackup:rw \
		--name cryodns \
		$(IMAGE_NAME)

aci:
	$(MAKE) stat
	mkdir -p stage.tmp/cryodns-layout/rootfs/
	tar -zxf deps/rootfs.tar.gz -C stage.tmp/cryodns-layout/rootfs/
	cp bin/cryodns* stage.tmp/cryodns-layout/rootfs/cryodns
	chmod +x deps/run.sh
	cp deps/run.sh stage.tmp/cryodns-layout/rootfs/
	sed -i "s/\$$DIFF/$(GIT_HASH)/g" stage.tmp/cryodns-layout/rootfs/run.sh
	cp config.gcfg stage.tmp/cryodns-layout/rootfs/
	cp deps/manifest.json stage.tmp/cryodns-layout/manifest
	cd stage.tmp/ && \
		actool build cryodns-layout cryodns.aci && \
		mv cryodns.aci ../
	@echo "cryodns.aci built"

testaci:
	deps/testrkt.sh

travisaci:
	wget https://github.com/appc/spec/releases/download/v0.8.7/appc-v0.8.7.tar.gz
	tar -zxf appc-v0.8.7.tar.gz
	$(MAKE) stat
	mkdir -p stage.tmp/cryodns-layout/rootfs/
	tar -zxf deps/rootfs.tar.gz -C stage.tmp/cryodns-layout/rootfs/
	cp bin/cryodns* stage.tmp/cryodns-layout/rootfs/cryodns
	chmod +x deps/run.sh
	cp deps/run.sh stage.tmp/cryodns-layout/rootfs/
	sed -i "s/\$$DIFF/$(GIT_HASH)/g" stage.tmp/cryodns-layout/rootfs/run.sh
	cp config.gcfg stage.tmp/cryodns-layout/rootfs/
	cp deps/manifest.json stage.tmp/cryodns-layout/manifest
	cd stage.tmp/ && \
		../appc-v0.8.7/actool build cryodns-layout cryodns.aci && \
		mv cryodns.aci ../
	@echo "cryodns.aci built"

clean:
	rm -rf bin/
	rm -rf stage.tmp/

stat:
	mkdir -p bin/
	$(CGOR) $(GOC) $(GOFLAGS) -o bin/cryodns-$(GIT_HASH)-linux-amd64 cryodns/*.go
