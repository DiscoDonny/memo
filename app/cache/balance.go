package cache

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/db"
)

func GetBalance(pkHash []byte) (int64, error) {
	var bal int64
	err := GetItem(getBalanceName(pkHash), &bal)
	if err == nil {
		return bal, nil
	} else if ! IsMissError(err) {
		return 0, jerr.Get("error getting balance", err)
	}
	outs, err := db.GetSpendableTransactionOutputsForPkHash(pkHash)
	if err != nil {
		return 0, jerr.Get("error getting outs", err)
	}
	var balance int64
	for _, out := range outs {
		balance += out.Value
	}
	err = SetBalance(pkHash, balance)
	if err != nil {
		jerr.Get("error setting balance in cache", err).Print()
	}
	return balance, nil
}

func SetBalance(pkHash []byte, balance int64) error {
	err := SetItem(getBalanceName(pkHash), balance)
	if err != nil {
		return jerr.Get("error setting balance", err)
	}
	return nil
}

func ClearBalance(pkHash []byte) error {
	err := DeleteItem(getBalanceName(pkHash))
	if err != nil {
		return jerr.Get("error clearing balance", err)
	}
	return nil
}

func getBalanceName(pkHash []byte) string {
	return fmt.Sprintf("balance-%x", pkHash)
}
