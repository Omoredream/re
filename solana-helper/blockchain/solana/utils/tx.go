package utils

import (
	"encoding/base64"

	"github.com/gagliardetto/binary"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/mr-tron/base58"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.wkr.moe/web3/solana-helper/consts"
)

type Transaction struct {
	Transaction solana.Transaction
	Meta        rpc.TransactionMeta
}

func SerializeTransaction(ctx g.Ctx, tx *solana.Transaction, strict bool) (txRaw []byte, err error) {
	accounts, err := tx.Message.GetAllKeys()
	if err != nil {
		err = gerror.Wrapf(err, "解析交易锁定账户失败")
		return
	}

	if count := len(accounts); count > consts.MaxTxAccountCount {
		err = gerror.Newf("交易锁定账户 %d 个 超出范围 (%d 个)", count, consts.MaxTxAccountCount)
		if strict {
			return
		} else {
			g.Log().Warningf(ctx, "%v", err)
			err = nil
		}
	}

	txRaw, err = tx.MarshalBinary()
	if err != nil {
		err = gerror.Wrapf(err, "序列化交易失败")
		return
	}

	if size := len(txRaw); size > consts.MaxTxSize {
		err = gerror.Newf("交易大小 %d Bytes 超出范围 (%d Bytes)", size, consts.MaxTxSize)
		if strict {
			return
		} else {
			g.Log().Warningf(ctx, "%v", err)
			err = nil
		}
	}

	return
}

func SerializeTransactionBase64(ctx g.Ctx, tx *solana.Transaction, strict bool) (txRaw string, err error) {
	raw, err := SerializeTransaction(ctx, tx, strict)
	if err != nil {
		err = gerror.Wrapf(err, "编码交易失败")
		return
	}

	txRaw = base64.StdEncoding.EncodeToString(raw)

	return
}

func SerializeTransactionBase58(ctx g.Ctx, tx *solana.Transaction, strict bool) (txRaw string, err error) {
	raw, err := SerializeTransaction(ctx, tx, strict)
	if err != nil {
		err = gerror.Wrapf(err, "编码交易失败")
		return
	}

	txRaw = base58.Encode(raw)

	return
}

func DeserializeTransaction(txRaw []byte) (tx *solana.Transaction, err error) {
	tx, err = solana.TransactionFromDecoder(bin.NewBinDecoder(txRaw))
	if err != nil {
		err = gerror.Wrapf(err, "反序列化交易失败")
		return
	}

	return
}

func DeserializeTransactionBase64(txRaw string) (tx *solana.Transaction, err error) {
	raw, err := gbase64.DecodeString(txRaw)
	if err != nil {
		err = gerror.Wrapf(err, "解码失败")
		return
	}

	tx, err = DeserializeTransaction(raw)
	if err != nil {
		err = gerror.Wrapf(err, "解码交易失败")
		return
	}

	return
}

func DeserializeTransactionBase58(txRaw string) (tx *solana.Transaction, err error) {
	raw, err := base58.Decode(txRaw)
	if err != nil {
		err = gerror.Wrapf(err, "解码失败")
		return
	}

	tx, err = DeserializeTransaction(raw)
	if err != nil {
		err = gerror.Wrapf(err, "解码交易失败")
		return
	}

	return
}
