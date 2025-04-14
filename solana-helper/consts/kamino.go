package consts

import (
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

var (
	KaminoMainMarket                      = Address.NewFromBase58("7u3HeHxYDLhnCoErrtycNokbQYbWGzLs6JSDqGAv5PfF")
	KaminoMainMarketAuthority             = Address.NewFromBase58("9DrvZvyWh1HuAoZxvYWMvkf2XCzryCpGgHqrMjyDWpmo")
	KaminoMainMarketSOLReserve            = Address.NewFromBase58("d4A2prbA2whesmvHaL88BH6Ewn5N4bTSU2Ze8P6Bc4Q")
	KaminoMainMarketSOLReserveLiquidity   = Address.NewFromBase58("GafNuUXj9rxGLn4y79dPu6MHSuPWeJR6UtTWuexpGh3U").AsTokenAccountAddress()
	KaminoMainMarketSOLReserveFeeReceiver = Address.NewFromBase58("3JNof8s453bwG5UqiXBLJc77NRQXezYYEBbk3fqnoKph").AsTokenAccountAddress()

	KaminoMainMarketUSDCReserve            = Address.NewFromBase58("D6q6wuQSrifJKZYpR1M8R4YawnLDtDsMmWM1NbBmgJ59")
	KaminoMainMarketUSDCReserveLiquidity   = Address.NewFromBase58("Bgq7trRgVMeq33yt235zM2onQ4bRDBsY5EWiTetF4qw6").AsTokenAccountAddress()
	KaminoMainMarketUSDCReserveFeeReceiver = Address.NewFromBase58("BbDUrk1bVtSixgQsPLBJFZEF7mwGstnD5joA1WzYvYFX").AsTokenAccountAddress()

	KaminoJitoMarket                      = Address.NewFromBase58("H6rHXmXoCQvq8Ue81MqNh7ow5ysPa1dSozwW3PU1dDH6")
	KaminoJitoMarketAuthority             = Address.NewFromBase58("Dx8iy2o46sK1DzWbEcznqSKeLbLVeu7otkibA3WohGAj")
	KaminoJitoMarketSOLReserve            = Address.NewFromBase58("6gTJfuPHEg6uRAijRkMqNc9kan4sVZejKMxmvx2grT1p")
	KaminoJitoMarketSOLReserveLiquidity   = Address.NewFromBase58("ywaaLvG7t1vXJo8sT3UzE8yzzZtxLM7Fmev64Jbooye").AsTokenAccountAddress()
	KaminoJitoMarketSOLReserveFeeReceiver = Address.NewFromBase58("EQ7hw63aBS7aPQqXsoxaaBxiwbEzaAiY9Js6tCekkqxf").AsTokenAccountAddress()

	KaminoJLPMarket                       = Address.NewFromBase58("DxXdAyU3kCjnyggvHmY5nAwg5cRbbmdyX3npfDMjjMek")
	KaminoJLPMarketAuthority              = Address.NewFromBase58("B9spsrMK6pJicYtukaZzDyzsUQLgc3jbx5gHVwdDxb6y")
	KaminoJLPMarketUSDCReserve            = Address.NewFromBase58("Ga4rZytCpq1unD4DbEJ5bkHeUz9g3oh9AAFEi6vSauXp")
	KaminoJLPMarketUSDCReserveLiquidity   = Address.NewFromBase58("GENey8es3EgGiNTM8H8gzA3vf98haQF8LHiYFyErjgrv").AsTokenAccountAddress()
	KaminoJLPMarketUSDCReserveFeeReceiver = Address.NewFromBase58("rywFxeqHfCWL5iGq9cDxutwiiR2TabZgzv8aRM1Lceq").AsTokenAccountAddress()
)
