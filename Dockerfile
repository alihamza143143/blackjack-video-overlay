# 构建阶段
FROM golang:1.24.1 as builder

WORKDIR /app/rouletteserv
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o roulette .

# 运行阶段
FROM alpine:3.21.3

WORKDIR /app/rouletteserv
COPY --from=builder /app/rouletteserv/roulette .

#审核需要
COPY README.md ./
COPY CHANGELOG.md ./
COPY genhash.sh ./
RUN chmod u+x genhash.sh
COPY runrtpindocker.sh ./
RUN chmod u+x runrtpindocker.sh
COPY VERSION ./
RUN sh genhash.sh

# 设置默认环境变量
ENV PORT=6000
ENV RNG=""

EXPOSE $PORT

CMD ["./roulette", "-mode", "roulette", "-port", "6000", "-rng", ""]