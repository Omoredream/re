package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"
)

func (pool *RPCs) SignTransaction(ctx g.Ctx, tx *solana.Transaction, signers []Wallet.HostedWallet) (err error) {
	messageContent, err := tx.Message.MarshalBinary()
	if err != nil {
		err = gerror.Wrapf(err, "编码待签名的交易失败")
		return
	}

	allSigners := tx.Message.Signers()
	tx.Signatures = Utils.GrowSize(tx.Signatures, len(allSigners))
	for i := range tx.Signatures {
		if tx.Signatures[i].IsZero() {
			var privateKey *solana.PrivateKey
			for j := range signers {
				if allSigners[i] == signers[j].Account.Address.PublicKey {
					privateKey = &signers[j].PrivateKey
					break
				}
			}
			if privateKey == nil {
				g.Log().Warningf(ctx, "未找到签名地址 %s 的私钥", allSigners[i])
				continue
			}
			tx.Signatures[i], err = privateKey.Sign(messageContent)
			if err != nil {
				err = gerror.Wrapf(err, "交易签名失败")
				return
			}
		}
	}

	err = tx.VerifySignatures()
	if err != nil {
		g.Log().Warningf(ctx, "交易签名不匹配, %v", err)
		err = nil
	}

	return
}
