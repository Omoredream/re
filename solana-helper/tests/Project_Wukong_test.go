package tests

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gmutex"
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/imroc/req/v3"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

type steamUser struct {
	SteamId     string
	UrlId       string
	PersonaName string
}

func parseSteamCommentUsers(html string) (users []steamUser, err error) {
	matchedCommentUsers, err := gregex.MatchAllString(`<a href="https://steamcommunity\.com/(profiles|id)/(.+?)/">(.*?)</a>`, html)
	if err != nil {
		err = gerror.Wrapf(err, "解析评论失败")
		return
	}
	if len(matchedCommentUsers) == 0 {
		err = gerror.Newf("未解析评论")
		return
	}

	users = make([]steamUser, 0, len(matchedCommentUsers))
	for _, matchedCommentUser := range matchedCommentUsers {
		switch matchedCommentUser[1] {
		case "profiles":
			users = append(users, steamUser{
				SteamId:     matchedCommentUser[2],
				UrlId:       "",
				PersonaName: matchedCommentUser[3],
			})
		case "id":
			users = append(users, steamUser{
				SteamId:     "",
				UrlId:       matchedCommentUser[2],
				PersonaName: matchedCommentUser[3],
			})
		}
	}

	return
}

func parseSteamCommentPaginator(html string) (gameId string, kvs map[string]string, err error) {
	matchedForm, err := gregex.MatchString(`<form method="GET" id="MoreContentForm\d+" name="MoreContentForm\d+" action="https://steamcommunity\.com/app/(\d+)/homecontent/">(?:\W*<input type="hidden" name="(?:.+?)" value="(?:.*?)">\W*)+</form>`, html)
	if err != nil {
		err = gerror.Wrapf(err, "寻找翻页表单失败")
		return
	}
	if len(matchedForm) == 0 {
		err = gerror.Newf("未找到翻页表单")
		return
	}
	gameId = matchedForm[1]

	matchedFormInputs, err := gregex.MatchAllString(`<input type="hidden" name="(.+?)" value="(.*?)">`, matchedForm[0])
	if err != nil {
		err = gerror.Wrapf(err, "解析翻页表单失败")
		return
	}
	if len(matchedFormInputs) == 0 {
		err = gerror.Newf("未解析翻页表单")
		return
	}

	kvs = make(map[string]string, len(matchedFormInputs))
	for _, matchedFormInput := range matchedFormInputs {
		kvs[matchedFormInput[1]] = matchedFormInput[2]
	}

	return
}

func TestSteamWukong(f *testing.T) {
	//commentURL := "https://steamcommunity.com/app/2358720/reviews/?browsefilter=mostrecent&snr=1_5_100010_&filterLanguage=all&p=1"
	//commentURL := "https://steamcommunity.com/app/2358720/reviews/?browsefilter=trendyear&snr=1_5_100010_&filterLanguage=all&p=1"
	commentURL := "https://steamcommunity.com/app/2358720/reviews/?browsefilter=funny&snr=1_5_100010_&filterLanguage=all&p=1"

	c := req.C().
		ImpersonateChrome().
		SetCommonHeader("referer", commentURL)

	resp, err := c.R().
		Get(commentURL)
	if err != nil {
		f.Fatalf("%+v", err)
	}

	for {
		users, err := parseSteamCommentUsers(resp.String())
		if err != nil {
			f.Fatalf("%+v", err)
		}

		for _, user := range users {
			if user.SteamId == "" {
				resp, err := c.R().
					Get(fmt.Sprintf("https://steamcommunity.com/id/%s/", user.UrlId))
				if err != nil {
					f.Fatalf("%+v", err)
				}

				if gstr.Contains(resp.String(), "profile_private_info") { // 私密账号
					continue
				}

				if !gstr.Contains(resp.String(), "/games/?tab=all") { // 私密库存
					continue
				}

				matchedUserInfo, err := gregex.MatchString(`"steamid":"(\d+)"`, resp.String())
				if err != nil {
					f.Fatalf("%+v", err)
				}
				if len(matchedUserInfo) == 0 {
					f.Fatalf("未找到用户信息")
					return
				}

				user.SteamId = matchedUserInfo[1]
			} else if user.UrlId == "" {
				resp, err := c.R().
					Get(fmt.Sprintf("https://steamcommunity.com/profiles/%s/", user.SteamId))
				if err != nil {
					f.Fatalf("%+v", err)
				}

				if gstr.Contains(resp.String(), "profile_private_info") { // 私密账号
					continue
				}

				if !gstr.Contains(resp.String(), "/games/?tab=all") { // 私密库存
					continue
				}
			}

			f.Log(user.SteamId, user.UrlId, user.PersonaName)
			err = gfile.PutContentsAppend("./data/wukongSteamId.csv", user.SteamId+","+user.SteamId+"\n")
			if err != nil {
				g.Log().Fatalf(ctx, "%v", err)
			}
		}

		if len(users) < 10 {
			f.Log("数据不足")
			break
		}

		gameId, kvs, err := parseSteamCommentPaginator(resp.String())
		if err != nil {
			f.Fatalf("%+v", err)
		}

		resp, err = c.R().
			SetQueryParams(kvs).
			Get(fmt.Sprintf("https://steamcommunity.com/app/%s/homecontent/", gameId))
		if err != nil {
			f.Fatalf("%+v", err)
		}
	}
}

type wukong struct {
	programAddress      Address.ProgramAddress
	tokenAddress        Address.TokenAddress
	tokenAccountAddress Address.TokenAccountAddress
	ownerAccountAddress Address.AccountAddress
}

func newWukong(programAddress, tokenAddress, tokenAccountAddress string) (wk wukong, err error) {
	wk = wukong{
		programAddress:      Address.NewFromBase58(programAddress).AsProgramAddress(),
		tokenAddress:        Address.NewFromBase58(tokenAddress).AsTokenAddress(),
		tokenAccountAddress: Address.NewFromBase58(tokenAccountAddress).AsTokenAccountAddress(),
	}

	wk.ownerAccountAddress, _, err = wk.programAddress.FindProgramDerivedAddress([][]byte{[]byte("owner")})
	if err != nil {
		err = gerror.Wrapf(err, "生成程序管理账户失败")
		return
	}

	return
}

func (wk wukong) ClaimWithRecommendInstruction(wallet Wallet.HostedWallet, referer Address.TokenAccountAddress, claimData wukongClaimData) (ix solana.Instruction, err error) {
	creditAddress, _, err := wk.programAddress.FindProgramDerivedAddress([][]byte{[]byte("credit"), []byte(claimData.SteamId)})
	if err != nil {
		err = gerror.Wrapf(err, "生成已购校验账户失败")
		return
	}

	recipientAddress, _, err := wk.programAddress.FindProgramDerivedAddress([][]byte{[]byte("recipient"), wallet.Account.Address.Bytes()})
	if err != nil {
		err = gerror.Wrapf(err, "生成接收校验账户失败")
		return
	}

	// todo: 改为直接到大号
	//tokenAccountAddress, err := wallet.Account.Address.FindAssociatedTokenAccountAddress(wk.tokenAddress)
	//if err != nil {
	//	err = gerror.Wrapf(err, "生成代币账户失败")
	//	return
	//}

	ix, err = Instruction.Custom{
		ProgramID: wk.programAddress,
		Accounts: []*solana.AccountMeta{
			creditAddress.Meta().WRITE(),                   // credit
			recipientAddress.Meta().WRITE(),                // recipient
			wallet.Account.Address.Meta().SIGNER().WRITE(), // user
			wk.ownerAccountAddress.Meta(),                  // owner_account
			wk.tokenAccountAddress.Meta().WRITE(),          // from_account
			// todo: 改为直接到大号
			//tokenAccountAddress.Meta().WRITE(),             // to_account
			referer.Meta().WRITE(),             // to_account
			referer.Meta().WRITE(),             // to_recommend_account
			consts.TokenProgramAddress.Meta(),  // token_program
			consts.SystemProgramAddress.Meta(), // system_program
		},
		Discriminator: []byte{30, 173, 19, 154, 26, 235, 118, 214}, // claim_with_recommend
		Data: Utils.Append(
			Utils.StringToBytes(claimData.SteamId),
			claimData.Signature[:],
			Utils.Uint8ToBytesL(claimData.RecoveryId),
		),
	}.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	return
}

type wukongClaimData struct {
	SteamId    string   `json:"msg"`
	Signature  [64]byte `json:"signature"`
	RecoveryId byte     `json:"recoveryId"`
}

type wukongResp struct {
	SteamId    string `json:"msg"`
	Signature  string `json:"signature"`
	RecoveryId byte   `json:"recoveryId"`
}

type wukongRespError struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Error   string    `json:"error"`
	Path    string    `json:"path"`
	Date    time.Time `json:"date"`
}

func TestWukongDog(t *testing.T) {
	// mainnet
	wk, err := newWukong(
		"FmeK9A3A8zBDRXjL4JaFNQBAV2dVdeLCFQrVs3HYTGnE",
		"9mq7jWxKEwy8kPUweuRGn2cNSnFYY28BhHdZS6ebnbtr",
		"BQvLt59G6BjTGYU66VsppWaFb8Ymey4gwBsCuwbTUq1y",
	)
	// testnet
	//wk, err := newWukong(
	//	"HyAAZKhBmcN5B1jwxVMHMZTHKj96TKoZhvdmaTXfswd9",
	//	"732tykiXDDNp8iH7vru8hSsuofKGXPWiYLMJNmyR6f6Q",
	//	"DT4HPUZPB7FUN7wBc9r8WJgD5STC18W8ABKkggT84D1r",
	//)

	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		g.Log().Fatalf(ctx, "导入钱包失败, %+v", err)
	}

	mainTokenAccountAddress, err := mainWallet.Account.Address.FindAssociatedTokenAccountAddress(wk.tokenAddress)
	if err != nil {
		g.Log().Fatalf(ctx, "生成代币账户失败, %+v", err)
	}

	steamIds := make(chan string, 0xffffff)
	for steamId := range NewTargets(true, false).MustLoadFromCSV("./data/wukongSteamId.csv").
		Diff(NewTargets(true, true).MustLoadFromCSV("./data/wukongVerified.csv")).
		Diff(NewTargets(true, true).MustLoadFromCSV("./data/wukongError.csv")).
		Map() {
		steamIds <- steamId
	}

	claimData := make(chan wukongClaimData, 0xffffff)
	for steamId, signatureAndRecoveryId := range NewTargets(true, false).MustLoadFromCSV("./data/wukongVerified.csv").
		Diff(NewTargets(true, true).MustLoadFromCSV("./data/wukongUsed.csv")).
		Diff(NewTargets(true, true).MustLoadFromCSV("./data/wukongOtherUsed.csv")).
		Map() {
		signatureStr, recoveryIdStr := gstr.List2(signatureAndRecoveryId, "|")

		signature, err := hex.DecodeString(signatureStr)
		if err != nil {
			g.Log().Fatalf(ctx, "解析已验证数据失败, %+v", err)
		}

		recoveryId, err := strconv.ParseUint(recoveryIdStr, 16, 8)
		if err != nil {
			g.Log().Fatalf(ctx, "解析已验证数据失败, %+v", err)
		}

		claimData <- wukongClaimData{
			SteamId:    steamId,
			Signature:  [64]byte(signature),
			RecoveryId: byte(recoveryId),
		}
	}
	mu := &gmutex.RWMutex{}
	for i := 0; i < 20; i++ {
		go func() {
			client := req.C().
				ImpersonateChrome().
				SetBaseURL("https://wukongdog.com/api")
			for {
				steamId := <-steamIds

				//startTime := gtime.Now()

				resp, err := client.R().
					SetQueryParam("steamId", steamId).
					Get("/airdrop/verify")
				if err != nil {
					g.Log().Warningf(ctx, "发送验证已购请求失败, %+v", err)
					time.Sleep(1 * time.Second)
					continue
				}
				if resp.StatusCode != http.StatusOK {
					g.Log().Warningf(ctx, "服务器响应验证已购请求失败, HTTP %d, %s", resp.StatusCode, resp.Status)
					time.Sleep(1 * time.Second)
					continue
				}

				var errResult wukongRespError
				err = gjson.DecodeTo(resp.Bytes(), &errResult)
				if err != nil {
					g.Log().Fatalf(ctx, "解析验证已购响应失败, %s, %+v", resp.String(), err)
					time.Sleep(1 * time.Second)
					continue
				}

				if errResult.Code != 0 {
					if errResult.Code == 603 && errResult.Error == "CONDITION_NOT_MET" {
						mu.LockFunc(func() {
							err = gfile.PutContentsAppend("./data/wukongError.csv", steamId+","+steamId+"\n")
							if err != nil {
								g.Log().Fatalf(ctx, "%v", err)
							}
						})
						g.Log().Infof(ctx, "steam %s 无法验证", steamId)
						time.Sleep(1 * time.Second)
						continue
					}

					g.Log().Warningf(ctx, "服务器拒绝验证已购请求, errno %d, %s", errResult.Code, errResult.Message)
					time.Sleep(1 * time.Second)
					continue
				}

				var result wukongResp
				err = gjson.DecodeTo(resp.Bytes(), &result)
				if err != nil {
					g.Log().Fatalf(ctx, "解析验证已购响应失败, %s, %+v", resp.String(), err)
					time.Sleep(1 * time.Second)
					continue
				}

				signature, err := hex.DecodeString(result.Signature)
				if err != nil {
					g.Log().Fatalf(ctx, "解析验证数据失败, %+v", err)
				}

				verified := wukongClaimData{
					SteamId:    result.SteamId,
					Signature:  [64]byte(signature),
					RecoveryId: result.RecoveryId,
				}

				g.Log().Infof(ctx, "steam %s 已验证", steamId)

				mu.LockFunc(func() {
					err = gfile.PutContentsAppend("./data/wukongVerified.csv", hex.EncodeToString(verified.Signature[:])+"|"+strconv.FormatUint(uint64(verified.RecoveryId), 10)+","+verified.SteamId+"\n")
					if err != nil {
						g.Log().Fatalf(ctx, "%v", err)
					}
				})

				claimData <- verified

				//time.Sleep(startTime.Add(2 * time.Second).Sub(gtime.Now())) // 10次 / 20s
			}
		}()
	}

	//time.Sleep(24 * time.Hour)

	wg := sync.WaitGroup{}
	for fork := uint32(0); fork < 20; fork++ {
		wg.Add(1)
		fork := fork
		go func() {
			defer wg.Done()
			for i := fork*200 + 50; i < (fork+1)*200; i++ {
				ctx := context.WithValue(ctx, consts.CtxDerivation, i)

				altWallet, err := officialPool.NewWalletFromMnemonic(ctx, testWalletMnemonic, i)
				if err != nil {
					g.Log().Fatalf(ctx, "导入钱包失败, %+v", err)
				}

				ctx = context.WithValue(ctx, consts.CtxAddress, altWallet.Account.Address.String())

				if altWallet.Account.SOL.IsZero() {
					g.Log().Infof(ctx, "余额为零")
					continue
				}

				nextWallet, err := officialPool.NewWalletFromMnemonic(ctx, testWalletMnemonic, i+1)
				if err != nil {
					g.Log().Fatalf(ctx, "导入钱包失败, %+v", err)
				}
				if (i+1)%200 == 0 {
					nextWallet = mainWallet
				}

				ixs := make([]solana.Instruction, 0, 2)

				// todo: 改为直接到大号
				//_, err = Instruction.CreateTokenAccount{
				//	Funder:  altWallet.Account.Address,
				//	Owner:   altWallet.Account.Address,
				//	Token:   wk.tokenAddress,
				//}.AppendIx(&ixs)
				//if err != nil {
				//	g.Log().Fatalf(ctx, "生成代币账户失败, %+v", err)
				//}

				verified := <-claimData

				claimIx, err := wk.ClaimWithRecommendInstruction(altWallet, mainTokenAccountAddress, verified)
				if err != nil {
					g.Log().Fatalf(ctx, "生成 claim 指令失败, %+v", err)
				}
				ixs = append(ixs, claimIx)

				err = Instruction.Transfer{
					Sender:   altWallet,
					Receiver: nextWallet.Account.Address,
					Amount:   altWallet.Account.SOL.Sub(lamports.Lamports2SOL(2568240 + 1865280 + 5000)),
				}.AppendIx(&ixs)
				if err != nil {
					g.Log().Fatalf(ctx, "生成转账指令失败, %+v", err)
				}

				txHash, err := officialPool.SendInstructions(ctx, ixs, []Wallet.HostedWallet{altWallet}, altWallet)
				if err != nil {
					g.Log().Fatalf(ctx, "发送交易失败, %+v", err)
				}
				ctx = context.WithValue(ctx, consts.CtxTransaction, txHash.String())
				g.Log().Infof(ctx, "已发送交易")

				spent, err := officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
				if err != nil {
					if err.Error() == "链上错误: map[InstructionError:[0 map[Custom:0]]]" {
						i--

						g.Log().Infof(ctx, "其他人已使用")

						mu.LockFunc(func() {
							err := gfile.PutContentsAppend("./data/wukongOtherUsed.csv", hex.EncodeToString(verified.Signature[:])+"|"+strconv.FormatUint(uint64(verified.RecoveryId), 10)+","+verified.SteamId+"\n")
							if err != nil {
								g.Log().Fatalf(ctx, "%v", err)
							}
						})

						continue
					} else if err.Error() == "链上错误: map[InstructionError:[1 map[Custom:1]]]" {
						i--

						g.Log().Infof(ctx, "转账余额不足")

						claimData <- verified

						continue
					}
					g.Log().Fatalf(ctx, "等待交易确认失败, %+v", err)
				}
				g.Log().Infof(ctx, "交易耗时 %s", spent)

				mu.LockFunc(func() {
					err := gfile.PutContentsAppend("./data/wukongUsed.csv", hex.EncodeToString(verified.Signature[:])+"|"+strconv.FormatUint(uint64(verified.RecoveryId), 10)+","+verified.SteamId+"\n")
					if err != nil {
						g.Log().Fatalf(ctx, "%v", err)
					}
				})

				time.Sleep(2 * time.Second)
			}
		}()
	}
	wg.Wait()
}
