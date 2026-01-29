package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gitee.com/heartfun/rouletteserv/proto"
	"google.golang.org/grpc"
)

func main() {
	// Connect to the running gRPC RNG server (default port 6000)
	conn, err := grpc.Dial("localhost:6000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC RNG server: %v", err)
	}
	defer conn.Close()
	rngClient := proto.NewRngClient(conn)

	       http.HandleFunc("/api/random_card", func(w http.ResponseWriter, r *http.Request) {
		       // Allow CORS
		       w.Header().Set("Access-Control-Allow-Origin", "*")
		       w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		       w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		       if r.Method == http.MethodOptions {
			       w.WriteHeader(http.StatusOK)
			       return
		       }
		       ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		       defer cancel()
		       // Request 1 random number
		       req := &proto.RequestRngs{Nums: 1}
		       resp, err := rngClient.GetRngs(ctx, req)
		       if err != nil || len(resp.Rngs) == 0 {
			       w.WriteHeader(http.StatusInternalServerError)
			       w.Write([]byte(`{"error": "Failed to get random number"}`))
			       return
		       }
		       cardNumber := resp.Rngs[0]
		       w.Header().Set("Content-Type", "application/json")
		       json.NewEncoder(w).Encode(map[string]uint32{"card_number": cardNumber})
	       })

	log.Println("HTTP bridge listening on :50497 (for /api/random_card)")
	http.ListenAndServe(":50497", nil)
}
