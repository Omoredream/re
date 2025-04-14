package ProjectArb

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address_lookup_table"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jito/rpcs"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/rpcs"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/types"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/official/rpcs"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	interConsts "git.wkr.moe/web3/solana-helper/project/arb/internal/consts"

	"git.wkr.moe/web3/solana-helper/project/arb/internal/types"
)

type Arb struct {
	officialPool *officialRPCs.RPCs
	jupiterPool  *jupiterRPCs.RPCs
	jitoPool     *jitoRPCs.RPCs
	timer        *gtimer.Timer
	// 项目设置
	debug                        bool
	metrics                      bool
	region                       string
	parallel                     int
	simulate                     bool
	wallet                       Wallet.HostedWallet
	programBalanceCacherAddress  Address.AccountAddress
	programBUnwrapTipWSOLAddress Address.TokenAccountAddress
	// 交易设置
	tradeToken               Token.Token
	tradeTokenAccountAddress Address.TokenAccountAddress
	tradeSize                []decimal.Decimal
	tradeSizeRounds          int
	flashLoan                decimal.Decimal
	alt                      *AddressLookupTable.AddressLookupTable
	extraCULimit             uint32
	// jupiter 设置
	quoteOptions []jupiterTypes.QuoteOption
	// spam 配置
	enableSpam               bool
	spamProfitMin            decimal.Decimal
	priorityFeeFromProfitBps decimal.Decimal
	// jito 设置
	enableJito       bool
	jitoProfitMin    decimal.Decimal
	thirdTipPayer    bool
	thirdTipPayers   []Wallet.HostedWallet
	dynamicTip       bool
	tipFromProfitBps []types.TipFromProfitBps
	tipMax           decimal.Decimal
	// redis 设置
	tokenBlacklists *gredis.Redis
	txBlacklists    *gredis.Redis
	arbMetricsTS    *gredis.Redis
}

func New(ctx context.Context, officialPool *officialRPCs.RPCs, jupiterPool *jupiterRPCs.RPCs, jitoPool *jitoRPCs.RPCs, region string, wif string, mnemonic string) (arb *Arb, err error) {
	arb = &Arb{
		officialPool: officialPool,
		jupiterPool:  jupiterPool,
		jitoPool:     jitoPool,
		timer:        gtimer.New(),
	}

	arb.debug = g.Cfg().MustGet(ctx, "project.arb.debug", false).Bool()
	arb.metrics = g.Cfg().MustGet(ctx, "project.arb.metrics", false).Bool()
	arb.region = region
	arb.parallel = g.Cfg().MustGet(ctx, "project.arb.parallel", 1).Int()
	arb.simulate = g.Cfg().MustGet(ctx, "project.arb.simulate", false).Bool()
	arb.wallet, err = officialPool.NewWalletFromWIF(ctx, wif, true)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}
	arb.programBalanceCacherAddress, _, err = interConsts.ArbProgramAddress.FindProgramDerivedAddress([][]byte{
		[]byte("balance_cacher"),
		arb.wallet.Account.Address.Bytes(),
	})
	if err != nil {
		err = gerror.Wrapf(err, "派生 Arb 余额缓存账户失败")
		return
	}
	programBUnwrapTipWSOLAddress, _, err := interConsts.ArbProgramAddress.FindProgramDerivedAddress([][]byte{
		[]byte("unwrap_tip_wsol_account"),
		arb.wallet.Account.Address.Bytes(),
	})
	if err != nil {
		err = gerror.Wrapf(err, "派生 Arb 解包小费 wSOL 账户失败")
		return
	}
	arb.programBUnwrapTipWSOLAddress = programBUnwrapTipWSOLAddress.AsTokenAccountAddress()

	arb.tradeToken = consts.SOL // todo 实现自定义交易代币
	arb.tradeTokenAccountAddress, err = arb.wallet.Account.Address.FindAssociatedTokenAccountAddress(arb.tradeToken.Address)
	if err != nil {
		err = gerror.Wrapf(err, "生成代币账户失败")
		return
	}
	halvedCalcRound := g.Cfg().MustGet(ctx, "project.arb.trade.halvedCalcRound", 8).Int()
	halvedCalcIndex := decimal.NewFromFloat(g.Cfg().MustGet(ctx, "project.arb.trade.halvedCalcIndex", 1.75).Float64())
	halvedCustoms := g.Cfg().MustGet(ctx, "project.arb.trade.halvedCustoms", nil).Float64s()
	arb.tradeSize = make([]decimal.Decimal, 1, 1+halvedCalcRound+len(halvedCustoms))
	arb.tradeSize[0] = decimal.NewFromFloat(g.Cfg().MustGet(ctx, "project.arb.trade.tradeSizeBase", 50).Float64())
	for round := range halvedCalcRound {
		arb.tradeSize = append(arb.tradeSize, arb.tradeSize[round].Div(halvedCalcIndex))
	}
	for _, halvedCustom := range halvedCustoms {
		arb.tradeSize = append(arb.tradeSize, decimal.NewFromFloat(halvedCustom))
	}
	slices.SortFunc(arb.tradeSize, func(a, b decimal.Decimal) int {
		return b.Cmp(a)
	})
	arb.tradeSizeRounds = len(arb.tradeSize)
	arb.flashLoan = decimal.NewFromFloat(g.Cfg().MustGet(ctx, "project.arb.trade.flashLoan", 0).Float64())
	if altAddress := g.Cfg().MustGet(ctx, "project.arb.trade.alt", nil); altAddress != nil {
		var alt AddressLookupTable.AddressLookupTable
		alt, err = officialPool.AddressLookupTableCacheGet(ctx, Address.NewFromBase58(altAddress.String()))
		if err != nil {
			err = gerror.Wrapf(err, "导入地址查找表失败")
			return
		}
		arb.alt = &alt
		arb.timer.AddSingleton(ctx, 1*time.Minute, arb.updateALT)
	}
	arb.extraCULimit = g.Cfg().MustGet(ctx, "project.arb.trade.extraCULimit", 5_0000).Uint32()

	arb.quoteOptions = []jupiterTypes.QuoteOption{
		jupiterTypes.QuoteOptionSlippageBps(0),
	}
	multiHop := g.Cfg().MustGet(ctx, "project.arb.jupiter.multiHop", false).Bool()
	if !multiHop {
		arb.quoteOptions = append(arb.quoteOptions, jupiterTypes.QuoteOptionOnlyDirectRoutes)
	}

	arb.enableSpam = g.Cfg().MustGet(ctx, "project.arb.spam.enable", false).Bool()
	arb.spamProfitMin = decimal.NewFromFloat(g.Cfg().MustGet(ctx, "project.arb.spam.profitMin", 0).Float64())
	arb.priorityFeeFromProfitBps = decimal.NewFromInt32(int32(g.Cfg().MustGet(ctx, "project.arb.spam.priorityFeeFromProfitBps", 5_00).Uint16()))

	arb.enableJito = g.Cfg().MustGet(ctx, "project.arb.jito.enable", false).Bool()
	arb.jitoProfitMin = decimal.NewFromFloat(g.Cfg().MustGet(ctx, "project.arb.jito.profitMin", 0).Float64())
	if arb.jitoProfitMin.LessThan(interConsts.JitoTipMin) {
		arb.jitoProfitMin = interConsts.JitoTipMin
	}
	arb.thirdTipPayer = g.Cfg().MustGet(ctx, "project.arb.jito.thirdPayer", false).Bool()
	if arb.thirdTipPayer {
		arb.thirdTipPayers, err = officialPool.NewWalletsFromMnemonic(ctx, mnemonic, g.Cfg().MustGet(ctx, "project.arb.jito.derivationFrom", 0).Uint32(), g.Cfg().MustGet(ctx, "project.arb.jito.derivationTo", 50).Uint32(), true)
		if err != nil {
			err = gerror.Wrapf(err, "导入钱包失败")
			return
		}
	}
	arb.dynamicTip = g.Cfg().MustGet(ctx, "project.arb.jito.dynamicTip", false).Bool()
	tipFromProfitBps := g.Cfg().MustGet(ctx, "project.arb.jito.tipFromProfitBps", map[string]uint16{"0": 85_00}).MapStrVar()
	arb.tipFromProfitBps = make([]types.TipFromProfitBps, 0, len(tipFromProfitBps))
	for minProfit, bps := range tipFromProfitBps {
		arb.tipFromProfitBps = append(arb.tipFromProfitBps, types.TipFromProfitBps{
			MinProfit: lo.Must(decimal.NewFromString(minProfit)),
			BpsU16:    bps.Uint16(),
			BpsDec:    decimal.NewFromInt32(int32(bps.Uint16())),
		})
	}
	slices.SortFunc(arb.tipFromProfitBps, func(a, b types.TipFromProfitBps) int {
		return a.MinProfit.Cmp(b.MinProfit) // 升序
	})
	arb.tipMax = decimal.NewFromFloat(g.Cfg().MustGet(ctx, "project.arb.jito.tipMax", 0.1).Float64())

	arb.tokenBlacklists = g.Redis("arbTokenBlackLists")
	arb.timer.AddSingleton(ctx, 1*time.Minute, arb.fixTokenBlackLists)
	arb.txBlacklists = g.Redis("arbTxBlackLists")
	if arb.metrics {
		arb.arbMetricsTS = g.Redis("arbMetricsTS")
		_, err = arb.arbMetricsTS.Do(ctx, "TS.CREATE", "sendSpentBySpam", "RETENTION", (30 * gtime.D).Milliseconds(), "DUPLICATE_POLICY", "MAX")
		if err != nil {
			var redisErr redis.Error
			if errors.As(err, &redisErr) {
				if redisErr.Error() == "ERR TSDB: key already exists" {
					err = nil
					_, err = arb.arbMetricsTS.Do(ctx, "TS.ALTER", "sendSpentBySpam", "RETENTION", (30 * gtime.D).Milliseconds(), "DUPLICATE_POLICY", "MAX")
					if err != nil {
						err = gerror.Wrapf(err, "更新时序数据集失败")
						return
					}
				}
			}
			if err != nil {
				err = gerror.Wrapf(err, "创建时序数据集失败")
				return
			}
		}
		_, err = arb.arbMetricsTS.Do(ctx, "TS.CREATE", "sendSpentByJito", "RETENTION", (30 * gtime.D).Milliseconds(), "DUPLICATE_POLICY", "MAX")
		if err != nil {
			var redisErr redis.Error
			if errors.As(err, &redisErr) {
				if redisErr.Error() == "ERR TSDB: key already exists" {
					err = nil
					_, err = arb.arbMetricsTS.Do(ctx, "TS.ALTER", "sendSpentByJito", "RETENTION", (30 * gtime.D).Milliseconds(), "DUPLICATE_POLICY", "MAX")
					if err != nil {
						err = gerror.Wrapf(err, "更新时序数据集失败")
						return
					}
				}
			}
			if err != nil {
				err = gerror.Wrapf(err, "创建时序数据集失败")
				return
			}
		}
		_, err = arb.arbMetricsTS.Do(ctx, "TS.CREATE", "jitoRateLimitTimes", "RETENTION", (30 * gtime.D).Milliseconds(), "DUPLICATE_POLICY", "SUM")
		if err != nil {
			var redisErr redis.Error
			if errors.As(err, &redisErr) {
				if redisErr.Error() == "ERR TSDB: key already exists" {
					err = nil
					_, err = arb.arbMetricsTS.Do(ctx, "TS.ALTER", "jitoRateLimitTimes", "RETENTION", (30 * gtime.D).Milliseconds(), "DUPLICATE_POLICY", "SUM")
					if err != nil {
						err = gerror.Wrapf(err, "更新时序数据集失败")
						return
					}
				}
			}
			if err != nil {
				err = gerror.Wrapf(err, "创建时序数据集失败")
				return
			}
		}
	}

	return
}

func (arb *Arb) updateALT(ctx context.Context) {
	var err error
	defer func() {
		if err != nil {
			g.Log("scheduler").Errorf(ctx, "%+v", err)
		}
	}()

	alt, err := arb.officialPool.AddressLookupTableCacheGet(ctx, arb.alt.Address)
	if err != nil {
		err = gerror.Wrapf(err, "更新地址查找表失败")
		return
	}
	arb.alt = &alt
}

func (arb *Arb) fixTokenBlackLists(ctx context.Context) {
	var err error
	defer func() {
		if err != nil {
			g.Log("scheduler").Errorf(ctx, "%+v", err)
		}
	}()

	var rpcNames []string
	for cursor := uint64(0); ; {
		var keys []string
		cursor, keys, err = arb.tokenBlacklists.Scan(ctx, cursor, gredis.ScanOption{
			Type: "hash",
		})
		if err != nil {
			err = gerror.Wrapf(err, "获取代币黑名单 RPC 失败")
			return
		}

		rpcNames = append(rpcNames, keys...)
		if cursor == 0 {
			break
		}
	}

	for _, rpcName := range rpcNames {
		var tokenBlacklistM *gvar.Var
		tokenBlacklistM, err = arb.tokenBlacklists.HGetAll(ctx, rpcName)
		if err != nil {
			err = gerror.Wrapf(err, "获取代币黑名单失败")
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

		var ttls []int64
		ttls, err = Utils.RedisHPTtl(ctx, arb.tokenBlacklists, rpcName, needJudge...)
		if err != nil {
			err = gerror.Wrapf(err, "获取代币黑名单有效期失败")
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

		_, err = arb.tokenBlacklists.HDel(ctx, rpcName, needDelete...)
		if err != nil {
			err = gerror.Wrapf(err, "删除异常代币黑名单失败")
			return
		}
	}
}
