package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"log"
)

func generateRedeemScript(preImage string) (string, error) {
	if preImage == "" {
		return "", errors.New("empty preImage")
	}
	// Create the lock script
	lockScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_SHA256).AddData([]byte(preImage)).AddOp(txscript.OP_EQUAL).Script()
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(lockScript), nil
}

func deriveAddress(redeemScript string) (string, error) {
	// Check if redeem script is empty
	if redeemScript == "" {
		return "", errors.New("empty redeemScript")
	}
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

func constructSpendingTransaction(previousTx *wire.MsgTx, lockingScript string, unlockingScript string) (*wire.MsgTx, error) {
	// Create a new spending transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	// Add an input to the transaction with the unlocking script
	if unlockingScript == "" {
		return nil, errors.New("empty unlocking script")
	}

	unlockingScriptBytes, err := hex.DecodeString(unlockingScript)
	if err != nil {
		return nil, errors.New("invalid hex format in unlocking script: " + err.Error())
	}

	// Check for empty unlocking script bytes
	if len(unlockingScriptBytes) == 0 {
		return nil, errors.New("empty unlocking script bytes")
	}

	// Add an input to the transaction with the unlocking script
	txIn := wire.NewTxIn(&wire.OutPoint{
		Hash:  previousTx.TxHash(),
		Index: 0,
	}, unlockingScriptBytes, nil)
	tx.AddTxIn(txIn)

	// Add an output to the transaction
	destAddress := "mr6M79HZLa2R9r5KKJrtNK3VpqaiEQ8C2b"
	destAddr, err := btcutil.DecodeAddress(destAddress, &chaincfg.RegressionNetParams)
	if err != nil {
		return nil, errors.New("error decoding destination address: " + err.Error())
	}

	// Create a new transaction output
	destPkScript, err := txscript.PayToAddrScript(destAddr)
	if err != nil {
		return nil, errors.New("error creating PayToAddrScript: " + err.Error())
	}

	// Add a fee to the transaction
	fee := int64(1000)
	txOut := wire.NewTxOut(previousTx.TxOut[0].Value-fee, destPkScript)
	tx.AddTxOut(txOut)

	// Check for insufficient funds
	if txOut.Value < 0 {
		return nil, errors.New("insufficient funds")
	}

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
	lockingScript := redeemScript

	//Provide the actual preimage data that satisfies the OP_SHA256 requirement
	actualPreimage := "Btrust Builders"
	hashedPreimage := sha256.Sum256([]byte(actualPreimage))
	unlockingScript := hex.EncodeToString(hashedPreimage[:])

	//Construct the spending transaction
	spendingTx, err := constructSpendingTransaction(tx, lockingScript, unlockingScript)
	if err != nil {
		log.Fatal("Error constructing spending transaction:", err)
	}
	//Print the spending transaction hash
	spendingTxHash := spendingTx.TxHash()
	fmt.Println("Spending Transaction Hash:", spendingTxHash)
}
