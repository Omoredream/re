package tests

import (
	"context"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/errcode"

	"git.wkr.moe/web3/solana-helper/consts"
	Utils "git.wkr.moe/web3/solana-helper/utils"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	AddressLookupTable "git.wkr.moe/web3/solana-helper/blockchain/solana/addresslookuptable"
	Instruction "git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
	jitoTypes "git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jito/types"
	jupiterHTTP "git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/rpc"
	jupiterTypes "git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/types"
	Token "git.wkr.moe/web3/solana-helper/blockchain/solana/token"
	Wallet "git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/worth"
)

type routeBalance struct {
	Balance uint64
	Max     uint64
	MaxAt   int
}

func TestArb(t *testing.T) {
	err := testArb(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testArb(ctx g.Ctx) (err error) {
	debug := g.Cfg().MustGet(nil, "project.arb.debug", false).Bool()

	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF, true)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	payerWallets, err := officialPool.NewWalletsFromMnemonic(ctx, testWalletMnemonic, g.Cfg().MustGet(nil, "project.arb.derivationFrom", 0).Uint32(), g.Cfg().MustGet(nil, "project.arb.derivationTo", 50).Uint32(), true)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	var alt *AddressLookupTable.AddressLookupTable
	if g.Cfg().MustGet(nil, "project.arb.alt") != nil {
		go func() {
			for { // 定时从缓存拉取新数据
				var alt_ AddressLookupTable.AddressLookupTable
				alt_, err = officialPool.AddressLookupTableCacheGet(ctx, Address.NewFromBase58(g.Cfg().MustGet(nil, "project.arb.alt").String()))
				if err != nil {
					err = gerror.Wrapf(err, "导入地址查找表失败")
					return
				}

				alt = &alt_
				time.Sleep(1 * time.Minute)
			}
		}()
	}

	arbProgram := Address.NewFromBase58("MoneyymapoTpHK5zNmo877RwgNN74Wx7r6bS3aS7Buq").AsProgramAddress()

	tokenPrincipal := consts.SOL
	tokenBlacklists := g.Redis("arb")
	gtimer.AddSingleton(ctx, 1*time.Minute, func(ctx context.Context) {
		var rpcNames []string
		for cursor := uint64(0); ; {
			nextCursor, keys, err := tokenBlacklists.Scan(ctx, cursor, gredis.ScanOption{
				Type: "hash",
			})
			if err != nil {
				err = gerror.Wrapf(err, "获取代币黑名单 RPC 失败")
				g.Log("scheduler").Errorf(ctx, "%+v", err)
				return
			}

			rpcNames = append(rpcNames, keys...)
			if nextCursor == 0 {
				break
			}
			cursor = nextCursor
		}

		for _, rpcName := range rpcNames {
			tokenBlacklistM, err := tokenBlacklists.HGetAll(ctx, rpcName)
			if err != nil {
				err = gerror.Wrapf(err, "获取代币黑名单失败")
				g.Log("scheduler").Errorf(ctx, "%+v", err)
				return
			}

			var needJudge []string
			for tokenAddress, detail := range tokenBlacklistM.MapStrStr() {
				blackedTimeS, _ := gstr.List2(detail, "|")
				blackedTime := gtime.NewFromStr(blackedTimeS)
				if gtime.Now().Sub(blackedTime) >= 10*time.Second {
					needJudge = append(needJudge, tokenAddress)
				}
			}
			if len(needJudge) == 0 {
				continue
			}

			ttls, err := Utils.RedisHPTtl(ctx, tokenBlacklists, rpcName, needJudge...)
			if err != nil {
				err = gerror.Wrapf(err, "获取代币黑名单有效期失败")
				g.Log("scheduler").Errorf(ctx, "%+v", err)
				return
			}

			var needDelete []string
			for i, ttl := range ttls {
				if ttl == -1 {
					needDelete = append(needDelete, needJudge[i])
				}
			}
			if len(needDelete) == 0 {
				continue
			}

			_, err = tokenBlacklists.HDel(ctx, rpcName, needDelete...)
			if err != nil {
				err = gerror.Wrapf(err, "删除异常代币黑名单失败")
				g.Log("scheduler").Errorf(ctx, "%+v", err)
				return
			}
		}
	})

	thirdPayer := g.Cfg().MustGet(nil, "project.arb.thirdPayer", false).Bool()
	dynamicTip := g.Cfg().MustGet(nil, "project.arb.dynamicTip", true).Bool()

	balanceBase := decimal.NewFromFloat(g.Cfg().MustGet(nil, "project.arb.balanceBase", 50).Float64())
	balanceHalvedMaxRound := g.Cfg().MustGet(nil, "project.arb.balanceHalvedMaxRound", 8).Int() // 50/(1.75^8)
	balanceHalvedBase := decimal.NewFromFloat(1.75)

	profitToJitoTip := g.Cfg().MustGet(nil, "project.arb.profitToJitoTip", 85_00).Uint16() // 小费占利润比例
	profitToJitoTipDec := decimal.NewFromInt32(int32(profitToJitoTip))
	baseFee := consts.SignFee.Copy() // 1 条签名
	if thirdPayer {
		baseFee = baseFee.Mul(decimal.NewFromInt(2)) // 2 条签名
	}
	profitMin := decimal.NewFromFloat(g.Cfg().MustGet(nil, "project.arb.profitMin", 0).Float64())
	jitoTipMin := decimal.NewFromFloat(0.000_001_000)
	jitoTipMax := decimal.NewFromFloat(g.Cfg().MustGet(nil, "project.arb.jitoTipMax", 0.1).Float64())

	worthSlippage := decimal.NewFromFloat(0.98)

	extraCULimit := g.Cfg().MustGet(nil, "project.arb.extraCULimit", 5_0000).Uint32()

	tokenAccount, err := wallet.Account.Address.FindAssociatedTokenAccountAddress(tokenPrincipal.Address)
	if err != nil {
		err = gerror.Wrapf(err, "生成代币账户失败")
		return
	}

	arbEventAuthority, err := arbProgram.FindProgramDerivedAddress([][]byte{
		[]byte("__event_authority"),
	})
	if err != nil {
		err = gerror.Wrapf(err, "派生 Arb 日志鉴权账户失败")
		return
	}

	arbBalanceCacherAccount, err := arbProgram.FindProgramDerivedAddress([][]byte{
		[]byte("balance_cacher"),
		wallet.Account.Address.Bytes(),
	})
	if err != nil {
		err = gerror.Wrapf(err, "派生 Arb 余额缓存账户失败")
		return
	}

	arbUnwrapTipWSOLAccount, err := arbProgram.FindProgramDerivedAddress([][]byte{
		[]byte("unwrap_tip_wsol_account"),
		wallet.Account.Address.Bytes(),
	})
	if err != nil {
		err = gerror.Wrapf(err, "派生 Arb 解包小费 wSOL 账户失败")
		return
	}

	Utils.Parallel(ctx, 0, g.Cfg().MustGet(nil, "project.arb.parallel", 1).Int(), func(_ int, thread int) (err error) {
		ctx := context.WithValue(ctx, consts.CtxDerivation, thread)
		err = func(ctx g.Ctx) (err error) {
			var jupiterHTTPRPC *jupiterHTTP.RPC
			var swapTx *solana.Transaction
			var profit, profitAsSOL decimal.Decimal
			var foundTime *gtime.Time
			err = jupiterPool.MultiStep(ctx, []string{tokenPrincipal.Address.String()}, func(ctx g.Ctx) (err error) {
				jupiterHTTPRPC = ctx.Value(consts.CtxValueJupiterHttpRPC).(*jupiterHTTP.RPC)

				tokenExcludes, err := tokenBlacklists.HKeys(ctx, jupiterHTTPRPC.Name())
				if err != nil {
					err = gerror.Wrapf(err, "拉取中间代币黑名单失败")
					return
				}
				tokenExcludes = append(tokenExcludes, tokenPrincipal.Address.String())

				tokenMiddleAddressS, err := jupiterPool.GetRandomSupportToken(ctx, tokenExcludes...)
				if err != nil {
					err = gerror.Wrapf(err, "随机选择中间代币失败")
					return
				}
				tokenMiddleAddress := Address.NewFromBase58(tokenMiddleAddressS).AsTokenAddress()

				var tokenMiddle Token.Token
				if debug {
					tokenMiddle, err = officialPool.TokenCacheGet(ctx, tokenMiddleAddress)
					if err != nil {
						err = gerror.Wrapf(err, "查询代币失败")
						return
					}
					ctx = context.WithValue(ctx, consts.CtxToken, tokenMiddle.DisplayName())
				} else {
					ctx = context.WithValue(ctx, consts.CtxToken, tokenMiddleAddress.String())
				}

				var quote *jupiterHTTP.GetQuoteResponse
				balanceBefore := balanceBase.Copy()
				for round := range balanceHalvedMaxRound {
					if round > 0 {
						balanceBefore = balanceBefore.Div(balanceHalvedBase)
					}

					var quote1 jupiterHTTP.GetQuoteResponse
					quote1, err = jupiterPool.GetQuote(
						ctx,
						tokenPrincipal.Address, tokenMiddleAddress,
						lamports.Token2Lamports(balanceBefore, tokenPrincipal.Info.Decimalx),
					)
					if err != nil {
						if gerror.HasCode(err, jupiterTypes.NoRoutesFound) ||
							gerror.HasCode(err, jupiterTypes.CouldNotFindAnyRoute) {
							_, err = tokenBlacklists.HSet(ctx, jupiterHTTPRPC.Name(), map[string]any{
								tokenMiddleAddress.String(): gtime.Now().String() + "|无路由",
							})
							if err != nil {
								err = gerror.Wrapf(err, "添加无路由的中间代币到黑名单失败")
								return
							}
							err = Utils.RedisHPExpire(ctx, tokenBlacklists, jupiterHTTPRPC.Name(), (10 * time.Minute).Milliseconds(), tokenMiddleAddress.String())
							if err != nil {
								err = gerror.Wrapf(err, "代币设置临时黑名单有效期失败")
								return
							}

							err = gerror.NewCodef(errcode.IgnoreError, "无路由")
							return
						} else if gerror.HasCode(err, jupiterTypes.TokenNotTradable) {
							_, err = tokenBlacklists.HSet(ctx, jupiterHTTPRPC.Name(), map[string]any{
								tokenMiddleAddress.String(): gtime.Now().String() + "|不可交易",
							})
							if err != nil {
								err = gerror.Wrapf(err, "添加不可交易的中间代币到黑名单失败")
								return
							}
							err = Utils.RedisHPExpire(ctx, tokenBlacklists, jupiterHTTPRPC.Name(), (10 * time.Minute).Milliseconds(), tokenMiddleAddress.String())
							if err != nil {
								err = gerror.Wrapf(err, "代币设置临时黑名单有效期失败")
								return
							}

							err = gerror.NewCodef(errcode.IgnoreError, "不可交易")
							return
						} else if gerror.HasCode(err, jupiterTypes.RoutePlanDoesNotConsumeAllTheAmount) {
							if round < balanceHalvedMaxRound-1 {
								continue
							}
							break
						}
						err = gerror.Wrapf(err, "生成 A-B 交易意向失败")
						return
					}

					//g.Dump(quote1)

					var quote2 jupiterHTTP.GetQuoteResponse
					quote2, err = jupiterPool.GetQuote(
						ctx,
						tokenMiddleAddress, tokenPrincipal.Address,
						quote1.OutAmount,
					)
					if err != nil {
						if gerror.HasCode(err, jupiterTypes.NoRoutesFound) ||
							gerror.HasCode(err, jupiterTypes.CouldNotFindAnyRoute) {
							_, err = tokenBlacklists.HSet(ctx, jupiterHTTPRPC.Name(), map[string]any{
								tokenMiddleAddress.String(): gtime.Now().String() + "|无路由",
							})
							if err != nil {
								err = gerror.Wrapf(err, "添加无路由的中间代币到黑名单失败")
								return
							}
							err = Utils.RedisHPExpire(ctx, tokenBlacklists, jupiterHTTPRPC.Name(), (10 * time.Minute).Milliseconds(), tokenMiddleAddress.String())
							if err != nil {
								err = gerror.Wrapf(err, "代币设置临时黑名单有效期失败")
								return
							}

							err = gerror.NewCodef(errcode.IgnoreError, "无路由")
							return
						} else if gerror.HasCode(err, jupiterTypes.TokenNotTradable) {
							_, err = tokenBlacklists.HSet(ctx, jupiterHTTPRPC.Name(), map[string]any{
								tokenMiddleAddress.String(): gtime.Now().String() + "|不可交易",
							})
							if err != nil {
								err = gerror.Wrapf(err, "添加不可交易的中间代币到黑名单失败")
								return
							}
							err = Utils.RedisHPExpire(ctx, tokenBlacklists, jupiterHTTPRPC.Name(), (10 * time.Minute).Milliseconds(), tokenMiddleAddress.String())
							if err != nil {
								err = gerror.Wrapf(err, "代币设置临时黑名单有效期失败")
								return
							}

							err = gerror.NewCodef(errcode.IgnoreError, "不可交易")
							return
						} else if gerror.HasCode(err, jupiterTypes.RoutePlanDoesNotConsumeAllTheAmount) {
							if round < balanceHalvedMaxRound-1 {
								continue
							}
							break
						}
						err = gerror.Wrapf(err, "生成 B-A 交易意向失败")
						return
					}

					//g.Dump(quote2)

					balanceAfter := lamports.Lamports2Token(quote2.OutAmount, tokenPrincipal.Info.Decimalx)

					profit = balanceAfter.Sub(balanceBefore)

					if !profit.IsPositive() {
						if round < balanceHalvedMaxRound-1 {
							continue
						}
						break
					}

					if tokenPrincipal.Address != consts.SOL.Address {
						profitAsSOL, err = worth.SOLAmount(ctx, profit)
						if err != nil {
							err = gerror.Wrapf(err, "计算币本位利润失败")
							return
						}
						profitAsSOL = profitAsSOL.Mul(worthSlippage) // 给 SOL 价格下跌留足空间
					} else {
						profitAsSOL = profit.Copy()
					}
					if debug {
						g.Log().Infof(ctx, "币本位利润: SOL %s", decimals.DisplayBalance(profitAsSOL))
					}

					if profitAsSOL.LessThan(baseFee) {
						if round < balanceHalvedMaxRound-1 {
							continue
						}
						break
					}
					profitAsSOL = profitAsSOL.Sub(baseFee)
					if profitAsSOL.LessThan(jitoTipMin) || profitAsSOL.LessThan(profitMin) {
						if round < balanceHalvedMaxRound-1 {
							continue
						}
						break
					}

					foundTime = gtime.Now()

					quote = new(jupiterHTTP.GetQuoteResponse)
					*quote = quote1
					quote.OutputMint = quote2.OutputMint
					quote.OutAmount = 0
					quote.OtherAmountThreshold = quote.OutAmount
					quote.RoutePlan = append(quote.RoutePlan, quote2.RoutePlan...)

				routeChecker:
					for {
						routeBalances := map[string]*routeBalance{
							tokenPrincipal.Address.String(): {
								Balance: quote.InAmount,
								Max:     quote.InAmount,
								MaxAt:   -1,
							},
						}
						for i, routePlan := range quote.RoutePlan {
							routeBalances[routePlan.SwapInfo.InputMint].Balance -= routePlan.SwapInfo.InAmount

							if _, ok := routeBalances[routePlan.SwapInfo.OutputMint]; !ok {
								routeBalances[routePlan.SwapInfo.OutputMint] = &routeBalance{
									Balance: 0,
									Max:     0,
									MaxAt:   -1,
								}
							}
							routeBalances[routePlan.SwapInfo.OutputMint].Balance += routePlan.SwapInfo.OutAmount

							if Utils.MapFilterCount(routeBalances, func(balance *routeBalance) bool {
								return balance.Balance > 5
							}) == 1 { // 无多个中间代币
								if routeBalances[routePlan.SwapInfo.OutputMint].Balance > routeBalances[routePlan.SwapInfo.OutputMint].Max {
									routeBalances[routePlan.SwapInfo.OutputMint].Max = routeBalances[routePlan.SwapInfo.OutputMint].Balance
									routeBalances[routePlan.SwapInfo.OutputMint].MaxAt = i
								} else { // 类似 A-B-C-B-A 路径，其中 B 在二次换出后未发生增长的情况
									if routeBalances[routePlan.SwapInfo.OutputMint].MaxAt == -1 {
										err = gerror.Newf("路径判断错误, %v", quote.RoutePlan)
										return
									}
									if quote.RoutePlan[routeBalances[routePlan.SwapInfo.OutputMint].MaxAt].SwapInfo.OutputMint != quote.RoutePlan[i].SwapInfo.OutputMint {
										err = gerror.Newf("路径解析错误, %v", quote.RoutePlan)
										return
									}
									slip := float64(routeBalances[routePlan.SwapInfo.OutputMint].Max) / float64(routeBalances[routePlan.SwapInfo.OutputMint].Balance)
									quote.RoutePlan = append(quote.RoutePlan[:routeBalances[routePlan.SwapInfo.OutputMint].MaxAt+1], quote.RoutePlan[i+1:]...)
									for j := routeBalances[routePlan.SwapInfo.OutputMint].MaxAt + 1; j < len(quote.RoutePlan); j++ {
										quote.RoutePlan[j].SwapInfo.InAmount = uint64(math.Floor(float64(quote.RoutePlan[j].SwapInfo.InAmount) * slip))
										quote.RoutePlan[j].SwapInfo.OutAmount = uint64(math.Ceil(float64(quote.RoutePlan[j].SwapInfo.OutAmount) * slip))
										quote.RoutePlan[j].SwapInfo.FeeAmount = uint64(math.Round(float64(quote.RoutePlan[j].SwapInfo.FeeAmount) * slip))
									}
									continue routeChecker
								}
							}
						}
						if Utils.MapFilterCount(routeBalances, func(balance *routeBalance) bool {
							return balance.Balance > 5
						}) > 1 {
							err = gerror.Newf("路径余额检查错误, %v", quote.RoutePlan)
							return
						}
						break
					}

					break
				}
				if quote == nil {
					err = gerror.NewCodef(errcode.IgnoreError, "无套利机会")
					return
				}

				swapTx, err = jupiterPool.CreateSwapTransaction(ctx, wallet.Account.Address, *quote)
				if err != nil {
					if gerror.HasCode(err, jupiterTypes.CircularArbitrageIsDisabled) {
						g.Log().Warningf(ctx, "节点 %s 禁止套利", jupiterHTTPRPC.Name())
						return
					}
					err = gerror.Wrapf(err, "根据意向生成交易失败")
					return
				}

				return
			})
			if err != nil {
				if gerror.HasCode(err, errcode.CoolDownError) {
					g.Log().Warningf(ctx, "需要冷却: %v", err)
					time.Sleep(10 * time.Second)
				}
				err = gerror.Wrapf(err, "创建 swap 失败")
				return
			}
			ctx = context.WithValue(ctx, consts.CtxRPC, jupiterHTTPRPC.Name())

			ixs, blockhash, _, addressLookupTables, err := officialPool.UnpackTransaction(ctx, swapTx)
			if err != nil {
				err = gerror.Wrapf(err, "拆解交易失败")
				return
			}
			if addressLookupTables == nil {
				addressLookupTables = make(map[solana.PublicKey]solana.PublicKeySlice)
			}
			if alt != nil {
				addressLookupTables[alt.Address.PublicKey] = alt.AddressLookupTable
			}

			for i, ix := range ixs {
				var ok bool
				ok, err = Instruction.IsSetCULimit(ix)
				if err != nil {
					err = gerror.Wrapf(err, "无法寻找指令")
					return
				}

				if ok {
					ixNew := Instruction.SetCULimit{}
					err = ixNew.Deserialize(ix)
					if err != nil {
						err = gerror.Wrapf(err, "无法解析指令")
						return
					}

					ixNew.Limit += extraCULimit

					ixs[i], err = ixNew.ToIx()
					if err != nil {
						err = gerror.Wrapf(err, "构建交易失败")
						return
					}

					break
				}

				if i == len(ixs)-1 {
					err = gerror.Newf("未找到 CULimit 指令")
					return
				}
			}

			swapIxIndex := slices.IndexFunc(ixs, func(ix solana.Instruction) bool {
				return ix.ProgramID() == consts.JupiterProgramV6Address.PublicKey
			})
			if swapIxIndex == -1 {
				err = gerror.Newf("无法切入 Swap")
				return
			}

			preSwapIx, err := Instruction.Custom{
				ProgramID: arbProgram,
				Accounts: solana.AccountMetaSlice{
					wallet.Account.Address.Meta().WRITE().SIGNER(), // searcher
					tokenAccount.Meta(),                            // token_account
					arbBalanceCacherAccount.Meta().WRITE(),         // balance_cacher

					consts.SystemProgramAddress.Meta(), // systemProgram
				},
				Discriminator: []byte{0xda, 0xb2, 0xcd, 0xbb, 0xe0, 0xf0, 0x0d, 0x78}, // BeforeArb
				Data:          nil,
			}.ToIx()
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}
			ixs = slices.Insert(ixs, swapIxIndex, preSwapIx)

			jitoTipAccount, err := jitoPool.GetRandomTipAccount(ctx)
			if err != nil {
				err = gerror.Wrapf(err, "获取 Jito 小费账户失败")
				return
			}

			var payer Wallet.HostedWallet
			if thirdPayer {
				payer = payerWallets[grand.Intn(len(payerWallets))]
			}

			var jitoTipStatic decimal.Decimal
			if !dynamicTip {
				jitoTipStatic = profitAsSOL.
					Mul(profitToJitoTipDec).
					Shift(-4) // Alias Div(100_00)
				if jitoTipStatic.GreaterThan(jitoTipMax) {
					jitoTipStatic = jitoTipMax.Copy()
				}
			}

			err = Instruction.Custom{
				ProgramID: arbProgram,
				Accounts: solana.AccountMetaSlice{
					wallet.Account.Address.Meta().WRITE().SIGNER(), // searcher
					tokenAccount.Meta().WRITE(),                    // token_account
					arbBalanceCacherAccount.Meta().WRITE(),         // balance_cacher
					lo.
						If(!thirdPayer, jitoTipAccount).
						Else(payer.Account.Address).
						Meta().WRITE(), // tip_account
					arbUnwrapTipWSOLAccount.Meta().WRITE(), // unwrap_tip_wsol_account

					consts.SystemProgramAddress.Meta(), // system_program
					consts.TokenProgramAddress.Meta(),  // token_program
					consts.SOL.Address.Meta(),          // wsol_mint
					arbEventAuthority.Meta(),           // event_authority
					arbProgram.Meta(),                  // program
				},
				Discriminator: []byte{0xa7, 0xde, 0xd6, 0x9d, 0x56, 0xa0, 0x1e, 0xf3}, // AfterArb
				Data: Utils.Append(
					Utils.Uint64ToBytesL(lamports.SOL2Lamports(baseFee)), // basic_fee
					Utils.BoolToBytes(thirdPayer),                        // third_payer
					Utils.Uint16ToBytesL(lo.
						If(dynamicTip, profitToJitoTip).
						Else(0),
					), // tip_bps
					Utils.Uint64ToBytesL(lamports.SOL2Lamports(lo.
						If(dynamicTip, jitoTipMax).
						Else(jitoTipStatic),
					)), // tip_max
					Utils.Int64ToBytesL(foundTime.UnixMilli()), // found_timestamp
					Utils.StringToBytes(Region),                // location
					Utils.BytesToBytes(Utils.XOR([]byte(jupiterHTTPRPC.Name()), []byte("6f91c859-788d-4cef-af2d-97f0c31d9394"))), // node
				),
			}.AppendIx(&ixs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}

			if !thirdPayer && !dynamicTip {
				err = Instruction.TransferSOL{
					Sender:   wallet,
					Receiver: jitoTipAccount,
					Amount:   jitoTipStatic,
				}.AppendIx(&ixs)
				if err != nil {
					err = gerror.Wrapf(err, "构建交易失败")
					return
				}
			}

			swapTx, err = officialPool.PackTransaction(ctx, ixs, blockhash, wallet.Account.Address, addressLookupTables)
			if err != nil {
				err = gerror.Wrapf(err, "组装交易失败")
				return
			}

			err = officialPool.SignTransaction(ctx, swapTx, []Wallet.HostedWallet{wallet})
			if err != nil {
				err = gerror.Wrapf(err, "签名交易失败")
				return
			}
			ctx = context.WithValue(ctx, consts.CtxTransaction, swapTx.Signatures[0].String())

			_, err = utils.SerializeTransaction(ctx, swapTx, true)
			if err != nil {
				err = gerror.WrapCodef(errcode.IgnoreError, err, "检查交易限制发现问题")
				return
			}

			txs := []*solana.Transaction{swapTx}

			if thirdPayer {
				ixs = make([]solana.Instruction, 0)

				cuLimit := Instruction.SetCULimit{
					Limit: 150,
				}

				if !dynamicTip {
					err = Instruction.TransferSOL{
						Sender:   payer,
						Receiver: jitoTipAccount,
						Amount:   jitoTipStatic,
					}.AppendIx(&ixs)
					if err != nil {
						err = gerror.Wrapf(err, "构建交易失败")
						return
					}
					cuLimit.Limit += 150

					err = Instruction.TransferSOL{
						Sender:   payer,
						Receiver: wallet.Account.Address,
						Amount:   lamports.Lamports2SOL(890880),
					}.AppendIx(&ixs)
					if err != nil {
						err = gerror.Wrapf(err, "构建交易失败")
						return
					}
					cuLimit.Limit += 150
				} else {
					err = Instruction.Custom{
						ProgramID: arbProgram,
						Accounts: solana.AccountMetaSlice{
							payer.Account.Address.Meta().WRITE().SIGNER(), // payer
							jitoTipAccount.Meta().WRITE(),                 // tip_account
							wallet.Account.Address.Meta().WRITE(),         // searcher

							consts.SystemProgramAddress.Meta(), // system_program
						},
						Discriminator: []byte{0x75, 0xa1, 0x1c, 0xe1, 0x82, 0x71, 0xb7, 0x26}, // PayTip
						Data:          nil,
					}.AppendIx(&ixs)
					if err != nil {
						err = gerror.Wrapf(err, "构建交易失败")
						return
					}
					cuLimit.Limit += 6468 // 5880 * 1.1
				}

				err = cuLimit.AppendIx(&ixs)
				if err != nil {
					err = gerror.Wrapf(err, "构建交易失败")
					return
				}

				var payTipTx *solana.Transaction
				payTipTx, err = officialPool.PackTransaction(ctx, ixs, blockhash, payer.Account.Address, addressLookupTables)
				if err != nil {
					err = gerror.Wrapf(err, "组装交易失败")
					return
				}

				err = officialPool.SignTransaction(ctx, payTipTx, []Wallet.HostedWallet{payer})
				if err != nil {
					err = gerror.Wrapf(err, "签名交易失败")
					return
				}

				txs = append(txs, payTipTx)
			}

			var bundleId string
			//if grand.MeetProb(0.5) {
			//	var bundleId_ *string
			//	_, bundleId_, err = jitoPool.SendTransaction(ctx, swapTx, true)
			//	if err != nil {
			//		err = gerror.Wrapf(err, "广播捆绑交易失败")
			//		return
			//	}
			//	bundleId = *bundleId_
			//} else {
			bundleId, err = jitoPool.SendBundle(ctx, txs...)
			if err != nil {
				if gerror.HasCode(err, gcode.New(-32602, "bundle contains an expired blockhash", nil)) {
					g.Log().Warningf(ctx, "Jupiter Swap API 区块滞后, %v", err)
					jupiterHTTPRPC.AddCoolDown((10 * time.Minute).Milliseconds())
					err = nil
					return
				}
				err = gerror.Wrapf(err, "广播捆绑交易失败")
				return
			}
			//}
			ctx = context.WithValue(ctx, consts.CtxBundle, bundleId)
			if debug {
				g.Log().Infof(ctx, "已发送交易")
			}

			if false {
				bundleStatus := jitoTypes.BundlePending
				for bundleStatus == jitoTypes.BundlePending || bundleStatus == jitoTypes.BundleInvalid {
					var bundleStatus_ jitoTypes.GetInflightBundleStatusesResponse
					bundleStatus_, err = jitoPool.GetInflightBundleStatus(ctx, bundleId)
					if err != nil {
						err = gerror.Wrapf(err, "获取捆绑交易状态失败")
						return
					}

					bundleStatus = bundleStatus_.Status
					g.Log().Infof(ctx, "捆绑交易状态: %s", bundleStatus)
				}
			}

			return
		}(ctx)
		if err != nil {
			if !gerror.HasCode(err, errcode.IgnoreError) {
				g.Log().Errorf(ctx, "%v", err)
			}
			err = nil
		}
		return
	})

	return
}
