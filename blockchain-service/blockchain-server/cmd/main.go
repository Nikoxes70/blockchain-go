package main

import (
	"blockchain/blockchain-service/blockchain"
	"context"
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"blockchain/blockchain-service/autominer"
	"blockchain/blockchain-service/blockchain-server"
	syncer "blockchain/blockchain-service/neighbor-nodes-syncer"
	"blockchain/foundation/cryptography"
)

func init() {
	log.SetPrefix("Blockchain server: ")
}

func main() {
	p := flag.Uint("port", 5000, "TCP Port Number for Blockchain server")
	//ami := flag.Int("automineInterval", 10, "Automine interval in minutes")
	//nsi := flag.Int("neighborNodeSyncInterval", 10, "Neighbor node sync interval in seconds")
	bcAddress := flag.String("bcAddress", "0xAF909ba846284732E2a4Ec7bE12574CA937AAdd4", "Blockchain address")
	flag.Parse()

	bc, err := blockchain.NewBlockchain(*bcAddress)
	if err != nil {
		log.Fatalf("Failed to instantiate Blockchain with err: %s", err)
	}
	managingSrv := blockchain_server.New(uint16(*p), bc, cryptography.GenerateBlockchainAddress)

	//am := autominer.New(time.Minute*time.Duration(*i), bc)
	am := autominer.New(time.Second*10, managingSrv)
	amCtx := context.Background()
	go am.Start(amCtx)

	ns := syncer.New(time.Second*10, managingSrv)
	nsCtx := context.Background()
	go ns.Start(nsCtx)

	transport := blockchain_server.NewTransport(managingSrv)

	http.HandleFunc("/chain", transport.HandleGetChain)
	http.HandleFunc("/transactions", transport.HandleTransactions)
	http.HandleFunc("/mining", transport.HandleMining)
	http.HandleFunc("/balance", transport.HandleBalance)
	http.HandleFunc("/consensus", transport.HandleConsensus)

	if err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(*p)), nil); err != nil {
		log.Fatalf("Failed ListenAndServe with err: %s", err)
	}
}
