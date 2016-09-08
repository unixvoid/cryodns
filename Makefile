GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
DOCKER_PREFIX=sudo
IMAGE_NAME=cryodns
GIT_HASH=$(shell git rev-parse HEAD | head -c 10)

cryodns:
	$(GOC) cryodns.go

run:
	go run \
		cryodns/cryodns.go \
		cryodns/api_listener.go \
		cryodns/api_list_dns.go \
		cryodns/api_add_dns.go \
		cryodns/api_remove_dns.go 

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
		-p 8053:8053 \
		-p 8080:8080 \
		-v /tmp/:/redisbackup:rw \
		cryodns

clean:
	rm -rf bin/
	rm -rf stage.tmp/

stat:
	mkdir -p bin/
	$(CGOR) $(GOC) $(GOFLAGS) -o bin/cryodns-$(GIT_HASH)-linux-amd64 cryodns/*.go
