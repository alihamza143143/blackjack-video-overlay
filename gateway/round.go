package gateway

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"gitee.com/heartfun/rouletteserv/proto"
	"gitee.com/heartfun/rouletteserv/game"
	"github.com/rs/zerolog/log"
)

// phases of one roulette round.
type phase string

const (
	phaseOpen   phase = "open"   // accepting bets
	phaseResult phase = "result" // publish winning number
	phasePause  phase = "pause"  // bets closed, waiting
)

// liveBet mirrors proto.Bet plus a player id.
type liveBet struct {
	Client string      `json:"client"`
	Bet    *proto.Bet  `json:"bet"`
}

// roundMgr drives the game loop and talks to the gRPC backend.
type roundMgr struct {
	betWin, pauseWin time.Duration

	h    *hub
	grpc proto.GameLogicClient

	round int64

	mu   sync.Mutex
	bets []liveBet
	curPhase  phase

	manually bool
}

func (rm *roundMgr) setPhase(p phase) {
    rm.mu.Lock()
    rm.curPhase = p
    rm.mu.Unlock()
}

func newRoundMgr(h *hub, cli proto.GameLogicClient, betWin, pauseWin time.Duration) *roundMgr {
	rm := &roundMgr{
		h:        h,
		grpc:     cli,
		betWin:   betWin,
		pauseWin: pauseWin,
		round:    1,
		manually:  betWin == 0 && pauseWin == 0,
	}

    if !rm.manually {
        go rm.loop()
    } else {
        rm.openPhase()
    }

	return rm
}

func (rm *roundMgr) loop() {
	for {
		rm.openPhase()
		rm.pausePhase()
		rm.resultPhase()
	}
}

func (rm *roundMgr) openPhase() {
	rm.resetBets()
	rm.setPhase(phaseOpen)

	rm.h.broadcast(map[string]interface{}{
		"type":  "state",
		"value": phaseOpen,
		"round": rm.round,
        "duration": int64(rm.betWin.Seconds()),

	})
	time.Sleep(rm.betWin)
}

func (rm *roundMgr) pausePhase() {
    rm.setPhase(phasePause)

	rm.h.broadcast(map[string]interface{}{
		"type":  "state",
		"value": phasePause,
		"round": rm.round,
        "duration": int64(rm.pauseWin.Seconds()),

	})
}

func (rm *roundMgr) resultPhase() {
    rm.setPhase(phaseResult)

	// 1) take a snapshot of all live bets
	live := rm.snapshotBets()

	// 2) convert liveBet → *proto.Bet (server does not need client id)
	bets := make([]*proto.Bet, 0, len(live))
	for _, lb := range live {
		bets = append(bets, lb.Bet)
	}

	// 3) JSON-encode the slice exactly as the backend expects
	j, _ := json.Marshal(struct {
        Bets []*proto.Bet `json:"bets"`
    }{Bets: bets})


	// 4) build the gRPC request
	req := &proto.RequestPlay{
        ClientParams: string(j),
    }

	// 5) call Play2
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := rm.grpc.Play2(ctx, req)
	if err != nil {
		log.Err(err).Msg("Play2 failed")
		rm.h.broadcast(map[string]interface{}{
			"type":  "state",
			"value": phaseResult,
			"round": rm.round,
			"error": err.Error(),
		})
		rm.round++
		return
	}

    // --- decode the wheel pocket -----------------------------------------
    var pocket int32
    if len(resp.RandomNumbers) != 0 {
        pocket = resp.RandomNumbers[0].Value % int32(game.NumberCount) // 0-36
    }

	// 6) broadcast the result
	rm.h.broadcast(map[string]interface{}{
		"type":  "state",
		"value": phaseResult,
		"round": rm.round,
		"data":  resp,
		"pocket": pocket,
	})

	rm.round++

    if !rm.manually {
        time.Sleep(rm.pauseWin)
    } else {
        rm.openPhase()
    }
}

func (rm *roundMgr) addBet(cl *client, b *proto.Bet) {
    if rm.curPhase != phaseOpen {
        data, _ := json.Marshal(map[string]interface{}{
            "type":  "error",
            "msg":   "bet window is closed",
            "round": rm.round,
        })
        select {
            case cl.tx <- data:
                default:
                    close(cl.tx)
                    _ = cl.ws.Close()
        }
        return
    }

	// 1) remember it for the current round
	rm.mu.Lock()
	rm.bets = append(rm.bets, liveBet{Client: cl.id, Bet: b})
	rm.mu.Unlock()

	// 2) log to console
	log.Info().
		Str("client", cl.id).
		Ints("numbers", ints32ToInts(b.Numbers)).
		Int64("amount", b.Amount).
		Msg("bet received")

    // 3) broadcast to all connected clients
    rm.h.broadcast(map[string]interface{}{
    	"type":   "bet",
    	"client": cl.id,
    	"bet":    b,
    })
}

func (rm *roundMgr) resetBets() {
	rm.mu.Lock()
	rm.bets = nil
	rm.mu.Unlock()
}

func (rm *roundMgr) snapshotBets() []liveBet {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	out := make([]liveBet, len(rm.bets))
	copy(out, rm.bets)
	return out
}

func ints32ToInts(src []int32) []int {
	dst := make([]int, len(src))
	for i, v := range src {
		dst[i] = int(v)
	}
	return dst
}

