package tests

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/samber/lo"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/address_lookup_table"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
)

func TestCreateAddressLookupTable(t *testing.T) {
	err := testCreateAddressLookupTable(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testCreateAddressLookupTable(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	lookupAddresses := []Address.AccountAddress{
		wallet.Account.Address,
	}

	ixs := make([]solana.Instruction, 0)

	slot, err := officialPool.GetLatestSlot(ctx)
	if err != nil {
		err = gerror.Wrapf(err, "获取最新插槽失败")
		return
	}

	altAddress, err := InstructionAddressLookupTable.Create{
		Funder: wallet.Account.Address,
		Owner:  wallet.Account.Address,
		Slot:   slot,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	err = InstructionAddressLookupTable.Extend{
		Funder:             wallet.Account.Address,
		AddressLookupTable: altAddress,
		Owner:              wallet.Account.Address,
		Addresses:          lookupAddresses,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	tx, err := officialPool.SignInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
	if err != nil {
		err = gerror.Wrapf(err, "构造交易失败")
		return
	}

	txRaw, err := utils.SerializeTransactionBase64(ctx, tx, false)
	if err != nil {
		err = gerror.Wrapf(err, "编码交易失败")
		return
	}
	g.Log().Infof(ctx, "构造交易 %s", txRaw)

	_, _, err = officialPool.SimulateTransaction(ctx, tx)
	if err != nil {
		err = gerror.Wrapf(err, "模拟交易失败")
		return
	}

	txHash, err := officialPool.SendTransaction(ctx, tx)
	if err != nil {
		err = gerror.Wrapf(err, "发送交易失败")
		return
	}
	ctx = context.WithValue(ctx, consts.CtxTransaction, txHash.String())
	g.Log().Infof(ctx, "已发送交易")

	spent, err := officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
	if err != nil {
		err = gerror.Wrapf(err, "等待交易确认失败")
		return
	}
	g.Log().Infof(ctx, "交易耗时 %s", spent)

	return
}

func TestExtendAddressLookupTable(t *testing.T) {
	err := testExtendAddressLookupTable(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testExtendAddressLookupTable(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	altAddress := Address.NewFromBase58("DfvMfkE312Ah8gE8zZxBWyrUsiGRrsFN7fGQKey2HeBU")
	alt, err := officialPool.AddressLookupTableCacheGet(ctx, altAddress)
	if err != nil {
		err = gerror.Wrapf(err, "读取地址查找表失败")
		return
	}

	arbProgram := Address.NewFromBase58("MoneyymapoTpHK5zNmo877RwgNN74Wx7r6bS3aS7Buq").AsProgramAddress()

	lookupAddresses := lo.Filter([]Address.AccountAddress{
		wallet.Account.Address,

		consts.SystemProgramAddress.AccountAddress,
		consts.TokenProgramAddress.AccountAddress,
		consts.SysVarInstructionsAddress.AccountAddress,
		consts.MemoProgramAddress.AccountAddress,
		lo.T2(lo.Must2(consts.JupiterProgramV6Address.FindProgramDerivedAddress([][]byte{[]byte("__event_authority")}))).A,

		arbProgram.AccountAddress,
		lo.T2(lo.Must2(arbProgram.FindProgramDerivedAddress([][]byte{[]byte("__event_authority")}))).A,
		lo.T2(lo.Must2(arbProgram.FindProgramDerivedAddress([][]byte{[]byte("balance_cacher"), wallet.Account.Address.Bytes()}))).A,
		lo.T2(lo.Must2(arbProgram.FindProgramDerivedAddress([][]byte{[]byte("unwrap_tip_wsol_account"), wallet.Account.Address.Bytes()}))).A,

		consts.KaminoLendingProgramAddress.AccountAddress,
		consts.KaminoMainMarket,
		consts.KaminoMainMarketAuthority,
		consts.KaminoMainMarketSOLReserve,
		consts.KaminoMainMarketSOLReserveLiquidity.AccountAddress,
		consts.KaminoMainMarketSOLReserveFeeReceiver.AccountAddress,
		consts.KaminoMainMarketUSDCReserve,
		consts.KaminoMainMarketUSDCReserveLiquidity.AccountAddress,
		consts.KaminoMainMarketUSDCReserveFeeReceiver.AccountAddress,
		consts.KaminoJitoMarket,
		consts.KaminoJitoMarketAuthority,
		consts.KaminoJitoMarketSOLReserve,
		consts.KaminoJitoMarketSOLReserveLiquidity.AccountAddress,
		consts.KaminoJitoMarketSOLReserveFeeReceiver.AccountAddress,
		consts.KaminoJLPMarket,
		consts.KaminoJLPMarketAuthority,
		consts.KaminoJLPMarketUSDCReserve,
		consts.KaminoJLPMarketUSDCReserveLiquidity.AccountAddress,
		consts.KaminoJLPMarketUSDCReserveFeeReceiver.AccountAddress,

		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(consts.SOL.Address)).AccountAddress,

		// 稳定币
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(consts.USDC.Address)).AccountAddress,
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(consts.USDT.Address)).AccountAddress,
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("DEkqHyPN7GMRJ5cArtQFAWefqbZb33Hyf6s5iCwjEonT").AsTokenAddress())).AccountAddress,     // USDe
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("Eh6XEPhSwoLv5wFApukmnaVSHQ6sAnoD9BmgmwQoN2sN").AsTokenAddress())).AccountAddress,     // sUSDe
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("USDSwr9ApdHk5bvJKMjzff41FfuX8bSxdKcR81vTwcA").AsTokenAddress())).AccountAddress,      // USDS
		lo.Must(wallet.Account.Address.FindAssociatedToken2022AccountAddress(Address.NewFromBase58("2b1kV6DkPAnxd5ixfnxCpjxmKwqjjaYmCZfHsFu24GXo").AsTokenAddress())).AccountAddress, // PYUSD
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("J9BcrQfX4p9D1bvLzRNCbMDv8f44a9LFdeqNE4Yk2WMD").AsTokenAddress())).AccountAddress,     // ISC
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("HzwqbKZw8HxMN6bF2yFZNrht3c2iXXzpKcFu7uBEDKtr").AsTokenAddress())).AccountAddress,     // EURC
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("BenJy1n3WTx9mTjEvy63e8Q1j4RqUc6E4VBMz3ir4Wo6").AsTokenAddress())).AccountAddress,     // USD*
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("9zNQRsGLjNKwCUU5Gq5LR8beUCPzQMVMqKAi3SSZh54u").AsTokenAddress())).AccountAddress,     // FDUSD

		// 价值币
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("3NZ9JMVBmGAqocybic2c7LQCJScmgsAZ6vQqTDzcqmJh").AsTokenAddress())).AccountAddress, // WBTC
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("7vfCXTUXx5WJV5JADk17DUJ4ksgau7utNKj4b963voxs").AsTokenAddress())).AccountAddress, // WETH
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("cbbtcf3aa214zXHbiAZQwf4122FBYbraNdFqgw4iMij").AsTokenAddress())).AccountAddress,  // cbBTC
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("zBTCug3er3tLyffELcvDNrKkCymbPWysGcWihESYfLg").AsTokenAddress())).AccountAddress,  // zBTC

		// 锚定币
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("27G8MtK7VtTcCHkpASjSDdkWWYfoqT6ggEuKidVJidD4").AsTokenAddress())).AccountAddress, // JLP
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("NUZ3FDWTtN5SP72BsefbsqpnbAY5oe21LE8bCSkqsEK").AsTokenAddress())).AccountAddress,  // FLP.1
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("9Fzv4s5t2bNwwJoeeywMwypop3JegsuDb1eDbMnPr4TX").AsTokenAddress())).AccountAddress, // sFLP.1
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("AbVzeRUss8QJYzv2WDizDJ2RtsD1jkVyRjNdAzX94JhG").AsTokenAddress())).AccountAddress, // FLP.2
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("CrdMPbjooMmz6RoVgUnczWoeZka2QF14pikcCTpzRMxz").AsTokenAddress())).AccountAddress, // sFLP.2
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("4PZTRNrHnxWBqLRvX5nuE6m1cNR8RqB4kWvVYjDkMd2H").AsTokenAddress())).AccountAddress, // FLP.3
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("6afu2XRPMg8JAhzBsJ9DXsQRCFhkzbC4UaFMZepm6AHb").AsTokenAddress())).AccountAddress, // sFLP.3
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("EngqvevoQ8yaNdtxY7sSh5J7NF74k3cDKi9v9pHi5H3B").AsTokenAddress())).AccountAddress, // FLP.4
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("GnxdTsSQNQ3FF72nTyWo4SUt59Tt1MqDkRRfoPtKjMvJ").AsTokenAddress())).AccountAddress, // sFLP.4
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("Ab6K8anKSwAz8VXJPVvAVjPQMJNoVhwzfF7FtAB5PNW9").AsTokenAddress())).AccountAddress, // FLP.5
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("EsdayVbDQYQdy54TQh5iASMTkCzmhxsx6MpCvyrtYaUZ").AsTokenAddress())).AccountAddress, // sFLP.5
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("C9HErFA8ABcAjfMRp4eySQx8dDB6u92CdqjsnSgYyHGa").AsTokenAddress())).AccountAddress, // FLP.6
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("Cwneu1mq39LExywGhY79nzxeAAGnQhPVijZr98P9hs3Q").AsTokenAddress())).AccountAddress, // sFLP.6
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("2aAQefifU14gxfc2FQHruFrp2UViLF4TYwzvbfyKFiFa").AsTokenAddress())).AccountAddress, // FLP.7
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("GZbxLBmvyQSzay1jozgykotcXFpLu2yKkW6u7huhis8X").AsTokenAddress())).AccountAddress, // sFLP.7

		// LST
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("5oVNBeEEQvYi1cX3ir8Dx5n1P7pdxydbGF2X4TxVusJm").AsTokenAddress())).AccountAddress, // INF
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("J1toso1uCk3RLmjorhTtrVwY9HJ7X8V9yYac6Y7kGCPn").AsTokenAddress())).AccountAddress, // JitoSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("BNso1VUJnh4zcfpZa6986Ea66P6TCp59hvtNJ8b1X85").AsTokenAddress())).AccountAddress,  // BNSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So").AsTokenAddress())).AccountAddress,  // mSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("jupSoLaHXQiZZTSfEWMTRRgpnyFm8f6sZdosWBjx93v").AsTokenAddress())).AccountAddress,  // JupSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("bSo13r4TkiE4KumL71LsHTPpL2euBYLFx6h9HP3piy1").AsTokenAddress())).AccountAddress,  // bSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("vSoLxydx6akxyMD9XEcPvGYNGq6Nn66oqVb3UkGkei7").AsTokenAddress())).AccountAddress,  // vSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("7Q2afV64in6N6SeZsAAB81TJzwDoD6zpqmHkzi9Dcavn").AsTokenAddress())).AccountAddress, // JSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("Bybit2vBJGhPF52GBdNaQfUJ6ZpThSgHBobjWZpLPb4B").AsTokenAddress())).AccountAddress, // bbSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("edge86g9cVz87xcpKpy3J77vbp4wYd9idEV562CCntt").AsTokenAddress())).AccountAddress,  // edgeSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("aeroXvCT6tjGVNyTvZy86tFDwE4sYsKCh7FbNDcrcxF").AsTokenAddress())).AccountAddress,  // aeroSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("Dso1bDeDjCQxTrWHqUUi63oBvV7Mdm6WaobLbQ7gnPQ").AsTokenAddress())).AccountAddress,  // dSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("he1iusmfkpAdwvxLNGV8Y1iSbj4rUy6yMhEA3fotn9A").AsTokenAddress())).AccountAddress,  // hSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("LSTxxxnJzKDFSLr4dUkPcmCf5VyryEqzPLz5j4bpxFp").AsTokenAddress())).AccountAddress,  // LST
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("jag58eRBC1c88LaAsRPspTMvoKJPbnzw9p9fREzHqyV").AsTokenAddress())).AccountAddress,  // jagSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("jucy5XJ76pHVvtPZb5TKRcGQExkwit2P5s4vY8UzmpC").AsTokenAddress())).AccountAddress,  // jucySOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("picobAEvs6w7QEknPce34wAE4gknZA9v5tTonnmHYdX").AsTokenAddress())).AccountAddress,  // picoSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("BonK1YhkXEGLZzwtcvRTip3gAL9nCeQD7ppZBLXhtTs").AsTokenAddress())).AccountAddress,  // bonkSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("LAinEtNLgpmCP9Rvsf5Hn8W6EhNiKLZQti1xfWMLy6X").AsTokenAddress())).AccountAddress,  // laineSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("MangmsBgFqJhW4cLUR9LxfVgMboY1xAoP8UUBiWwwuY").AsTokenAddress())).AccountAddress,  // mangoSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("strng7mqqc1MBJJV6vMzYbEqnwVGvKKGKedeCvtktWA").AsTokenAddress())).AccountAddress,  // strongSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("HUBsveNpjo5pWqNkH57QzxjQASdTVXcSK7bVKTSZtcSX").AsTokenAddress())).AccountAddress, // hubSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("Comp4ssDzXcLeu2MnLuGNNFC4cmLPMng8qWHPvzAMU1h").AsTokenAddress())).AccountAddress, // compassSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("sSo14endRuUbvQaJS3dq36Q829a3A6BEfoeeRGJywEh").AsTokenAddress())).AccountAddress,  // sSOL
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("7dHbWXmci3dT8UFYWYZweBLXgycu7Y3iL6trKn1Y7ARj").AsTokenAddress())).AccountAddress, // stSOL

		// 项目币
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN").AsTokenAddress())).AccountAddress,  // JUP
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("jtojtomepa8beP8AuQc6eXt5FriJwfFMwQx2v2f9mCL").AsTokenAddress())).AccountAddress,  // JTO
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("KMNo3nJsBXfcpJTVhZcXLW7RmTwTt4GVFE7suUBo9sS").AsTokenAddress())).AccountAddress,  // KMNO
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("HZ1JovNiVvGrGNiiYvEozEVgZ58xaU3RKwX8eACQBCt3").AsTokenAddress())).AccountAddress, // PYTH
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("METAewgxyPbgwsseH8T16a39CQ5VyVxZi9zXiDPY18m").AsTokenAddress())).AccountAddress,  // MPLX
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("Grass7B4RdKfBCjTKgSqnXkqjwiGvQyFbuSCUJr3XXjs").AsTokenAddress())).AccountAddress, // GRASS
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R").AsTokenAddress())).AccountAddress, // RAY
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("85VBFQZC9TZkfaptBWjvUw7YbZjy52A6mjtPGjstQAmQ").AsTokenAddress())).AccountAddress, // W
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("BZLbGTNCSFfoth2GYDtwr7e4imWzpR5jqcUuGEwr646K").AsTokenAddress())).AccountAddress, // IO
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("DriFtupJYLTosbwoN8koMbEYSx54aFAVLddWsbksjwg7").AsTokenAddress())).AccountAddress, // DRIFT
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("TNSRxcUxoT9xBG3de7PiJyTDYu7kskLqcpddxnEJAS6").AsTokenAddress())).AccountAddress,  // TNSR
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("ZEUS1aR7aX8DFFJf5QjWj2ftDDdNTroMNGo8YoQm3Gq").AsTokenAddress())).AccountAddress,  // ZEUS
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("CLoUDKc4Ane7HeQcPpE3YHnznRxhMimJ4MyaUqyHFzAu").AsTokenAddress())).AccountAddress, // CLOUD
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("MEFNBXixkEbait3xn9bkm8WsJzXtVsaJEn4c8Sam21u").AsTokenAddress())).AccountAddress,  // ME
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("SENDdRQtYMWaQrBroBrJ2Q53fgVuq95CV9UPGEvpCxa").AsTokenAddress())).AccountAddress,  // SEND
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("KENJSUYLASHUMfHyy5o4Hp2FdNqZg1AsUPhfH2kYvEP").AsTokenAddress())).AccountAddress,  // GRIFFAIN
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("8x5VqbHA8D7NkD52uNuS5nnt3PwA8pLD34ymskeSo2Wn").AsTokenAddress())).AccountAddress, // ZEREBRO
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("DBRiDgJAMsM95moTzJs7M9LnkGErpbv9v6CUR1DXnUu5").AsTokenAddress())).AccountAddress, // DBR
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("MNDEFzGvMt87ueuHvVU9VcTqsAP5b3fTGPsHuuPA5ey").AsTokenAddress())).AccountAddress,  // MNDE
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("rndrizKT3MK1iimdxRdWabcF7Zg7AR5T4nud4EkHBof").AsTokenAddress())).AccountAddress,  // RNDR
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("StepAscQoEioFxxWGnh2sLBDFp9d8rvKz2Yp39iDpyT").AsTokenAddress())).AccountAddress,  // STEP
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("ATLASXmbPQxBUYbxPsV97usA3fPQYEqzQBUHgiFCUsXx").AsTokenAddress())).AccountAddress, // ATLAS
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("poLisWXnNRwC6oBu1vHiuKQzFjGL4XDSu4g9qjz9qVk").AsTokenAddress())).AccountAddress,  // POLIS
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("hntyVP6YFm1Hg25TN9WGLqM12b8TQmcknKrdu1oxWux").AsTokenAddress())).AccountAddress,  // HNT

		// Meme 币

		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263").AsTokenAddress())).AccountAddress,     // BONK
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("7GCihgDB8fe6KNjn2MYtkzZcRjQy3t9GHdC8uHYmW2hr").AsTokenAddress())).AccountAddress,     // POPCAT
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("MEW1gQWJ3nEXg2qgERiKu7FAFj79PHvQVREQUzScPP5").AsTokenAddress())).AccountAddress,      // MEW
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("EKpQGSJtjMFqKZ9KQanSqYXRcF8fBopzLHYxdM65zcjm").AsTokenAddress())).AccountAddress,     // $WIF
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("CzLSujWBLFsSjncfkh59rUFqvafWcY5tzedWJSuypump").AsTokenAddress())).AccountAddress,     // GOAT
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("ukHH6c7mMyiWCf1b9pnWe25TSpkDDt3H5pQZgZ74J82").AsTokenAddress())).AccountAddress,      // BOME
		lo.Must(wallet.Account.Address.FindAssociatedToken2022AccountAddress(Address.NewFromBase58("7atgF8KQo4wJrD5ATGX7t1V2zVvykPJbFfNeVf1icFv1").AsTokenAddress())).AccountAddress, // $CWIF
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("GJAFwWjJ3vnTsrQVabjBVK2TYB1YtRCQXRDfDgUnpump").AsTokenAddress())).AccountAddress,     // ACT
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("ED5nyyWEzpPPiWimP8vYm7sD7TD3LAt3Q3gRTWHzPJBY").AsTokenAddress())).AccountAddress,     // MOODENG
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("Df6yfrKC8kZE3KNkrHERKzAetSxbrWeniQfyJY4Jpump").AsTokenAddress())).AccountAddress,     // CHILLGUY
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("A8C3xuqscfmyLrte3VmTqrAq8kgMASius9AFNANwpump").AsTokenAddress())).AccountAddress,     // FWOG
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("5z3EqYQo9HiCEs3R84RCDMu2n7anpDMxRhdK8PSWmrRC").AsTokenAddress())).AccountAddress,     // PONKE
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("3S8qX1MsMqRbiwKg2cQyx7nis1oHMgaCuc9c4VfvVdPN").AsTokenAddress())).AccountAddress,     // MOTHER
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("7BgBvyjrZX1YKz4oh9mjb8ZScatkkwb8DzFx7LoiVkM3").AsTokenAddress())).AccountAddress,     // SLERF
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("9BB6NFEcjBCtnNLFko2FqVQBq8HHM13kCyYcdQbgpump").AsTokenAddress())).AccountAddress,     // Fartcoin
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("2zMMhcVQEXDtdE6vsFS7S7D5oUodfJHE8vd1gnBouauv").AsTokenAddress())).AccountAddress,     // PENGU
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("eL5fUxj2J4CiQsmW85k5FG9DvuQjjUoBHoQBi2Kpump").AsTokenAddress())).AccountAddress,      // UFD
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("61V8vBaqAGMpgDQi4JcAwo1dmBGHsyhzodcPqnEVpump").AsTokenAddress())).AccountAddress,     // arc
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("CBdCxKo9QavR9hfShgpEBG3zekorAeD7W1jfq2o3pump").AsTokenAddress())).AccountAddress,     // LUCE
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("Hjw6bEcHtbHGpQr8onG3izfJY5DJiWdt7uk2BfdSpump").AsTokenAddress())).AccountAddress,     // SNAI
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("2qEHjDLDLbuBgRYvsxhc5D6uDWAivNFZGan56P1tpump").AsTokenAddress())).AccountAddress,     // Pnut
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("GJtJuWD9qYcCkrwMBmtY1tpapV1sKfB2zUv9Q4aqpump").AsTokenAddress())).AccountAddress,     // $RIF
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("FvgqHMfL9yn39V79huDPy3YUNDoYJpuLWng2JfmQpump").AsTokenAddress())).AccountAddress,     // $URO
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("M3M3pSFptfpZYnWNUgAbyWzKKgPo5d1eWmX6tbiSF2K").AsTokenAddress())).AccountAddress,      // M3M3
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("25hAyBQfoDhfWx9ay6rarbgvWGwDdNqcHsXS3jQ3mTDJ").AsTokenAddress())).AccountAddress,     // MANEKI
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("63LfDmNb3MQ8mw9MtZ2To9bEA2M71kZUUGq5tiJxcqj9").AsTokenAddress())).AccountAddress,     // GIGA
		lo.Must(wallet.Account.Address.FindAssociatedToken2022AccountAddress(Address.NewFromBase58("HeLp6NuQkmYB4pYWo2zYs22mESHXPQYzXbB8n4V98jwC").AsTokenAddress())).AccountAddress, // ai16z
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("5mbK36SZ7J19An8jFochhQS4of8g6BwUjbeCSxBSoWdp").AsTokenAddress())).AccountAddress,     // $michi
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("74SBV4zDXxTRgv1pEMoECskKBkZHc2yGPnc7GYVepump").AsTokenAddress())).AccountAddress,     // swarms
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("BLDiYcvm3CLcgZ7XUBPgz6idSAkNmWY6MBbm8Xpjpump").AsTokenAddress())).AccountAddress,     // TRISIG
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("7ZCm8WBN9aLa3o47SoYctU6iLdj7wkGG5SV2hE5CgtD5").AsTokenAddress())).AccountAddress,     // ELON(Wormhole)
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("5voS9evDjxF589WuEub5i4ti7FWQmZCsAsyD5ucbuRqM").AsTokenAddress())).AccountAddress,     // ELIZA
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("bobaM3u8QmqZhY1HwAtnvze9DLXvkgKYk3td3t8MLva").AsTokenAddress())).AccountAddress,      // BOBAOPPA
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("9DHe3pycTuymFk4H4bbPoAJ4hQrr2kaLDF6J6aAKpump").AsTokenAddress())).AccountAddress,     // BUZZ
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("7XJiwLDrjzxDYdZipnJXzpr1iDTmK55XixSFAa7JgNEL").AsTokenAddress())).AccountAddress,     // MLG
		lo.Must(wallet.Account.Address.FindAssociatedTokenAccountAddress(Address.NewFromBase58("6p6xgHyF7AeE6TZkSmFsko444wqoP15icUSqi2jfGiPN").AsTokenAddress())).AccountAddress,     // TRUMP
	}, func(address Address.AccountAddress, _ int) bool {
		return !alt.AddressLookupTable.Contains(address.PublicKey)
	})

	ixs := make([]solana.Instruction, 0)

	err = InstructionAddressLookupTable.Extend{
		Funder:             wallet.Account.Address,
		AddressLookupTable: altAddress,
		Owner:              wallet.Account.Address,
		Addresses:          lookupAddresses,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	tx, err := officialPool.SignInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
	if err != nil {
		err = gerror.Wrapf(err, "构造交易失败")
		return
	}

	txRaw, err := utils.SerializeTransactionBase64(ctx, tx, false)
	if err != nil {
		err = gerror.Wrapf(err, "编码交易失败")
		return
	}
	g.Log().Infof(ctx, "构造交易 %s", txRaw)

	_, _, err = officialPool.SimulateTransaction(ctx, tx)
	if err != nil {
		err = gerror.Wrapf(err, "模拟交易失败")
		return
	}

	txHash, err := officialPool.SendTransaction(ctx, tx)
	if err != nil {
		err = gerror.Wrapf(err, "发送交易失败")
		return
	}
	ctx = context.WithValue(ctx, consts.CtxTransaction, txHash.String())
	g.Log().Infof(ctx, "已发送交易")

	spent, err := officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
	if err != nil {
		err = gerror.Wrapf(err, "等待交易确认失败")
		return
	}
	g.Log().Infof(ctx, "交易耗时 %s", spent)

	return
}

func TestDeactivateAddressLookupTable(t *testing.T) {
	err := testDeactivateAddressLookupTable(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testDeactivateAddressLookupTable(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	altAddress := Address.NewFromBase58("DfvMfkE312Ah8gE8zZxBWyrUsiGRrsFN7fGQKey2HeBU")

	ixs := make([]solana.Instruction, 0)

	err = InstructionAddressLookupTable.Deactivate{
		AddressLookupTable: altAddress,
		Owner:              wallet.Account.Address,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	tx, err := officialPool.SignInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
	if err != nil {
		err = gerror.Wrapf(err, "构造交易失败")
		return
	}

	txRaw, err := utils.SerializeTransactionBase64(ctx, tx, false)
	if err != nil {
		err = gerror.Wrapf(err, "编码交易失败")
		return
	}
	g.Log().Infof(ctx, "构造交易 %s", txRaw)

	_, _, err = officialPool.SimulateTransaction(ctx, tx)
	if err != nil {
		err = gerror.Wrapf(err, "模拟交易失败")
		return
	}

	txHash, err := officialPool.SendTransaction(ctx, tx)
	if err != nil {
		err = gerror.Wrapf(err, "发送交易失败")
		return
	}
	ctx = context.WithValue(ctx, consts.CtxTransaction, txHash.String())
	g.Log().Infof(ctx, "已发送交易")

	spent, err := officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
	if err != nil {
		err = gerror.Wrapf(err, "等待交易确认失败")
		return
	}
	g.Log().Infof(ctx, "交易耗时 %s", spent)

	return
}

func TestCloseAddressLookupTable(t *testing.T) {
	err := testCloseAddressLookupTable(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testCloseAddressLookupTable(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	altAddress := Address.NewFromBase58("DfvMfkE312Ah8gE8zZxBWyrUsiGRrsFN7fGQKey2HeBU")

	ixs := make([]solana.Instruction, 0)

	err = InstructionAddressLookupTable.Close{
		AddressLookupTable: altAddress,
		Owner:              wallet.Account.Address,
		Beneficiary:        wallet.Account.Address,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	tx, err := officialPool.SignInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
	if err != nil {
		err = gerror.Wrapf(err, "构造交易失败")
		return
	}

	txRaw, err := utils.SerializeTransactionBase64(ctx, tx, false)
	if err != nil {
		err = gerror.Wrapf(err, "编码交易失败")
		return
	}
	g.Log().Infof(ctx, "构造交易 %s", txRaw)

	_, _, err = officialPool.SimulateTransaction(ctx, tx)
	if err != nil {
		err = gerror.Wrapf(err, "模拟交易失败")
		return
	}

	txHash, err := officialPool.SendTransaction(ctx, tx)
	if err != nil {
		err = gerror.Wrapf(err, "发送交易失败")
		return
	}
	ctx = context.WithValue(ctx, consts.CtxTransaction, txHash.String())
	g.Log().Infof(ctx, "已发送交易")

	spent, err := officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
	if err != nil {
		err = gerror.Wrapf(err, "等待交易确认失败")
		return
	}
	g.Log().Infof(ctx, "交易耗时 %s", spent)

	return
}
