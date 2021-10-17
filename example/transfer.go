package main

import (
	"log"
	"math/big"

	"github.com/minswap/pab-go/cli"
	"github.com/minswap/pab-go/ledger"
	"github.com/minswap/pab-go/txbuilder"
)

const senderAddr = "addr_test1qpmtp5t0t5y6cqkaz7rfsyrx7mld77kpvksgkwm0p7en7qum7a589n30e80tclzrrnj8qr4qvzj6al0vpgtnmrkkksnqd8upj0"
const senderSkeyPath = "sender.skey"
const receiverAddr = "addr_test1qr2tn9mcgzu08lskmnekswwg9ghfrzhtlzx9t2vm4r96skx5hxthss9c70lpdh8ndquus23wjx9wh7yv2k5eh2xt4pvqzs2er6"
const transferLovelace = 10_000_000

func main() {
	cli, err := cli.New(cli.Options{
		NetworkID: cli.NetworkTestnet,
		Debug:     true,
	})
	if err != nil {
		log.Fatal(err)
	}

	utxos, err := cli.GetUtxosByAddress(senderAddr)
	if err != nil {
		log.Fatal(err)
	}

	transferVal := ledger.NewValue()
	transferVal.Add(ledger.ADA, big.NewInt(transferLovelace))

	txb := txbuilder.New(
		txbuilder.SpendPubKeyUtxos(utxos...),
		txbuilder.PayChangeTo(senderAddr),
		txbuilder.PayToPubKey(receiverAddr, transferVal),
	)

	tx, err := cli.BuildTx(txb)
	if err != nil {
		log.Fatal(err)
	}
	if err := cli.SubmitTxWithSkey(tx, senderSkeyPath); err != nil {
		log.Fatal(err)
	}
}
