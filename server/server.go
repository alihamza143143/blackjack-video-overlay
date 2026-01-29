

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"gitee.com/heartfun/rouletteserv/game"
	"gitee.com/heartfun/rouletteserv/proto"
	"gitee.com/heartfun/rouletteserv/rng"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
)

// TestServiceServer implements the echo test service
type TestServiceServer struct {
    proto.UnimplementedTestServiceServer
}

// TestBackend echoes the input message
func (s *TestServiceServer) TestBackend(ctx context.Context, req *proto.TestBackendRequest) (*proto.TestBackendReply, error) {
    return &proto.TestBackendReply{Reply: req.Message}, nil
}

// RouletteServer 轮盘服务
type RouletteServer struct {
	proto.UnimplementedGameLogicServer
	game *game.Roulette
}

// NewRouletteServer 创建新的轮盘服务
func NewRouletteServer(rngClient game.RNGClient) *RouletteServer {
	return &RouletteServer{
		game: game.NewRoulette(rngClient),
	}
}

// Play2 处理下注请求
func (s *RouletteServer) Play2(ctx context.Context, req *proto.RequestPlay) (*proto.ReplyPlay, error) {
	// 旋转轮盘获取获胜数字
	winningNumber, err := s.game.Spin()
	if err != nil {
		log.Err(err).Msg("failed to spin roulette")
		return nil, fmt.Errorf("failed to spin roulette")
	}

	if req.Cheat != "" {
		// 解析作弊数据
		parts := strings.Split(req.Cheat, ",")
		num, err := strconv.Atoi(parts[0])
		if err == nil {
			winningNumber = num % game.NumberCount
			log.Debug().Any("cheatnum", num).Msg(req.Command)
		} else {
			log.Err(err).Msg("invalid cheat data")
		}
	}

	result := &proto.ReplyPlay{
		RandomNumbers: []*proto.RngInfo{{Value: int32(winningNumber + (winningNumber+1024)*game.NumberCount), Bits: 0, Range: 0}},
		PlayerState: &proto.PlayerState{
			Public:  nil,
			Private: nil,
		},
		Finished:          true,
		Results:           make([]*proto.GameResult, 0, 1),
		NextCommands:      nil,
		NextCommandParams: nil,
	}

	var breq proto.BetRequest
	// 解析嵌套 JSON 数据到结构体
	err = json.Unmarshal([]byte(req.ClientParams), &breq)
	if err != nil {
		log.Err(err).Msg("failed to unmarshal nested JSON data")
		return nil, fmt.Errorf("invaild bet request")
	}

	curGameModParam := &proto.GameModParam{
		WinningNumber: int32(winningNumber),
		Wins:          make([]*proto.BetWin, 0, len(breq.Bets)),
		TotalWin:      0,
	}

	// 处理每个下注
	for _, bet := range breq.Bets {
		// 转换下注数字
		numbers := make([]int, len(bet.Numbers))
		for i, n := range bet.Numbers {
			numbers[i] = int(n)
		}

		// 判断下注类型
		betType, err := game.DetermineBetType(numbers)
		if err != nil {
			log.Err(err).Msg("invalid bet")
			return nil, fmt.Errorf("invalid bet")
		}

		// 检查是否获胜
		win := game.CheckWin(betType, numbers, winningNumber)
		winAmount := int64(0)
		if win {
			winAmount = game.CalculatePayout(betType, bet.Amount)
		}

		// 添加结果
		curGameModParam.Wins = append(curGameModParam.Wins, &proto.BetWin{
			Bet: &proto.Bet{
				Numbers: bet.Numbers,
				Amount:  bet.Amount,
			},
			BetType:   string(betType),
			Win:       win,
			WinAmount: winAmount,
			Payout:    int32(game.Payouts[betType]),
		})

		curGameModParam.TotalWin += winAmount
	}

	anyMsg, err := anypb.New(curGameModParam)
	if err != nil {
		log.Err(err).Msg("failed to marshal nested JSON data")
		return nil, fmt.Errorf("invaild mod param")
	}

	result.Results = append(result.Results, &proto.GameResult{
		CoinWin: curGameModParam.TotalWin,
		CashWin: curGameModParam.TotalWin,
		ClientData: &proto.PlayResult{
			CurGameMod:      "bg",
			CurGameModParam: anyMsg,
		},
	})

	return result, nil
}

// GetConfig 获取配置
func (s *RouletteServer) GetConfig(ctx context.Context, req *proto.RequestConfig) (*proto.GameConfig, error) {
	result := &proto.GameConfig{
		Ver:          game.Version,
		CoreVer:      game.Version,
		DefaultScene: &proto.GameScene{},
		Data:         "",
	}

	return result, nil
}

// Initialize 初始化
func (s *RouletteServer) Initialize(ctx context.Context, req *proto.RequestInitialize) (*proto.PlayerState, error) {
	result := &proto.PlayerState{
		Public:  nil,
		Private: nil,
	}

	return result, nil
}

// StartServer 启动GRPC服务
func StartServer(port string, rngAddr string) error {
	var rngClient game.RNGClient

	// 如果有RNG服务地址，则创建RNG客户端
	if rngAddr != "" {
		client, err := rng.NewRNGClient(rngAddr)
		if err != nil {
			return fmt.Errorf("failed to create RNG client: %v", err)
		}
		defer client.Close()
		rngClient = client
	}

	// 创建并启动服务
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterGameLogicServer(grpcServer, NewRouletteServer(rngClient))
	proto.RegisterTestServiceServer(grpcServer, &TestServiceServer{})

	log.Info().Str("rngAddr", rngAddr).Msg("Starting roulette server on port " + port)
	return grpcServer.Serve(lis)
}
