package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"blockchain/blockchain-service/autominer"
	"blockchain/blockchain-service/blockchain-server"
	"blockchain/foundation/cryptography"
)

func init() {
	log.SetPrefix("Blockchain server: ")
}

func main() {
	p := flag.Uint("port", 5000, "TCP Port Number for Blockchain server")
	//i := flag.Int("automineInterval", 10, "Automine interval in minutes")
	flag.Parse()

	managingSrv := blockchain_server.New(uint16(*p), cryptography.GenerateBlockchainAddress)
	bc, err := managingSrv.GetBlockchain("BLOCKCHAIN")
	if err != nil {
		log.Fatalf("Failed to GetBlockchain with err: %s", err)
	}

	//am := autominer.New(time.Minute*time.Duration(*i), bc)
	am := autominer.New(time.Second*10, bc)
	ctx := context.Background()
	am.Start(ctx)

	transport := blockchain_server.NewTransport(managingSrv)

	http.HandleFunc("/chains", transport.HandleGetChain)
	http.HandleFunc("/transactions", transport.HandleTransactions)
	http.HandleFunc("/mining", transport.HandleMining)

	if err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(*p)), nil); err != nil {
		log.Fatalf("Failed ListenAndServe with err: %s", err)
	}
}
