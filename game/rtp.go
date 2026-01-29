package game

import (
	"runtime"
	"sync"

	"gitee.com/heartfun/rouletteserv/rng"
	"github.com/rs/zerolog/log"
)

func CalculateRTP(numRounds int, rngAddr string) float64 {
	var rngClient RNGClient

	// 如果有RNG服务地址，则创建RNG客户端
	if rngAddr != "" {
		client, err := rng.NewRNGClient(rngAddr)
		if err != nil {
			log.Err(err).Msg("failed to create RNG client")
		} else {
			defer client.Close()
			rngClient = client
		}
	}

	roulette := NewRoulette(rngClient)

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

	// 增加工作池设置
	maxWorkers := runtime.NumCPU() * 2
	taskChan := make(chan struct {
		bt struct {
			name    string
			numbers []int
		}
		simulations int
	}, maxWorkers*10)

	resultChan := make(chan struct {
		bet int
		win int64
	}, maxWorkers*10)

	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				// 每个任务创建独立实例
				localRoulette := NewRoulette(rngClient)

				betType, err := DetermineBetType(task.bt.numbers)
				if err != nil {
					log.Err(err).Msg("DetermineBetType() error")
					return
				}

				// 计算理论RTP
				// probability := float64(len(task.bt.numbers)) / 37.0
				// payout := float64(Payouts[betType])
				// expectedRTP := probability * (payout + 1)

				// 模拟下注
				totalBet := 0
				totalWin := int64(0)

				// 批量获取随机数（关键优化）
				batchSize := 1000
				batches := task.simulations / batchSize
				remainder := task.simulations % batchSize

				for b := 0; b < batches; b++ {
					// 一次性获取批量随机数
					rngs := make([]uint32, batchSize)
					if localRoulette.rngClient == nil {
						for i := 0; i < batchSize; i++ {
							winningNumber, _ := roulette.Spin()
							rngs[i] = uint32(winningNumber)
						}
					} else {
						nums, _ := localRoulette.rngClient.GetRandomNumbers(int32(batchSize+256), NumberCount)
						rngs = nums
					}
					for _, rng := range rngs {
						// 使用批量随机数处理下注
						totalBet += batchSize
						if CheckWin(betType, task.bt.numbers, int(rng)%NumberCount) {
							totalWin += CalculatePayout(betType, int64(1)) * int64(batchSize)
						}
					}
				}
				// 剩余获取随机数
				for i := 0; i < remainder; i++ {
					// 每次下注1单位
					betAmount := 1
					totalBet += betAmount

					// 旋转轮盘
					winningNumber, err := roulette.Spin()
					if err != nil {
						log.Err(err).Msg("Spin() error")
						continue
					}

					// 检查是否获胜
					if CheckWin(betType, task.bt.numbers, winningNumber) {
						totalWin += CalculatePayout(betType, int64(betAmount))
					}
				}

				// 添加实际RTP计算
				// actualRTP := 0.0
				// if totalBet > 0 {
				// 	actualRTP = float64(totalWin) / float64(totalBet)
				// }

				// 发送结果到通道
				resultChan <- struct {
					bet int
					win int64
				}{totalBet, totalWin}

				// log.Info().Msgf("%s: Expected RTP = %.2f%%, Actual RTP = %.2f%% (after %d spins)", task.bt.name, expectedRTP*100, actualRTP*100, task.simulations) // 现在有定义
			}
		}()
	}

	// 分发任务（每个下注类型拆分成多个任务）
	go func() {
		for _, bt := range betTypes {
			simulations := numRounds / len(betTypes)
			batchSize := 10000 // 每个任务处理10000次

			batches := simulations / batchSize
			remainder := simulations % batchSize

			for i := 0; i < batches; i++ {
				taskChan <- struct {
					bt struct {
						name    string
						numbers []int
					}
					simulations int
				}{bt, batchSize}
			}
			if remainder > 0 {
				taskChan <- struct {
					bt struct {
						name    string
						numbers []int
					}
					simulations int
				}{bt, remainder}
			}
		}
		close(taskChan)
	}()

	// 启动单独的goroutine等待结果
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 汇总结果
	sumBet := 0
	sumWin := int64(0)
	for res := range resultChan {
		sumBet += res.bet
		sumWin += res.win
	}

	// 计算整体RTP
	overallRTP := float64(sumWin) / float64(sumBet)
	log.Info().Msgf("Overall RTP = %.2f%% (after %d rounds) totalWagered=%d, totalWon=%d", overallRTP*100, numRounds, sumBet, sumWin)
	return overallRTP
}
