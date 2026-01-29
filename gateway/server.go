package gateway

import (
	"net/http"
	"time"
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"gitee.com/heartfun/rouletteserv/proto"
)

// Start boots the websocket gateway and never returns unless an error occurs.
func Start(addr, rouletteAddr string, betWin, pauseWin time.Duration) error {
	grpcConn, err := grpc.Dial(rouletteAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer grpcConn.Close()

	h := newHub()
	rm := newRoundMgr(
    		h,
    		proto.NewGameLogicClient(grpcConn),
    		betWin,
    		pauseWin,
    )


	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		c := &client{
			ws: ws,
			tx: make(chan []byte, 16),
			id: r.RemoteAddr,
		}
		h.register <- c

		// writer
		go func() {
			for msg := range c.tx {
				_ = c.ws.WriteMessage(websocket.TextMessage, msg)
			}
		}()

		// reader
		go func(c *client) {
        	defer func() { _ = c.ws.Close() }()

        	for {
        		_, msg, err := c.ws.ReadMessage()
        		if err != nil {
        			return
        		}

        		//------------------------------------------------------------------
                // 1) detect a Spin command
                //------------------------------------------------------------------
                var spinProbe struct {
                	Spin json.RawMessage `json:"spin"`
                }
                if err := json.Unmarshal(msg, &spinProbe); err == nil && spinProbe.Spin != nil {
                	if rm.manually {
                		go rm.resultPhase()
                	}
                	continue // nothing else to do with this message
                }
                //------------------------------------------------------------------

                // 2) otherwise expect a bet-list payload
                var wrap struct {
                    Bets []*proto.Bet `json:"bets"`
                }
                if err := json.Unmarshal(msg, &wrap); err != nil || len(wrap.Bets) == 0 {
                	log.Warn().Err(err).Msg("invalid bet payload")
                	continue
                }

                // forward every bet to the round-manager
                for _, b := range wrap.Bets {
                	rm.addBet(c, b)
                }
        	}
        }(c)
	})

	log.Info().Str("addr", addr).Msg("gateway listening")
	return http.ListenAndServe(":"+addr, nil)
}