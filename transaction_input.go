package main

import "bytes"

type TXInput struct {
	Txid      []byte  // Txid = Transaction{nil, inputs, outputs}.Hash()
	Vout      int     // always 0 in our forked code
	Signature []byte  // where the signed data is transaction.TrimmedCopy(). See Sign() in transaction.go
	PubKey    []byte  // public key associated with the signature
}

// UsesKey checks whether the address initiated the transaction
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
