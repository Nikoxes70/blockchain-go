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
	flag.Parse()
	print(p)
	managingSrv := blockchain_server.New(uint16(*p), cryptography.GenerateBlockchainAddress)

	//am := autominer.New(time.Minute*time.Duration(*i), bc)
	am := autominer.New(time.Second*10, managingSrv)
	ctx := context.Background()
	go am.Start(ctx)

	ns := syncer.New(time.Second*10, managingSrv)
	go ns.Start(ctx)

	transport := blockchain_server.NewTransport(managingSrv)

	http.HandleFunc("/chains", transport.HandleGetChain)
	http.HandleFunc("/transactions", transport.HandleTransactions)
	http.HandleFunc("/mining", transport.HandleMining)
	http.HandleFunc("/balance", transport.HandleBalance)

	if err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(*p)), nil); err != nil {
		log.Fatalf("Failed ListenAndServe with err: %s", err)
	}
}
