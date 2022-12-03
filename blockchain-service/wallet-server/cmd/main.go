package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	wallet_server "blockchain/blockchain-service/wallet-server"
)

func init() {
	log.SetPrefix("Wallet server: ")
}

func main() {
	p := flag.Uint("port", 5001, "TCP Port Number for wallet server")
	gateway := flag.String("gateway", "http://127.0.0.1:5001", "TCP Port Number for wallet server")
	flag.Parse()

	walletSrv := wallet_server.New(uint16(*p), *gateway)
	transport := wallet_server.NewTransport(walletSrv)

	http.HandleFunc("/", transport.HandleIndex)
	http.HandleFunc("/wallet", transport.HandleWallet)
	http.HandleFunc("/transactions", transport.HandleTransaction)

	if err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(*p)), nil); err != nil {
		log.Fatalf("Failed ListenAndServe with err: %s", err)
	}
}