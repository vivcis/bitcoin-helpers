package main

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"reflect"
	"strings"
	"testing"
)

func TestGenerateRedeemScript(t *testing.T) {
	tests := []struct {
		name     string
		preImage string
		want     string
		wantErr  bool
	}{
		{
			name:     "Valid PreImage",
			preImage: "Btrust Builders",
			want:     "a80f427472757374204275696c6465727387",
			wantErr:  false,
		},
		{
			name:     "Empty PreImage",
			preImage: "",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateRedeemScript(tt.preImage)

			if (err != nil) != tt.wantErr {
				t.Errorf("generateRedeemScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && err.Error() != "empty preImage" {
				t.Errorf("generateRedeemScript() unexpected error message, got = %v, want = empty preImage", err)
				return
			}

			if got != tt.want {
				t.Errorf("generateRedeemScript() got = %v, want %v", got, tt.want)
				t.Logf("Actual Hex Decoded: %v", hex.EncodeToString([]byte(got)))
			}
		})
	}
}

func TestDeriveAddress(t *testing.T) {
	tests := []struct {
		name         string
		redeemScript string
		want         string
		wantErr      bool
	}{
		{
			name:         "Valid RedeemScript",
			redeemScript: "a80f427472757374204275696c6465727387",
			want:         "2MytaPKkM6FYRt7PgUSSfwvMwYsHrQLbH9W",
			wantErr:      false,
		},
		{
			name:         "Invalid RedeemScript",
			redeemScript: "invalidhex",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "Empty RedeemScript",
			redeemScript: "",
			want:         "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deriveAddress(tt.redeemScript)

			if (err != nil) != tt.wantErr {
				t.Errorf("deriveAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("deriveAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstructTransaction(t *testing.T) {
	type args struct {
		address string
		amount  int64
	}
	tests := []struct {
		name    string
		args    args
		want    *wire.MsgTx
		wantErr bool
	}{
		{
			name: "Valid Transaction",
			args: args{
				address: "2MytaPKkM6FYRt7PgUSSfwvMwYsHrQLbH9W",
				amount:  100000,
			},
			want:    createExpectedTransaction("2MytaPKkM6FYRt7PgUSSfwvMwYsHrQLbH9W", 100000),
			wantErr: false,
		},
		{
			name: "Invalid Address",
			args: args{
				address: "invalid_address",
				amount:  1000,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := constructTransaction(tt.args.address, tt.args.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("constructTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err.Error() != "decoded address is of unknown format" {
					t.Errorf("constructTransaction() unexpected error message, got = %v, want = decoded address is of unknown format", err)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("constructTransaction() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstructSpendingTransaction(t *testing.T) {
	// Create a sample previous transaction
	previousTx := createExpectedTransaction("2MytaPKkM6FYRt7PgUSSfwvMwYsHrQLbH9W", 100000)

	type args struct {
		previousTx      *wire.MsgTx
		lockingScript   string
		unlockingScript string
	}

	tests := []struct {
		name                 string
		args                 args
		want                 *wire.MsgTx
		wantErr              bool
		expectedErrSubstring string
	}{
		{
			name: "Valid Spending Transaction",
			args: args{
				previousTx:      previousTx,
				lockingScript:   "a80f427472757374204275696c6465727387",
				unlockingScript: "b3a1e0bb961a57a5e105fb102a29b8266994292a69c25bfe4c8f7b781d40c944",
			},
			want:    createExpectedSpendingTransaction(previousTx, 1000),
			wantErr: false,
		},
		//{
		//	name: "Invalid Unlocking Script",
		//	args: args{
		//		previousTx:      previousTx,
		//		lockingScript:   "a80f427472757374204275696c6465727387",
		//		unlockingScript: "68656c6c6f20776f726c64",
		//	},
		//	want:                 nil,
		//	wantErr:              true,
		//	expectedErrSubstring: "invalid hex format in unlocking script",
		//},
		//{
		//	name: "Insufficient Funds",
		//	args: args{
		//		previousTx:      previousTx,
		//		lockingScript:   "a80f427472757374204275696c6465727387",
		//		unlockingScript: "b3a1e0bb961a57a5e105fb102a29b8266994292a69c25bfe4c8f7b781d40c944",
		//	},
		//	want:                 nil,
		//	wantErr:              true,
		//	expectedErrSubstring: "insufficient funds",
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := constructSpendingTransaction(tt.args.previousTx, tt.args.lockingScript, tt.args.unlockingScript)
			if (err != nil) != tt.wantErr {
				t.Errorf("constructSpendingTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.expectedErrSubstring) {
					t.Errorf("constructSpendingTransaction() unexpected error message, got = %v, want = %v", err, tt.expectedErrSubstring)
				}

				fmt.Println("Unlocking Script:", tt.args.unlockingScript)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("constructSpendingTransaction() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create an expected transaction based on address and amount
func createExpectedTransaction(address string, amount int64) *wire.MsgTx {
	tx := wire.NewMsgTx(wire.TxVersion)
	addr, _ := btcutil.DecodeAddress(address, &chaincfg.RegressionNetParams)
	pkScript, _ := txscript.PayToAddrScript(addr)
	txOut := wire.NewTxOut(amount, pkScript)
	tx.AddTxOut(txOut)
	return tx
}

// Helper function to create an expected spending transaction based on previous transaction and fee
func createExpectedSpendingTransaction(previousTx *wire.MsgTx, fee int64) *wire.MsgTx {
	tx := wire.NewMsgTx(wire.TxVersion)
	unlockingScriptBytes, _ := hex.DecodeString("b3a1e0bb961a57a5e105fb102a29b8266994292a69c25bfe4c8f7b781d40c944")
	txIn := wire.NewTxIn(&wire.OutPoint{
		Hash:  previousTx.TxHash(),
		Index: 0,
	}, unlockingScriptBytes, nil)
	tx.AddTxIn(txIn)

	destAddress := "mr6M79HZLa2R9r5KKJrtNK3VpqaiEQ8C2b"
	destAddr, _ := btcutil.DecodeAddress(destAddress, &chaincfg.RegressionNetParams)
	destPkScript, _ := txscript.PayToAddrScript(destAddr)
	txOut := wire.NewTxOut(previousTx.TxOut[0].Value-fee, destPkScript)
	tx.AddTxOut(txOut)

	return tx
}
