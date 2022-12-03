package main

import (
	"blockchain/blockchain-service/blockchain-server"
	"blockchain/foundation/cryptography"
	"flag"
	"log"
	"net/http"
	"strconv"
)

func init() {
	log.SetPrefix("Blockchain server: ")
}

func main() {
	p := flag.Uint("port", 5000, "TCP Port Number for Blockchain server")
	flag.Parse()

	managingSrv := blockchain_server.New(uint16(*p), cryptography.GenerateBlockchainAddress)

	transport := blockchain_server.NewTransport(managingSrv)

	http.HandleFunc("/chains", transport.HandleGetChain)
	http.HandleFunc("/transactions", transport.HandleTransactions)

	if err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(*p)), nil); err != nil {
		log.Fatalf("Failed ListenAndServe with err: %s", err)
	}
}
