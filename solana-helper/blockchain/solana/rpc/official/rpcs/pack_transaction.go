package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address_lookup_table"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
)

func (pool *RPCs) UnpackTransaction(ctx g.Ctx, tx *solana.Transaction) (ixs []solana.Instruction, blockhash solana.Hash, feePayer Address.AccountAddress, addressLookupTables map[solana.PublicKey]solana.PublicKeySlice, err error) {
	if len(tx.Message.AddressTableLookups) > 0 {
		var addressLookupTables_ AddressLookupTable.AddressLookupTables
		addressLookupTables_, err = pool.AddressLookupTableCacheGets(ctx, Address.NewsFromBytes32(tx.Message.AddressTableLookups.GetTableIDs()))
		if err != nil {
			err = gerror.Wrapf(err, "解析地址查找表失败")
			return
		}
		addressLookupTables = addressLookupTables_.ToAddressLookupTableMap()
		err = tx.Message.SetAddressTables(addressLookupTables)
		if err != nil {
			err = gerror.Wrapf(err, "设置地址查找表失败")
			return
		}
	}

	for _, ixCompiled := range tx.Message.Instructions {
		var programAddress solana.PublicKey
		programAddress, err = tx.ResolveProgramIDIndex(ixCompiled.ProgramIDIndex)
		if err != nil {
			err = gerror.Wrapf(err, "解析交易指令程序失败")
			return
		}

		var accounts []*solana.AccountMeta
		accounts, err = ixCompiled.ResolveInstructionAccounts(&tx.Message)
		if err != nil {
			err = gerror.Wrapf(err, "解析交易指令账户失败")
			return
		}

		err = Instruction.Custom{
			ProgramID: Address.NewFromBytes32(programAddress).AsProgramAddress(),
			Accounts:  accounts,
			Data:      ixCompiled.Data,
		}.AppendIx(&ixs)
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}
	}

	blockhash = tx.Message.RecentBlockhash

	if signers := tx.Message.Signers(); len(signers) > 0 {
		feePayer = Address.NewFromBytes32(signers[0])
	}

	return
}

func (pool *RPCs) PackTransaction(ctx g.Ctx, ixs []solana.Instruction, blockhash solana.Hash, feePayer Address.AccountAddress, addressLookupTables map[solana.PublicKey]solana.PublicKeySlice) (tx *solana.Transaction, err error) {
	if blockhash.IsZero() {
		blockhash, err = pool.GetLatestBlockhash(ctx)
		if err != nil {
			err = gerror.Wrapf(err, "获取最新区块失败")
			return
		}
	}

	tx, err = solana.NewTransaction(ixs, blockhash, solana.TransactionPayer(feePayer.PublicKey), solana.TransactionAddressTables(addressLookupTables))
	if err != nil {
		err = gerror.Wrapf(err, "构造交易失败")
		return
	}

	return
}
