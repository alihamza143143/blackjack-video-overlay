package rng

import (
	"context"
	"fmt"

	"gitee.com/heartfun/rouletteserv/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// RNGClient 随机数生成客户端
type RNGClient struct {
	conn   *grpc.ClientConn
	client proto.RngClient
}

const (
	GameCode = "roulette"
)

// NewRNGClient 创建新的RNG客户端
func NewRNGClient(addr string) (*RNGClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RNG service: %v", err)
	}

	return &RNGClient{
		conn:   conn,
		client: proto.NewRngClient(conn),
	}, nil
}

// GetRandomNumber 获取随机数
func (c *RNGClient) GetRandomNumber(r int) (uint32, error) {
	ctx := context.Background()
	resp, err := c.client.GetRngs(ctx, &proto.RequestRngs{
		Nums:     0,
		Gamecode: GameCode,
	})
	if err != nil {
		log.Err(err).Msg("Failed to get random number")
		return 0, err
	}
	rn, _, err := c.ScalingRandom(resp.Rngs, r)
	return rn, err
}

// GetRandomNumbers 获取随机数
func (c *RNGClient) GetRandomNumbers(nums int32, r int) ([]uint32, error) {
	ctx := context.Background()
	resp, err := c.client.GetRngs(ctx, &proto.RequestRngs{
		Nums:     nums,
		Gamecode: GameCode,
	})
	if err != nil {
		log.Err(err).Msg("Failed to get random number")
		return []uint32{0}, err
	}
	rnArr := make([]uint32, 0, nums)
	rngs := resp.Rngs
	for i := 0; i < int(nums); i++ {
		rn, remainRngs, err := c.ScalingRandom(rngs, r)
		if err != nil {
			log.Err(err).Msg("Failed to scaling random number")
			return []uint32{0}, err
		}
		rngs = remainRngs
		rnArr = append(rnArr, rn)
	}
	return rnArr, nil
}

// ScalingRandom [0, r)
func (c *RNGClient) ScalingRandom(rngs []uint32, r int) (uint32, []uint32, error) {
	ctx := context.Background()
	curRngs := append([]uint32(nil), rngs...)
	if len(curRngs) == 0 {
		resp, err := c.client.GetRngs(ctx, &proto.RequestRngs{
			Nums:     0,
			Gamecode: GameCode,
		})
		if err != nil {
			return 0, []uint32{0}, err
		}
		curRngs = resp.Rngs
	}

	maxval := int64(r)

	cr := uint32(0)
	MAX_RANGE := int64(1) << 32
	limit := MAX_RANGE - (MAX_RANGE % maxval)

	for {
		if len(curRngs) == 0 {
			resp, err := c.client.GetRngs(ctx, &proto.RequestRngs{
				Nums:     0,
				Gamecode: GameCode,
			})
			if err != nil {
				return 0, []uint32{0}, err
			}
			curRngs = resp.Rngs
		}

		cr = curRngs[0]
		curRngs = curRngs[1:]

		if int64(cr) < limit {
			break
		}
	}

	return cr, curRngs, nil
}

// Close 关闭连接
func (c *RNGClient) Close() error {
	return c.conn.Close()
}
