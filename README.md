# gserver

以 QueueService 为例的 Go 游戏服务器基础框架

## 框架目录结构

模块已充分拆分解耦

```
api # 通信接口
  - queue # 服务目录
    - queue.proto # RPC接口定义
cmd # 启动程序
  - queue # 具体服务
    - codes # 协议号
    - dao # 数据库操作
    - handler # 逻辑处理
      - process # 协议处理方法管理中心
      - token # token相关的处理方法
    - hub # 连接分发
    - model # 数据模型
    - Dockerfile # Docker编译配置
  - queue-cli # 服务客户端
docker # 容器相关配置文件
  - docker-compose.yml # 发布版本
  - docker-compose.cmd.yml # 本地测试版本，单独启动程序
  - docker-compose.lib.yml # 本地测试版本，单独启动基础服务
pkg # 公共方法和类
  - utils # 工具包
config.yml # 服务配置文件
Makefile # 命令封装
```

## 通信协议

`[CODE:2B][BODY_LEN:4B][BODY:nB]`

- CODE - 协议号 `uint16`
- BODY_LEN - 包体长度 `uint32`
- BODY - Protobuff 压缩后的实例数据

## QueueService 技术实现

用 Redis 的 Sorted Set 结构作为请求队列，客户端唯一 ID 作为 Member，精简的微秒时间戳作为 Score `score = float64(当前微秒时间戳 - 2020年开始时微秒时间戳)`

有独立的协程负责分批颁发 token，位置在 `cmd/queue/hub/token/token.go`

**设计原因**

- 查询其中一个元素效率远高于链表
- 方便以 Rank 为分隔批量获取请求、批量颁发令牌、批量删除请求
- 用 Redis 的 `TxPipeline` 命令批量操作，提升效率

## 本地使用

1. 安装 golang
2. 安装 protoc
3. 安装 docker、docker-compose
4. 执行`make install`：安装 protoc-gen-go，docker 创建 gserver 网络
5. 执行`make up-lib`：启动 redis 容器服务
6. 执行`make run`：编译 queue 镜像并启动服务
7. 执行`make run-cli`：启动测试客户端

## 远程发布

1. 执行`make image-tag`，推送镜像到远程仓库
2. 远程机器创建`/docker/gserver`目录，放入`docker/docker-compose.yml`容器文件
3. 创建`/docker/gserver/queue`目录，放入`config.yml`服务配置文件
4. 启动服务`docker-compose up -d`

## 测试数据

> 测试前请确保：
>
> 1. 系统放开了打开文件数量的限制 `ulimit -n`
> 2. 优化了内核参数`net.ipv4.tcp_tw_reuse = 1`

5000 个客户端，Dial 时间间隔为 1ms

```
go run cmd/queue-cli/main.go -s=localhost:8080 -n=5000
```

服务器每秒限制颁发 1000 个 token，颁发逻辑在 10ms 以内

```
queue_1  | 6:12AM DBG issue tokens count=858 dur=5.8419 limit=1000
queue_1  | 6:12AM DBG issue tokens count=920 dur=6.0305 limit=1000
queue_1  | 6:12AM DBG issue tokens count=923 dur=6.1036 limit=1000
queue_1  | 6:12AM DBG issue tokens count=921 dur=9.0467 limit=1000
queue_1  | 6:12AM DBG issue tokens count=920 dur=7.4056 limit=1000
queue_1  | 6:12AM DBG issue tokens count=458 dur=2.7069 limit=1000
```

服务器从接收请求到返回时间，在 0.3ms 左右

```
queue_1  | 6:16AM DBG  code=100 duration=0.3093 user=4034
queue_1  | 6:16AM DBG  code=100 duration=0.3452 user=4035
queue_1  | 6:16AM DBG  code=100 duration=0.3121 user=4036
queue_1  | 6:16AM DBG  code=100 duration=0.3303 user=4037
queue_1  | 6:16AM DBG  code=100 duration=0.3365 user=4038
queue_1  | 6:16AM DBG  code=100 duration=0.3607 user=4039
queue_1  | 6:16AM DBG  code=100 duration=0.3693 user=4040
queue_1  | 6:16AM DBG  code=100 duration=0.3326 user=4041
queue_1  | 6:16AM DBG  code=100 duration=0.3256 user=4042
```

服务器 CPU 高峰：47%，内存 0.2%

```
  PID USER      PR  NI    VIRT    RES    SHR S  %CPU %MEM     TIME+ COMMAND
13625 root      20   0 1454032  31200   5916 S  47.0  0.2   0:01.56 main
```
