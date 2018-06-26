package transaction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/bitcoin/memo"
	"github.com/memocash/memo/app/bitcoin/wallet"
)

type TxInfo struct {
	Error           error
	Raw             []byte
	Hash            string
	TotalInputValue int64
	OutputValue     int64
	Change          int64
	Fee             int64
	Size            int
	Inputs          []*TxInfoInput
	Outputs         []*TxInfoOutput
}

func (t TxInfo) PrintJson() {
	jsonData, err := json.Marshal(t)
	if err != nil {
		jerr.Get("error marshalling tx info", err).Print()
		return
	}
	fmt.Printf("%s\n", jsonData)
}

func (t TxInfo) Print() {
	if t.Error != nil {
		jerr.Get("error with tx info", t.Error).Print()
		return
	}
	var txnInfo = fmt.Sprintf("Txn: %s\n", t.Hash)
	txnInfo += fmt.Sprintf("Inputs:\n")
	for _, in := range t.Inputs {
		txnInfo = txnInfo + fmt.Sprintf("  - Value: %d, PrevOut: %s\n", in.Value, in.PrevOutHash)
	}
	txnInfo += fmt.Sprintf("Outputs:\n")
	for _, out := range t.Outputs {
		if out.Value > 0 {
			txnInfo = txnInfo + fmt.Sprintf("  - Value: %d, Address: %s\n    LockScript: %s\n",
				out.Value, out.Address.Address, out.LockScript)
		} else {
			txnInfo = txnInfo + fmt.Sprintf("  - Value: %d\n    LockScript: %s\n", out.Value, out.LockScript)
		}
	}
	txnInfo += fmt.Sprintf("TxSize: %d, InputValue: %d, OutputValue: %d, Change: %d, Fee: %d\n",
		t.Size, t.TotalInputValue, t.OutputValue, t.Change, t.Fee)
	txnInfo += fmt.Sprintf("Raw: %x\n", t.Raw)
	fmt.Printf("%s\n", txnInfo)
}

func (t TxInfo) PrintVerbose() {
	if t.Error != nil {
		jerr.Get("error with tx info", t.Error).Print()
		return
	}
	var txnInfo = fmt.Sprintf("Txn: %s\nRaw: %x\n", t.Hash, t.Raw)
	for _, in := range t.Inputs {
		txnInfo = txnInfo + fmt.Sprintf("  TxIn - value: %d\n"+
			"    Sequence: %d\n"+
			"    prevOut: %s\n"+
			"    unlockScript: %s\n"+
			"    signature: %x\n",
			in.Value, in.Sequence, in.PrevOutHash, in.UnlockScript, in.Signature)
	}
	for _, out := range t.Outputs {
		txnInfo = txnInfo + fmt.Sprintf("  TxOut - value: %d\n"+
			"    lockScript: %s\n", out.Value, out.LockScript)
		txnInfo = txnInfo + fmt.Sprintf("    address: %s\n"+
			"    scriptClass: %s\n"+
			"    requiredSigs: %d\n",
			out.Address.Address, out.Address.ScriptClass, out.Address.RequiredSigs)
	}
	txnInfo += fmt.Sprintf("TxSize: %d, InputValue: %d, Fee: %d, OutputValue: %d, Change: %d\n",
		t.Size, t.TotalInputValue, t.Fee, t.OutputValue, t.Change)
	fmt.Printf(txnInfo)
}

type TxInfoInput struct {
	Value        int64
	Signature    []byte
	Sequence     uint32
	PrevOutHash  string
	UnlockScript string
}

type TxInfoOutput struct {
	Value      int64
	LockScript string
	Address    TxInfoAddress
}

type TxInfoAddress struct {
	Address      string
	ScriptClass  string
	RequiredSigs int
}

func GetTxInfo(tx *memo.Tx) TxInfo {
	msg := tx.MsgTx
	writer := new(bytes.Buffer)
	err := msg.BtcEncode(writer, 1)
	if err != nil {
		return TxInfo{Error: jerr.Get("error encoding transaction", err)}
	}
	var txInfo = TxInfo{
		Raw:  writer.Bytes(),
		Hash: msg.TxHash().String(),
		Size: msg.SerializeSize(),
	}

	for _, in := range msg.TxIn {
		unlockScript, err := txscript.DisasmString(in.SignatureScript)
		if err != nil {
			return TxInfo{Error: jerr.Get("error disassembling unlockScript", err)}
		}
		txInfo.Inputs = append(txInfo.Inputs, &TxInfoInput{
			Signature:    in.SignatureScript,
			Sequence:     in.Sequence,
			PrevOutHash:  in.PreviousOutPoint.String(),
			UnlockScript: unlockScript,
		})
	}
	for _, out := range msg.TxOut {
		lockScript, err := txscript.DisasmString(out.PkScript)
		if err != nil {
			return TxInfo{Error: jerr.Get("error disassembling lockScript", err)}
		}
		scriptClass, addresses, sigCount, err := txscript.ExtractPkScriptAddrs(out.PkScript, &wallet.MainNetParamsOld)
		var txInfoAddress TxInfoAddress
		if out.Value > 0 {
			if len(addresses) != 1 {
				return TxInfo{Error: jerr.Newf("error unexpected address count (expected 1, got %d)", len(addresses))}
			}
			address := addresses[0]
			txInfoAddress = TxInfoAddress{
				Address:      address.String(),
				ScriptClass:  string(scriptClass),
				RequiredSigs: sigCount,
			}
			if bytes.Equal(address.ScriptAddress(), tx.SelfPkHash) {
				txInfo.Change += out.Value
			} else {
				txInfo.OutputValue += out.Value
			}
		}
		txInfo.Outputs = append(txInfo.Outputs, &TxInfoOutput{
			Value:      out.Value,
			LockScript: lockScript,
			Address:    txInfoAddress,
		})
	}
	for _, input := range tx.Inputs {
		txInfo.TotalInputValue += input.Value
		for _, txInfoInput := range txInfo.Inputs {
			if "out:"+txInfoInput.PrevOutHash == input.PrevOutHash {
				txInfoInput.Value = input.Value
			}
		}
	}
	txInfo.Fee = txInfo.TotalInputValue - txInfo.OutputValue - txInfo.Change
	return txInfo
}
