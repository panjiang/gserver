CMD_LIST := queue

# 初次配置
install:
	go version
	go install github.com/golang/protobuf/protoc-gen-go
	docker network create gserver

# 编译proto文件
protoc:
	protoc --go_out=paths=source_relative:. api/**/*.proto

# 创建docker镜像
# make image CMD=queue
image:
	@set -e;
	for app in $(CMD_LIST); do \
		set -e; \
		CMDDIR=./cmd/$$app; \
		IMAGE=gserver/$$app; \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $$CMDDIR $$CMDDIR; \
		cp config.yml $$CMDDIR/config.yml; \
		docker build -t $$IMAGE $$CMDDIR; \
		rm -f $$CMDDIR/$$app; \
		rm $$CMDDIR/config.yml; \
	done

# 启动基础服务容器
up-lib:
	docker-compose -f docker/docker-compose.lib.yml -p gserver up -d

down-lib:
	docker-compose -f docker/docker-compose.lib.yml down

# 启动自建服务容器
up:
	docker-compose -f docker/docker-compose.cmd.yml -p gserver up -d

down:
	docker-compose -f docker/docker-compose.cmd.yml down

# 重新编译并启动容器
run-cli: CMD_LIST=queue
run: image
	docker-compose -f docker/docker-compose.cmd.yml -p gserver up

# 运行客户端
run-cli: CMD_LIST=queue-cli
run-cli: image
	docker run --rm --net=host --name=gserver-queue-cli gserver/queue-cli ./queue-cli -s=localhost:8080 -n=10000

# 运行客户端
run-cli1: CMD_LIST=queue-cli
run-cli1: image
	docker run --rm --net=host --name=gserver-queue-cli gserver/queue-cli ./queue-cli -s=wuyabiji.com:8080 -n=5000

# 镜像发布到远程仓库
image-tag:
	@set -e;
	docker login -u=docker  docker.panjiang.xyz

	for app in $(CMD_LIST); do \
		image=gserver/$$app; \
		docker tag $$image docker.panjiang.xyz/$$image; \
		docker push docker.panjiang.xyz/$$image; \
	done

test:
	go run tests/test.go