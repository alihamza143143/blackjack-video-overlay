package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	"gitee.com/heartfun/rouletteserv/game"
	"gitee.com/heartfun/rouletteserv/rng"
	"gitee.com/heartfun/rouletteserv/server"
	"gitee.com/heartfun/rouletteserv/gateway"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// 解析命令行参数
	mode := flag.String("mode", "roulette", "Service mode: roulette or rng")
	port := flag.String("port", "6000", "Port to listen on")
	rngAddr := flag.String("rng", "", "Address of RNG service (optional for roulette mode)")
	numRounds := flag.String("count", "100000000", "Number of rounds to calculate RTP for (optional for rtp mode)")
	debug := flag.Bool("debug", false, "sets log level to debug")
	betWindow  := flag.Int("betWindow", 30, "bet window length in seconds")
    pauseWindow := flag.Int("pauseWindow", 10, "pause window length in seconds")
    rouletteAddr := flag.String("roulette", "localhost:6000", "Address of Roulette service")
	flag.Parse()

	if modeStr := os.Getenv("MODE"); modeStr != "" {
		*mode = modeStr
	}
	if portStr := os.Getenv("PORT"); portStr != "" {
		*port = portStr
	}
	if rngAddrStr := os.Getenv("RNG"); rngAddrStr != "" {
		*rngAddr = rngAddrStr
	}
	if debugStr := os.Getenv("DEBUG"); debugStr != "" {
		*debug = debugStr == "true"
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.MessageFieldName = "msg"
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	switch *mode {
	case "roulette":
		if err := server.StartServer(*port, *rngAddr); err != nil {
			log.Err(err).Msg("Failed to start roulette server")
		}
	case "rng":
		if err := rng.StartServer(*port); err != nil {
			log.Err(err).Msg("Failed to start RNG server")
		}
	case "rtp":
		log.Info().Msg("start run rtp")
		numRounds, _ := strconv.Atoi(*numRounds)
		game.CalculateRTP(numRounds, *rngAddr)
		log.Info().Msg("rtp over")
		os.Exit(0)
	case "gateway":
        if err := gateway.Start(*port,
                                *rouletteAddr,
                                time.Duration(*betWindow)*time.Second,
                                time.Duration(*pauseWindow)*time.Second); err != nil {
        	log.Err(err).Msg("gateway exited with error")
        }
	default:
		log.Error().Msg("Invalid mode: " + *mode)
		os.Exit(1)
	}
}
