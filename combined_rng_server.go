package main

import (
    "context"
    "encoding/json"
    "log"
    crand "crypto/rand"
    "math/big"
    "math/rand"
    "net"
    "net/http"
    "sync"
    "time"

    "gitee.com/heartfun/rouletteserv/proto"
    "google.golang.org/grpc"
)

// Rng implements the proto.RngServer interface
// and provides random numbers
// (copied from your rng/server.go)
type Rng struct {
	proto.UnimplementedRngServer
}

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
			   num := rd.Intn(int(Max-Min+1)) + int(Min)
			   rngs = append(rngs, uint32(num))
		   }
	   }
	return &proto.ReplyRngs{
		Rngs: rngs,
	}, nil
}

func main() {
	var wg sync.WaitGroup
	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", ":6000")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		proto.RegisterRngServer(grpcServer, NewRng())
		log.Println("gRPC RNG server listening on :6000")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		http.HandleFunc("/api/random_card", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			// Directly use the RNG logic
			   // Simulate proto.RequestRngs{Nums: 1}
			   Max := 36
			   Min := 0
			   n, err := crand.Int(crand.Reader, big.NewInt(int64(Max+1)))
			   var cardNumber uint32
			   if err == nil {
				   cardNumber = uint32(n.Int64())
			   } else {
				   rd := rand.New(rand.NewSource(time.Now().UnixNano()))
				   cardNumber = uint32(rd.Intn(int(Max-Min+1)) + int(Min))
			   }
			   w.Header().Set("Content-Type", "application/json")
			   json.NewEncoder(w).Encode(map[string]uint32{"card_number": cardNumber})
		})
		log.Println("HTTP bridge listening on :50497 (for /api/random_card)")
		if err := http.ListenAndServe(":50497", nil); err != nil {
			log.Fatalf("failed to serve HTTP: %v", err)
		}
	}()

	wg.Wait()
}
