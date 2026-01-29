package test

import (
	"fmt"
	"math"
	"testing"

	"gitee.com/heartfun/rouletteserv/game"
)

// TestRTP 测试游戏的理论RTP
func TestGameLogic(t *testing.T) {
	// 创建没有RNG客户端的游戏实例（使用本地随机数）
	roulette := game.NewRoulette(nil)

	winningNumber, err := roulette.Spin()
	if err != nil {
		t.Fatalf("Spin() error = %v", err)
	}

	if winningNumber < 0 || winningNumber >= game.NumberCount {
		t.Errorf("Spin() returned an invalid number: %d", winningNumber)
	}
}

// TestSimulation 模拟测试实际RTP
func TestSimulation(t *testing.T) {
	// 创建没有RNG客户端的游戏实例（使用本地随机数）
	roulette := game.NewRoulette(nil)

	// 定义下注类型和模拟次数
	betTypes := []struct {
		name    string
		numbers []int
	}{
		{"Straight", []int{1}},
		{"Split", []int{1, 2}},
		{"Street", []int{1, 2, 3}},
		{"Corner", []int{1, 2, 4, 5}},
		{"Line", []int{1, 2, 3, 4, 5, 6}},
		{"Dozen", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}},
		{"Column", []int{1, 4, 7, 10, 13, 16, 19, 22, 25, 28, 31, 34}},
		{"Odd", []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 33, 35}},
		{"Red", []int{1, 3, 5, 7, 9, 12, 14, 16, 18, 19, 21, 23, 25, 27, 30, 32, 34, 36}},
		{"Low", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}},
	}

	const simulations = 1000000

	for _, bt := range betTypes {
		t.Run(bt.name, func(t *testing.T) {
			// 自动判断下注类型
			betType, err := game.DetermineBetType(bt.numbers)
			if err != nil {
				t.Fatalf("DetermineBetType() error = %v", err)
			}

			// 计算理论RTP
			probability := float64(len(bt.numbers)) / 37.0
			payout := float64(game.Payouts[betType])
			expectedRTP := probability * (payout + 1)

			// 模拟下注
			totalBet := 0
			totalWin := int64(0)

			for i := 0; i < simulations; i++ {
				// 每次下注1单位
				betAmount := 1
				totalBet += betAmount

				// 旋转轮盘
				winningNumber, err := roulette.Spin()
				if err != nil {
					t.Fatalf("Spin() error = %v", err)
				}

				// 检查是否获胜
				if game.CheckWin(betType, bt.numbers, winningNumber) {
					totalWin += game.CalculatePayout(betType, int64(betAmount))
				}
			}

			// 计算实际RTP
			actualRTP := float64(totalWin) / float64(totalBet)

			// 允许1%的误差
			if math.Abs(actualRTP-expectedRTP) > 0.01 {
				t.Errorf("%s: Expected RTP = %.4f, Actual RTP = %.4f",
					bt.name, expectedRTP, actualRTP)
			}

			fmt.Printf("%s: Expected RTP = %.2f%%, Actual RTP = %.2f%% (after %d spins)\n",
				bt.name, expectedRTP*100, actualRTP*100, simulations)
		})
	}
}
