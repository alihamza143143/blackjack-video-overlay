1. 构建和运行本地RNG服务:
```bash
# 单独运行本地RNG服务
go run main.go -mode rng -port 50000

# 或者使用Docker
docker build -t roulette-rng -f Dockerfile.rng .
docker run -p 50000:50000 roulette-rng
```
2. 构建和运行轮盘服务:
```bash
# 单独运行轮盘服务(带RNG)
go run main.go -mode roulette -port 6000 -rng localhost:50000

# 单独运行轮盘服务(不带RNG，使用本地随机数)
go run main.go -mode roulette -port 6000

# 或者使用Docker
docker build -t roulette .
docker run -p 6000:6000 roulette
# 或者使用Docker Compose
docker-compose up --build
```
3. 运行测试:
```bash
go test -v ./test
```
4. proto生成Go代码:
```bash
protoc --go_out=. --go-grpc_out=. proto/gameLogic.proto proto/roulette.proto
protoc --go_out=. --go-grpc_out=. proto/rng.proto
```
5. 统计rtp:
```bash
# 使用本地rng
go run main.go -mode rtp -count 1000000000
# 使用线上rng
go run main.go -mode rtp -rng localhost:50000 -count 1000000000
```