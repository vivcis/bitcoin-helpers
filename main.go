package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

func generateRedeemScript(preImage string) (string, error) {
	// Create the lock script
	lockScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_SHA256).AddData([]byte(preImage)).AddOp(txscript.OP_EQUAL).Script()
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(lockScript), nil
}

func deriveAddress(redeemScript string) (string, error) {
	// Parse redeem script
	script, err := hex.DecodeString(redeemScript)
	if err != nil {
		return "", err
	}

	// Create P2SH address
	address, err := btcutil.NewAddressScriptHash(script, &chaincfg.RegressionNetParams)
	if err != nil {
		return "", err
	}

	return address.EncodeAddress(), nil
}

func constructTransaction(address string, amount int64) (*wire.MsgTx, error) {
	// Create a new transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	// Add an output to the transaction
	addr, err := btcutil.DecodeAddress(address, &chaincfg.RegressionNetParams)
	if err != nil {
		return nil, err
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, err
	}
	txOut := wire.NewTxOut(amount, pkScript)
	tx.AddTxOut(txOut)

	return tx, nil
}

func constructSpendingTransaction(previousTx *wire.MsgTx, redeemScript string) (*wire.MsgTx, error) {
	// Create a new spending transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	// Add an input to the transaction with the unlocking script
	unlockingScript, err := hex.DecodeString(redeemScript)
	if err != nil {
		return nil, err
	}
	txIn := wire.NewTxIn(&wire.OutPoint{
		Hash:  previousTx.TxHash(),
		Index: 0,
	}, unlockingScript, nil)
	tx.AddTxIn(txIn)

	// Add an output to the transaction
	destAddress := "mr6M79HZLa2R9r5KKJrtNK3VpqaiEQ8C2b"
	destAddr, err := btcutil.DecodeAddress(destAddress, &chaincfg.RegressionNetParams)
	if err != nil {
		return nil, err
	}
	destPkScript, err := txscript.PayToAddrScript(destAddr)
	if err != nil {
		return nil, err
	}
	txOut := wire.NewTxOut(previousTx.TxOut[0].Value-1000, destPkScript) // Deduct a small fee
	tx.AddTxOut(txOut)

	return tx, nil
}

func main() {
	// Task 1: Generate the redeem script
	preImage := "Btrust Builders"
	redeemScript, err := generateRedeemScript(preImage)
	if err != nil {
		log.Fatal("Error generating redeem script:", err)
	}
	fmt.Println("Redeem Script:", redeemScript)

	// Task 2: Derive an address from the redeem script
	address, err := deriveAddress(redeemScript)
	if err != nil {
		log.Fatal("Error deriving address:", err)
	}
	fmt.Println("Derived Address:", address)

	// Task 3: Construct a transaction that sends Bitcoins to the address
	tx, err := constructTransaction(address, 100000)
	if err != nil {
		log.Fatal("Error constructing transaction:", err)
	}
	txHash := tx.TxHash()
	fmt.Println("Transaction Hash:", txHash)

	// Task 4: Construct another transaction that spends from the previous transaction
	spendingTx, err := constructSpendingTransaction(tx, redeemScript)
	if err != nil {
		log.Fatal("Error constructing spending transaction:", err)
	}
	spendingTxHash := spendingTx.TxHash()
	fmt.Println("Spending Transaction Hash:", spendingTxHash)
}
