package rng

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"time"

	"gitee.com/heartfun/rouletteserv/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// Rng RNG服务实现
type Rng struct {
	proto.UnimplementedRngServer
}

// NewRng 创建新的RNG服务
func NewRng() *Rng {
	return &Rng{}
}

func (s *Rng) GetRngs(ctx context.Context, req *proto.RequestRngs) (*proto.ReplyRngs, error) {
		Max := 36
		Min := 0
		nums := int(req.Nums)
		if nums <= 0 {
			nums = 1
		}
		rngs := make([]uint32, 0, nums)
		rd := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < nums; i++ {
			n, err := crand.Int(crand.Reader, big.NewInt(int64(Max+1)))
			if err == nil {
				rngs = append(rngs, uint32(n.Int64()))
			} else {
				log.Err(err).Msg("failed to get random number from crypto/rand, using math/rand fallback")
				num := rd.Intn(int(Max-Min+1)) + int(Min)
				rngs = append(rngs, uint32(num))
			}
		}
		return &proto.ReplyRngs{
			Rngs: rngs,
		}, nil
}

// StartServer 启动RNG服务
func StartServer(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterRngServer(grpcServer, NewRng())

	log.Info().Msg("Starting RNG server on port " + port)
	return grpcServer.Serve(lis)
}
