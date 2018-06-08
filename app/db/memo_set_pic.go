package db

import (
	"bytes"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/bitcoin/script"
	"github.com/memocash/memo/app/bitcoin/wallet"
	"html"
	"sort"
	"strings"
	"time"
)

type MemoSetPic struct {
	Id         uint   `gorm:"primary_key"`
	TxHash     []byte `gorm:"unique;size:50"`
	ParentHash []byte
	PkHash     []byte `gorm:"index:pk_hash"`
	PkScript   []byte `gorm:"size:500"`
	Address    string
	Url        string `gorm:"size:500"`
	BlockId    uint
	Block      *Block
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (m MemoSetPic) Save() error {
	result := save(&m)
	if result.Error != nil {
		return jerr.Get("error saving memo test", result.Error)
	}
	return nil
}

func (m MemoSetPic) GetTransactionHashString() string {
	hash, err := chainhash.NewHash(m.TxHash)
	if err != nil {
		jerr.Get("error getting chainhash from memo post", err).Print()
		return ""
	}
	return hash.String()
}

func (m MemoSetPic) GetAddressString() string {
	pkHash, err := btcutil.NewAddressPubKeyHash(m.PkHash, &wallet.MainNetParamsOld)
	if err != nil {
		jerr.Get("error getting pubkeyhash from memo post", err).Print()
		return ""
	}
	return pkHash.EncodeAddress()
}

func (m MemoSetPic) GetScriptString() string {
	return html.EscapeString(script.GetScriptString(m.PkScript))
}

func (m MemoSetPic) GetTimeString() string {
	if m.BlockId != 0 {
		return m.Block.Timestamp.Format("2006-01-02 15:04:05")
	}
	return "Unconfirmed"
}

func (m MemoSetPic) GetExtension() string {
	if strings.HasSuffix(m.Url, "jpg") {
		return "jpg"
	} else {
		return "png"
	}
}

func GetMemoSetPicById(id uint) (*MemoSetPic, error) {
	var memoSetPic MemoSetPic
	err := find(&memoSetPic, MemoSetPic{
		Id: id,
	})
	if err != nil {
		return nil, jerr.Get("error getting memo set pic", err)
	}
	return &memoSetPic, nil
}

func GetMemoSetPic(txHash []byte) (*MemoSetPic, error) {
	var memoSetPic MemoSetPic
	err := find(&memoSetPic, MemoSetPic{
		TxHash: txHash,
	})
	if err != nil {
		return nil, jerr.Get("error getting memo set pic", err)
	}
	return &memoSetPic, nil
}

type memoSetPicSortByDate []*MemoSetPic

func (txns memoSetPicSortByDate) Len() int      { return len(txns) }
func (txns memoSetPicSortByDate) Swap(i, j int) { txns[i], txns[j] = txns[j], txns[i] }
func (txns memoSetPicSortByDate) Less(i, j int) bool {
	if bytes.Equal(txns[i].ParentHash, txns[j].TxHash) {
		return true
	}
	if bytes.Equal(txns[i].TxHash, txns[j].ParentHash) {
		return false
	}
	if txns[i].Block == nil && txns[j].Block == nil {
		return false
	}
	if txns[i].Block == nil {
		return true
	}
	if txns[j].Block == nil {
		return false
	}
	return txns[i].Block.Height > txns[j].Block.Height
}

func GetPicForPkHash(pkHash []byte) (*MemoSetPic, error) {
	pics, err := GetSetPicsForPkHash(pkHash)
	if err != nil {
		return nil, jerr.Get("error getting set pics for pk hash", err)
	}
	if len(pics) == 0 {
		return nil, nil
	}
	return pics[0], nil
}

func GetPicsForPkHashes(pkHashes [][]byte) ([]*MemoSetPic, error) {
	db, err := getDb()
	if err != nil {
		return nil, jerr.Get("error getting db", err)
	}
	joinSelect := "JOIN (" +
		"	SELECT MAX(id) AS id" +
		"	FROM memo_set_pics" +
		"	GROUP BY pk_hash" +
		") sq ON (sq.id = memo_set_pics.id)"
	query := db.
		Table("memo_set_pics").
		Select("memo_set_pics.*, blocks.*").
		Joins(joinSelect).
		Joins("JOIN blocks ON (memo_set_pics.block_id = blocks.id)").
		Order("blocks.timestamp DESC").
		Where("pk_hash IN (?)", pkHashes)
	rows, err := query.Rows()
	if err != nil {
		return nil, jerr.Get("error getting set pics", err)
	}
	var memoSetPics []*MemoSetPic
	for rows.Next() {
		var memoSetPic = MemoSetPic{
			Block: &Block{},
		}
		err = rows.Scan(
			&memoSetPic.Id,
			&memoSetPic.TxHash,
			&memoSetPic.ParentHash,
			&memoSetPic.PkHash,
			&memoSetPic.PkScript,
			&memoSetPic.Address,
			&memoSetPic.Url,
			&memoSetPic.BlockId,
			&memoSetPic.CreatedAt,
			&memoSetPic.UpdatedAt,
			&memoSetPic.Block.Id,
			&memoSetPic.Block.Height,
			&memoSetPic.Block.Timestamp,
			&memoSetPic.Block.Hash,
			&memoSetPic.Block.PrevBlock,
			&memoSetPic.Block.MerkleRoot,
			&memoSetPic.Block.Nonce,
			&memoSetPic.Block.TxnCount,
			&memoSetPic.Block.Version,
			&memoSetPic.Block.Bits,
			&memoSetPic.Block.CreatedAt,
			&memoSetPic.Block.UpdatedAt,
		)
		if err != nil {
			return nil, jerr.Get("error scanning set pic", err)
		}
		memoSetPics = append(memoSetPics, &memoSetPic)
	}

	var setPics []*MemoSetPic
SetPicLoop:
	for _, memoSetPic := range memoSetPics {
		for _, setPic := range setPics {
			if bytes.Equal(setPic.PkHash, memoSetPic.PkHash) {
				continue SetPicLoop
			}
		}
		setPics = append(setPics, memoSetPic)
	}
	return setPics, nil
}

func GetSetPicsForPkHash(pkHash []byte) ([]*MemoSetPic, error) {
	var memoSetPics []*MemoSetPic
	err := findPreloadColumns([]string{
		BlockTable,
	}, &memoSetPics, &MemoSetPic{
		PkHash: pkHash,
	})
	if err != nil {
		return nil, jerr.Get("error getting memo pics", err)
	}
	sort.Sort(memoSetPicSortByDate(memoSetPics))
	return memoSetPics, nil
}
