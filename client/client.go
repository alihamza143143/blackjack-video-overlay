package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitee.com/heartfun/rouletteserv/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func main() {
	// 连接GRPC服务
	conn, err := grpc.Dial("localhost:6000", grpc.WithInsecure())
	if err != nil {
		log.Err(err).Msg("Failed to connect")
	}
	defer conn.Close()

	client := proto.NewGameLogicClient(conn)

	// 定义一些下注
	bets := []*proto.Bet{
		{Numbers: []int32{1}, Amount: 10},                                                                // 直接下注1
		{Numbers: []int32{1, 2}, Amount: 5},                                                              // 分注1-2
		{Numbers: []int32{1, 2, 3}, Amount: 3},                                                           // 街注1-2-3
		{Numbers: []int32{1, 2, 4, 5}, Amount: 2},                                                        // 角注1-2-4-5
		{Numbers: []int32{1, 4, 7, 10, 13, 16, 19, 22, 25, 28, 31, 34}, Amount: 1},                       // 第一列
		{Numbers: []int32{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 33, 35}, Amount: 1}, // 奇数
	}

	// 创建下注请求
	jsonData, err := json.Marshal(bets)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// 将字节切片转换为字符串
	jsonStr := string(jsonData)
	req := &proto.RequestPlay{ClientParams: jsonStr}

	// 发送下注请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Play2(ctx, req)
	if err != nil {
		log.Err(err).Msg("Play2 failed")
	}

	unpackedAny := resp.Results[0].ClientData.CurGameModParam
	var curGameModParam proto.GameModParam
	// 解包 anypb.Any 类型的数据
	err = unpackedAny.UnmarshalTo(&curGameModParam)
	if err != nil {
		fmt.Println("Failed to unpack Any message:", err)
		return
	}
	// 打印结果
	fmt.Printf("Winning number: %d\n", curGameModParam.WinningNumber)
	fmt.Printf("Total win: %d\n", resp.Results[0].CashWin)
	fmt.Println("Bet results:")
	for _, win := range curGameModParam.Wins {
		fmt.Printf("  Bet on %v (type: %s, amount: %d) - ", win.Bet.Numbers, win.BetType, win.Bet.Amount)
		if win.Win {
			fmt.Printf("WON %d (payout: %dx)\n", win.WinAmount, win.Payout)
		} else {
			fmt.Println("LOST")
		}
	}
}
