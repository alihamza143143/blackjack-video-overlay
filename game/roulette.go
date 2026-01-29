package game

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"
)

// 欧洲轮盘数字 (0-36)
// var wheelNumbers = []int{0, 32, 15, 19, 4, 21, 2, 25, 17, 34, 6, 27, 13, 36, 11, 30, 8, 23, 10, 5, 24, 16, 33, 1, 20, 14, 31, 9, 22, 18, 29, 7, 28, 12, 35, 3, 26}
const (
	NumberCount = 37
)

// 红色数字
var redNumbers = map[int]bool{
	1: true, 3: true, 5: true, 7: true, 9: true, 12: true, 14: true, 16: true, 18: true,
	19: true, 21: true, 23: true, 25: true, 27: true, 30: true, 32: true, 34: true, 36: true,
}

// 下注类型
type BetType string

const (
	Straight BetType = "Straight"  // 直接数字
	Split    BetType = "Split"     // 分注
	Street           = "Street"    // 街注
	Corner           = "Corner"    // 角注
	Line             = "Line"      // 线注
	Column           = "Column"    // 列注
	Dozen            = "Dozen"     // 打注
	OddEven          = "Odd/Even"  // 奇偶
	RedBlack         = "Red/Black" // 红黑
	HighLow          = "High/Low"  // 高低
	Invalid          = "Invalid"   // 无效下注
)

// 赔付倍数
var Payouts = map[BetType]int{
	Straight: 35,
	Split:    17,
	Street:   11,
	Corner:   8,
	Line:     5,
	Column:   2,
	Dozen:    2,
	OddEven:  1,
	RedBlack: 1,
	HighLow:  1,
}

// Roulette 游戏结构体
type Roulette struct {
	rngClient RNGClient
}

// RNGClient 随机数生成接口
type RNGClient interface {
	GetRandomNumber(r int) (uint32, error)
	GetRandomNumbers(nums int32, r int) ([]uint32, error)
	ScalingRandom(rngs []uint32, r int) (uint32, []uint32, error)
}

// NewRoulette 创建新的轮盘游戏实例
func NewRoulette(rngClient RNGClient) *Roulette {
	return &Roulette{
		rngClient: rngClient,
	}
}

// Spin 旋转轮盘，返回获胜数字
func (r *Roulette) Spin() (int, error) {
	if r.rngClient != nil {
		num, err := r.rngClient.GetRandomNumber(NumberCount)
		if err != nil {
			log.Err(err).Msg("failed to get random number from RNG server")
			return 0, err
		} else {
			return int(num % NumberCount), nil
		}
	}

	// 本地随机数生成
	n, err := crand.Int(crand.Reader, big.NewInt(NumberCount))
	if err == nil {
		return int(n.Int64()), nil
	} else {
		log.Err(err).Msg("failed to get random number from crypto/rand")
	}
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rd.Intn(NumberCount), nil // 0-36
}

// DetermineBetType 根据下注数字判断下注类型
func DetermineBetType(numbers []int) (BetType, error) {
	if len(numbers) == 0 {
		return Invalid, fmt.Errorf("empty bet numbers")
	}

	// 检查数字是否有效
	for _, num := range numbers {
		if num < 0 || num > 36 {
			return Invalid, fmt.Errorf("invalid number %d, must be between 0 and 36", num)
		}
	}

	switch len(numbers) {
	case 1:
		return Straight, nil
	case 2:
		if isValidSplit(numbers) {
			return Split, nil
		}
	case 3:
		if isValidStreet(numbers) {
			return Street, nil
		}
	case 4:
		if isValidCorner(numbers) {
			return Corner, nil
		}
	case 6:
		if isValidLine(numbers) {
			return Line, nil
		}
	}

	// 检查特殊下注类型
	if len(numbers) == 12 {
		if isDozen(numbers) {
			return Dozen, nil
		}
		if isColumn(numbers) {
			return Column, nil
		}
	}

	if len(numbers) == 18 {
		if isOddEven(numbers) {
			return OddEven, nil
		}
		if isRedBlack(numbers) {
			return RedBlack, nil
		}
		if isHighLow(numbers) {
			return HighLow, nil
		}
	}

	return Invalid, fmt.Errorf("invalid bet combination")
}

// 检查是否是有效的分注
func isValidSplit(numbers []int) bool {
	if len(numbers) != 2 {
		return false
	}

	n1, n2 := numbers[0], numbers[1]

	// 水平相邻
	if n1+1 == n2 && n1%3 != 0 {
		return true
	}

	// 垂直相邻
	if n1+3 == n2 {
		return true
	}

	// 特殊规则，含0也是有效的分注
	if n1 == 0 && n2 >= 1 && n2 <= 3 {
		return true
	}

	return false
}

// 检查是否是有效的街注
func isValidStreet(numbers []int) bool {
	if len(numbers) != 3 {
		return false
	}

	n1, n2, n3 := numbers[0], numbers[1], numbers[2]

	// 特殊规则，含0也是有效的街注
	if n1 == 0 && n2 == 1 && n3 == 2 {
		return true
	} else if n1 == 0 && n2 == 2 && n3 == 3 {
		return true
	}

	// 检查是否是连续的三个数字 (n, n+1, n+2) 并且在同一行
	return n2 == n1+1 && n3 == n2+1 && n1%3 == 1
}

// 检查是否是有效的角注
func isValidCorner(numbers []int) bool {
	if len(numbers) != 4 {
		return false
	}

	// 特殊规则，含0也是有效的角注
	if numbers[0] == 0 && numbers[1] == 1 && numbers[2] == 2 && numbers[3] == 3 {
		return true
	}

	// 检查是否是四个相邻数字的交汇角
	// 例如: 1,2,4,5 或 2,3,5,6 等
	contains := func(n int) bool {
		for _, num := range numbers {
			if num == n {
				return true
			}
		}
		return false
	}

	n := numbers[0]
	return contains(n) && contains(n+1) && contains(n+3) && contains(n+4)
}

// 检查是否是有效的线注
func isValidLine(numbers []int) bool {
	if len(numbers) != 6 {
		return false
	}

	// 检查是否是两行连续的街注
	// 例如: 1,2,3,4,5,6 或 4,5,6,7,8,9 等
	for i := 0; i < 5; i++ {
		if numbers[i+1] != numbers[i]+1 {
			return false
		}
	}

	return numbers[0]%3 == 1
}

// 检查是否是打注 (1-12, 13-24, 25-36)
func isDozen(numbers []int) bool {
	if len(numbers) != 12 {
		return false
	}

	// 检查是否在同一个打注区间
	first := numbers[0]
	var min, max int

	if first >= 1 && first <= 12 {
		min, max = 1, 12
	} else if first >= 13 && first <= 24 {
		min, max = 13, 24
	} else if first >= 25 && first <= 36 {
		min, max = 25, 36
	} else {
		return false
	}

	// 检查所有数字是否在区间内
	nums := make(map[int]bool)
	for _, n := range numbers {
		if n < min || n > max {
			return false
		}
		nums[n] = true
	}

	// 检查是否包含区间所有数字
	for i := min; i <= max; i++ {
		if !nums[i] {
			return false
		}
	}

	return true
}

// 检查是否是列注
func isColumn(numbers []int) bool {
	if len(numbers) != 12 {
		return false
	}

	// 检查是否在同一列 (1,4,7,... 或 2,5,8,... 或 3,6,9,...)
	col := numbers[0] % 3
	if col == 0 {
		col = 3
	}

	nums := make(map[int]bool)
	for _, n := range numbers {
		if n%3 != col%3 {
			return false
		}
		nums[n] = true
	}

	// 检查是否包含列中所有数字
	for i := col; i <= 36; i += 3 {
		if !nums[i] {
			return false
		}
	}

	return true
}

// 检查是否是奇偶注
func isOddEven(numbers []int) bool {
	if len(numbers) != 18 {
		return false
	}

	// 检查是否是所有奇数或所有偶数
	isOdd := numbers[0]%2 == 1
	for _, n := range numbers {
		if n == 0 {
			return false // 0不算奇数也不算偶数
		}
		if (n%2 == 1) != isOdd {
			return false
		}
	}

	// 检查是否包含所有奇数或偶数
	nums := make(map[int]bool)
	for _, n := range numbers {
		nums[n] = true
	}

	var requiredNumbers []int
	if isOdd {
		for i := 1; i <= 35; i += 2 {
			requiredNumbers = append(requiredNumbers, i)
		}
	} else {
		for i := 2; i <= 36; i += 2 {
			requiredNumbers = append(requiredNumbers, i)
		}
	}

	for _, n := range requiredNumbers {
		if !nums[n] {
			return false
		}
	}

	return true
}

// 检查是否是红黑注
func isRedBlack(numbers []int) bool {
	if len(numbers) != 18 {
		return false
	}

	// 检查是否是所有红色或所有黑色
	isRed := redNumbers[numbers[0]]
	for _, n := range numbers {
		if n == 0 {
			return false // 0不算红色也不算黑色
		}
		if redNumbers[n] != isRed {
			return false
		}
	}

	// 检查是否包含所有红色或黑色数字
	nums := make(map[int]bool)
	for _, n := range numbers {
		nums[n] = true
	}

	for n, isRedNum := range redNumbers {
		if isRedNum == isRed && !nums[n] {
			return false
		}
	}

	return true
}

// 检查是否是高低注 (1-18 或 19-36)
func isHighLow(numbers []int) bool {
	if len(numbers) != 18 {
		return false
	}

	// 检查是否在同一个高低区间
	first := numbers[0]
	var min, max int

	if first >= 1 && first <= 18 {
		min, max = 1, 18
	} else if first >= 19 && first <= 36 {
		min, max = 19, 36
	} else {
		return false
	}

	// 检查所有数字是否在区间内
	nums := make(map[int]bool)
	for _, n := range numbers {
		if n < min || n > max {
			return false
		}
		nums[n] = true
	}

	// 检查是否包含区间所有数字
	for i := min; i <= max; i++ {
		if !nums[i] {
			return false
		}
	}

	return true
}

// CheckWin 检查下注是否获胜
func CheckWin(betType BetType, betNumbers []int, winningNumber int) bool {
	if len(betNumbers) == 0 {
		return false
	}
	if winningNumber == 0 {
		// 特殊规则，0 也可以参与其它下注类型
		return betNumbers[0] == winningNumber
	}

	switch betType {
	case Straight:
		return len(betNumbers) == 1 && betNumbers[0] == winningNumber
	case Split:
		for _, n := range betNumbers {
			if n == winningNumber {
				return true
			}
		}
	case Street:
		// 检查是否在同一个街注中
		min := betNumbers[0]
		if min == 0 {
			for _, n := range betNumbers {
				if n == winningNumber {
					return true
				}
			}
		}
		max := min + 2
		return winningNumber >= min && winningNumber <= max
	case Corner:
		// 检查是否在四个角注数字中
		for _, n := range betNumbers {
			if n == winningNumber {
				return true
			}
		}
	case Line:
		// 检查是否在两行线注中
		min := betNumbers[0]
		max := min + 5
		return winningNumber >= min && winningNumber <= max
	case Dozen:
		// 检查是否在打注区间
		min := betNumbers[0]
		max := min + 11
		return winningNumber >= min && winningNumber <= max
	case Column:
		// 检查是否在同一列
		col := betNumbers[0] % 3
		return winningNumber%3 == col
	case OddEven:
		// 检查奇偶性
		return (winningNumber%2 == 1) == (betNumbers[0]%2 == 1)
	case RedBlack:
		// 检查颜色
		return redNumbers[winningNumber] == redNumbers[betNumbers[0]]
	case HighLow:
		// 检查高低区间
		if betNumbers[0] <= 18 {
			return winningNumber <= 18
		}
		return winningNumber >= 19
	default:
		return false
	}

	return false
}

// CalculatePayout 计算赔付金额
func CalculatePayout(betType BetType, amount int64) int64 {
	payout, ok := Payouts[betType]
	if !ok {
		return 0
	}
	return int64(amount) * int64(payout+1)
}
