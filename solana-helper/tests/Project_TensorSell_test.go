package tests

import (
	"crypto/ed25519"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gagliardetto/binary"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gmutex"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"github.com/shopspring/decimal"

	"github.com/blocto/solana-go-sdk/pkg/hdwallet"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/gagliardetto/solana-go"
	"github.com/tyler-smith/go-bip39"

	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

type GraphQLQuery struct {
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
	Query         string         `json:"query"`
}

const tensorGraphQLCollectionMintsV2Query = `
query CollectionMintsV2($slug: String!, $sortBy: CollectionMintsSortBy!, $filters: CollectionMintsFilters, $cursor: String, $limit: Int) {
  collectionMintsV2(
    slug: $slug
    sortBy: $sortBy
    filters: $filters
    cursor: $cursor
    limit: $limit
  ) {
    mints {
      ...MintWithTx
      mint {
        numMints
        __typename
      }
      __typename
    }
    page {
      endCursor
      hasMore
      __typename
    }
    __typename
  }
}

fragment MintWithTx on MintWithTx {
  mint {
    ...MintV2
    __typename
  }
  tx {
    ...ReducedParsedTx
    __typename
  }
  __typename
}

fragment MintV2 on MintV2 {
  onchainId
  slug
  compressed
  owner
  name
  imageUri
  animationUri
  metadataUri
  metadataFetchedAt
  files {
    type
    uri
    __typename
  }
  sellRoyaltyFeeBPS
  tokenStandard
  tokenEdition
  attributes {
    trait_type
    value
    __typename
  }
  lastSale {
    price
    txAt
    __typename
  }
  accState
  hidden
  rarityRankHrtt
  rarityRankStat
  rarityRankTeam
  rarityRankTn
  inscription {
    ...InscriptionData
    __typename
  }
  tokenProgram
  metadataProgram
  transferHookProgram
  listingNormalizedPrice
  hybridAmount
  __typename
}

fragment InscriptionData on InscriptionData {
  inscription
  inscriptionData
  immutable
  order
  spl20 {
    p
    tick
    amt
    __typename
  }
  __typename
}

fragment ReducedParsedTx on ParsedTransaction {
  source
  txKey
  txId
  txType
  grossAmount
  sellerId
  buyerId
  txAt
  blockNumber
  txMetadata {
    auctionHouse
    urlId
    sellerRef
    tokenAcc
    __typename
  }
  poolOnchainId
  lockOnchainId
  __typename
}
`

type TensorGraphQLCollectionMintsV2ResultCollectionMintsV2 struct {
	Mints []struct {
		Mint struct {
			OnchainId              string      `json:"onchainId"`
			Slug                   uuid.UUID   `json:"slug"`
			Compressed             bool        `json:"compressed"`
			Owner                  string      `json:"owner"`
			Name                   string      `json:"name"`
			ImageUri               string      `json:"imageUri"`
			AnimationUri           *string     `json:"animationUri"`
			MetadataUri            string      `json:"metadataUri"`
			MetadataFetchedAt      *gtime.Time `json:"metadataFetchedAt"`
			Files                  any         `json:"files"`
			SellRoyaltyFeeBPS      int         `json:"sellRoyaltyFeeBPS"`
			TokenStandard          string      `json:"tokenStandard"`
			TokenEdition           any         `json:"tokenEdition"`
			Attributes             []any       `json:"attributes"`
			LastSale               any         `json:"lastSale"`
			AccState               any         `json:"accState"`
			Hidden                 bool        `json:"hidden"`
			RarityRankHrtt         any         `json:"rarityRankHrtt"`
			RarityRankStat         any         `json:"rarityRankStat"`
			RarityRankTeam         any         `json:"rarityRankTeam"`
			RarityRankTn           any         `json:"rarityRankTn"`
			Inscription            any         `json:"inscription"`
			TokenProgram           any         `json:"tokenProgram"`
			MetadataProgram        any         `json:"metadataProgram"`
			TransferHookProgram    any         `json:"transferHookProgram"`
			ListingNormalizedPrice any         `json:"listingNormalizedPrice"`
			HybridAmount           any         `json:"hybridAmount"`
			Typename               string      `json:"__typename"` // MintV2
			NumMints               int         `json:"numMints"`
		} `json:"mint"`
		Tx       any    `json:"tx"`
		Typename string `json:"__typename"` // MintWithTx
	} `json:"mints"`
	Page struct {
		EndCursor string `json:"endCursor"`
		HasMore   bool   `json:"hasMore"`
		Typename  string `json:"__typename"` // CollectionMintsV2Page
	} `json:"page"`
	Typename string `json:"__typename"` // CollectionMintsV2
}

const tensorGraphQLSwapOrdersQuery = `
query SwapOrders($slug: String!, $owner: String) {
  tswapOrders(slug: $slug, owner: $owner) {
    ...ReducedTSwapPool
    __typename
  }
  hswapOrders(slug: $slug, owner: $owner) {
    ...ReducedHSwapPool
    __typename
  }
  tcompBids(slug: $slug, owner: $owner) {
    ...ReducedTCompBid
    __typename
  }
}

fragment ReducedTSwapPool on TSwapPool {
  address
  ownerAddress
  whitelistAddress
  poolType
  curveType
  startingPrice
  delta
  mmCompoundFees
  mmFeeBalance
  mmFeeBps
  takerSellCount
  takerBuyCount
  nftsHeld
  solBalance
  createdUnix
  statsTakerSellCount
  statsTakerBuyCount
  statsAccumulatedMmProfit
  margin
  marginNr
  lastTransactedAt
  maxTakerSellCount
  nftsForSale {
    ...ReducedMint
    __typename
  }
  __typename
}

fragment ReducedMint on TLinkedTxMintTV2 {
  onchainId
  compressed
  owner
  name
  imageUri
  animationUri
  metadataUri
  metadataFetchedAt
  files {
    type
    uri
    __typename
  }
  sellRoyaltyFeeBPS
  tokenStandard
  tokenEdition
  attributes {
    trait_type
    value
    __typename
  }
  lastSale {
    price
    txAt
    __typename
  }
  accState
  hidden
  ...MintRarityFields
  staked {
    stakedAt
    activatedAt
    stakedByOwner
    __typename
  }
  inscription {
    ...InscriptionData
    __typename
  }
  tokenProgram
  metadataProgram
  transferHookProgram
  listingNormalizedPrice
  hybridAmount
  __typename
}

fragment MintRarityFields on TLinkedTxMintTV2 {
  rarityRankHrtt
  rarityRankStat
  rarityRankTeam
  rarityRankTn
  __typename
}

fragment InscriptionData on InscriptionData {
  inscription
  inscriptionData
  immutable
  order
  spl20 {
    p
    tick
    amt
    __typename
  }
  __typename
}

fragment ReducedHSwapPool on HSwapPool {
  address
  pairType
  delta
  curveType
  baseSpotPrice
  feeBps
  mathCounter
  assetReceiver
  boxes {
    address
    vaultTokenAccount
    mint {
      ...ReducedMint
      __typename
    }
    __typename
  }
  feeBalance
  buyOrdersQuantity
  fundsSolOrTokenBalance
  createdAt
  lastTransactedAt
  __typename
}

fragment ReducedTCompBid on TCompBid {
  address
  target
  targetId
  field
  fieldId
  amount
  solBalance
  ownerAddress
  filledQuantity
  quantity
  margin
  marginNr
  createdAt
  attributes {
    trait_type
    value
    __typename
  }
  __typename
}
`

type TensorGraphQLSwapOrdersResultTcompBid struct {
	Address        string      `json:"address"`
	Target         string      `json:"target"`
	TargetId       string      `json:"targetId"`
	Field          any         `json:"field"`
	FieldId        any         `json:"fieldId"`
	Amount         uint64      `json:"amount,string"`
	SolBalance     uint64      `json:"solBalance,string"`
	OwnerAddress   string      `json:"ownerAddress"`
	FilledQuantity int         `json:"filledQuantity"`
	Quantity       int         `json:"quantity"`
	Margin         any         `json:"margin"`
	MarginNr       any         `json:"marginNr"`
	CreatedAt      *gtime.Time `json:"createdAt"`
	Attributes     any         `json:"attributes"`
	Typename       string      `json:"__typename"` // TCompBid
}

const tensorGraphQLTcompTakeBidTxQuery = `
query TcompTakeBidTx($minPrice: Decimal!, $seller: String!, $mint: String!, $bidStateAddress: String, $optionalRoyaltyPct: Int, $buyer: String, $priorityMicroLamports: Int!, $blockhash: String) {
  tcompTakeBidTx(
    minPrice: $minPrice
    seller: $seller
    mint: $mint
    bidStateAddress: $bidStateAddress
    optionalRoyaltyPct: $optionalRoyaltyPct
    buyer: $buyer
    priorityMicroLamports: $priorityMicroLamports
    blockhash: $blockhash
  ) {
    txs {
      ...TxResponse
      __typename
    }
    __typename
  }
}

fragment TxResponse on OnchainTx {
  tx
  txV0
  lastValidBlockHeight
  metadata
  __typename
}
`

type TensorGraphQLTcompTakeBidTxResultTcompTakeBidTx struct {
	Txs []struct {
		Tx   any `json:"tx"`
		TxV0 struct {
			Type string `json:"type"`
			Data []byte `json:"data"`
		} `json:"txV0"`
		LastValidBlockHeight any    `json:"lastValidBlockHeight"`
		Metadata             any    `json:"metadata"`
		Typename             string `json:"__typename"` // OnchainTx
	} `json:"txs"`
	Typename string `json:"__typename"` // TCompBidTxResponse
}

func TestTensorSell(t *testing.T) {
	const nftSlug = "44b31dfe-cf2b-4da5-8dd6-34c859f9d75c"

	var mintWallets []Wallet.HostedWallet

	baseDerivationPath, err := accounts.ParseDerivationPath("m/44'/501'")
	if err != nil {
		g.Log().Fatalf(ctx, "生成派生路径失败, %+v", err)
	}

	var (
		extendedKey    hdwallet.Key
		derivativeFrom int
		derivativeTo   int
	)

	extendedKey = hdwallet.CreateMasterKey(bip39.NewSeed(testWalletMnemonic, ""))
	for _, n := range baseDerivationPath {
		extendedKey = hdwallet.CKDPriv(extendedKey, n)
	}
	derivativeFrom, derivativeTo = 0, 3 // 改这里, 助记词派生出来的钱包序号, 从0开始
	mintWallets = Utils.Grow(mintWallets, len(mintWallets)+derivativeTo-derivativeFrom)
	for derivation := derivativeFrom; derivation < derivativeTo; derivation++ {
		addressDerivationPath, err := accounts.ParseDerivationPath(fmt.Sprintf("m/%d'/0'", derivation))
		if err != nil {
			g.Log().Fatalf(ctx, "派生钱包失败, %+v", err)
		}
		privateKey := ed25519.NewKeyFromSeed(hdwallet.CKDPriv(hdwallet.CKDPriv(extendedKey, addressDerivationPath[0]), addressDerivationPath[1]).PrivateKey)
		mintWallet, err := officialPool.NewWalletFromPrivateKey(ctx, privateKey)
		if err != nil {
			g.Log().Fatalf(ctx, "派生钱包失败, %+v", err)
		}
		g.Log().Debugf(ctx, "(%d) %s: SOL %s", derivation, mintWallet.Account.Address.String(), decimals.DisplayBalance(mintWallet.Account.SOL))
		mintWallets = append(mintWallets, mintWallet)
	}

	client := req.C().
		ImpersonateChrome()

	for _, mintWallet := range mintWallets {
		for ; ; time.Sleep(30 * time.Second) {
			resp, err := client.R().
				SetBodyJsonMarshal([]GraphQLQuery{
					{
						OperationName: "CollectionMintsV2",
						Variables: g.MapStrAny{
							"slug":   nftSlug,
							"sortBy": "ListingPriceDesc",
							"filters": g.MapStrAny{
								"ownerFilter": g.MapStrAny{
									"include": g.ArrayStr{
										mintWallet.Account.Address.String(),
									},
								},
							},
							"limit": 50,
						},
						Query: tensorGraphQLCollectionMintsV2Query,
					},
					{
						OperationName: "SwapOrders",
						Variables: g.MapStrAny{
							"slug":  nftSlug,
							"owner": mintWallet.Account.Address.String(),
						},
						Query: tensorGraphQLSwapOrdersQuery,
					},
				}).
				Post("https://graphql.tensor.trade/graphql")
			if err != nil {
				g.Log().Fatalf(ctx, "%+v", err)
			}
			if resp.StatusCode != http.StatusOK {
				g.Log().Fatalf(ctx, "HTTP/%d: %s", resp.StatusCode, resp.Status)
			}

			respJ, err := gjson.DecodeToJson(resp.Bytes())
			if err != nil {
				g.Log().Fatalf(ctx, "%+v", err)
			}

			if !respJ.Contains("0.data.collectionMintsV2") {
				g.Log().Fatalf(ctx, "响应非预期: %s", respJ)
			}
			var collectionMintsV2 TensorGraphQLCollectionMintsV2ResultCollectionMintsV2
			err = respJ.Get("0.data.collectionMintsV2").Scan(&collectionMintsV2)
			if err != nil {
				g.Log().Fatalf(ctx, "%+v", err)
			}

			if len(collectionMintsV2.Mints) == 0 {
				g.Log().Infof(ctx, "无 NFT 可卖")
				break
			}

			if !respJ.Contains("1.data.tcompBids") {
				g.Log().Fatalf(ctx, "响应非预期: %s", respJ)
			}
			var tcompBids []TensorGraphQLSwapOrdersResultTcompBid
			err = respJ.Get("1.data.tcompBids").Scan(&tcompBids)
			if err != nil {
				g.Log().Fatalf(ctx, "%+v", err)
			}

			slices.SortFunc(tcompBids, func(a, b TensorGraphQLSwapOrdersResultTcompBid) int {
				if a.Amount < b.Amount { // 贵的排前面
					return 1
				} else if a.Amount > b.Amount {
					return -1
				} else {
					if a.Quantity-a.FilledQuantity < b.Quantity-b.FilledQuantity { // 多的排前面
						return 1
					} else if a.Quantity-a.FilledQuantity > b.Quantity-b.FilledQuantity {
						return -1
					} else {
						return 0
					}
				}
			})

			blockhash, err := officialPool.GetLatestBlockhash(ctx)
			if err != nil {
				g.Log().Fatalf(ctx, "%+v", err)
			}

			var queries []GraphQLQuery
			j := 0
			for i := range collectionMintsV2.Mints {
				for {
					if len(tcompBids) <= j {
						break
					}
					if tcompBids[j].FilledQuantity >= tcompBids[j].Quantity {
						j++
						continue
					}
					break
				}
				if len(tcompBids) <= j {
					break
				}
				g.Log().Infof(ctx, "单价: SOL %s", decimals.DisplayBalance(lamports.Lamports2SOL(tcompBids[j].Amount)))
				if tcompBids[j].Amount < lamports.SOL2Lamports(decimal.NewFromFloat(0.0015)) {
					break
				}
				queries = append(queries, GraphQLQuery{
					OperationName: "TcompTakeBidTx",
					Variables: g.MapStrAny{
						"mint":                  collectionMintsV2.Mints[i].Mint.OnchainId,
						"minPrice":              strconv.Itoa(int(tcompBids[j].Amount)),
						"seller":                mintWallet.Account.Address.String(),
						"optionalRoyaltyPct":    100, // todo: 不知道怎么处理的版税
						"buyer":                 tcompBids[j].OwnerAddress,
						"bidStateAddress":       tcompBids[j].Address,
						"blockhash":             blockhash.String(),
						"priorityMicroLamports": 0,
					},
					Query: tensorGraphQLTcompTakeBidTxQuery,
				})
				tcompBids[j].FilledQuantity++
			}

			if len(queries) == 0 {
				g.Log().Infof(ctx, "无交易")
				continue
			}

			resp, err = client.R().
				SetBodyJsonMarshal(queries).
				Post("https://graphql-txs.tensor.trade/graphql")
			if err != nil {
				g.Log().Fatalf(ctx, "%+v", err)
			}
			if resp.StatusCode != http.StatusOK {
				g.Log().Fatalf(ctx, "HTTP/%d: %s", resp.StatusCode, resp.Status)
			}

			respJ, err = gjson.DecodeToJson(resp.Bytes())
			if err != nil {
				g.Log().Fatalf(ctx, "%+v", err)
			}

			txHashs := make([]solana.Signature, 0)

			wg := sync.WaitGroup{}
			mu := gmutex.Mutex{}
			for _, respJ := range respJ.GetJsons(".") {
				if !respJ.Contains("data.tcompTakeBidTx") {
					g.Log().Fatalf(ctx, "响应非预期: %s", respJ)
				}
				var tcompTakeBidTx TensorGraphQLTcompTakeBidTxResultTcompTakeBidTx
				err = respJ.Get("data.tcompTakeBidTx").Scan(&tcompTakeBidTx)
				if err != nil {
					g.Log().Fatalf(ctx, "%+v", err)
				}

				if len(tcompTakeBidTx.Txs) > 1 {
					g.Log().Fatalf(ctx, "返回交易条数非预期: %v", tcompTakeBidTx)
				}

				if tcompTakeBidTx.Txs[0].TxV0.Type != "Buffer" {
					g.Log().Fatalf(ctx, "返回交易格式非预期: %v", tcompTakeBidTx.Txs[0].TxV0)
				}

				tx := &solana.Transaction{}
				err = tx.UnmarshalWithDecoder(bin.NewBinDecoder(tcompTakeBidTx.Txs[0].TxV0.Data))
				if err != nil {
					g.Log().Fatalf(ctx, "解析交易失败, %+v", err)
				}

				wg.Add(1)
				go func() {
					defer wg.Done()

					err := officialPool.SignTransaction(ctx, tx, []Wallet.HostedWallet{mintWallet})
					if err != nil {
						g.Log().Errorf(ctx, "签名交易失败, %+v", err)
						return
					}

					txHash, err := officialPool.SendTransaction(ctx, tx)
					if err != nil {
						g.Log().Errorf(ctx, "发送交易失败, %+v", err)
						return
					}

					g.Log().Infof(ctx, "交易ID: %s", txHash)
					mu.LockFunc(func() {
						txHashs = append(txHashs, txHash)
					})
				}()
			}
			wg.Wait()

			success, failed, all := 0, 0, len(txHashs)
			wg = sync.WaitGroup{}
			mu = gmutex.Mutex{}
			for _, txHash := range txHashs {
				wg.Add(1)
				txHash := txHash
				go func() {
					defer wg.Done()

					spent, err := officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
					if err != nil {
						mu.LockFunc(func() {
							failed++
						})
						g.Log().Errorf(ctx, "等待交易确认失败, %d (%d) / %d, %+v", success+failed, success, all, err)
						return
					}

					mu.LockFunc(func() {
						success++
					})
					g.Log().Infof(ctx, "交易 %s 耗时 %s, %d (%d) / %d", txHash, spent, success+failed, success, all)
				}()
			}
			wg.Wait()
		}
	}
}
