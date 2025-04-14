package Address

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/binary"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gagliardetto/solana-go"
)

type AccountAddress struct {
	PublicKey solana.PublicKey
	base58    string // 用于加速 .String() 实现
}
type TokenAccountAddress struct {
	AccountAddress
}
type TokenAddress struct {
	AccountAddress
}
type TokenMetadataAddress struct {
	AccountAddress
}
type ProgramAddress struct {
	AccountAddress
}

type cachedAddress struct {
	PublicKey solana.PublicKey
}

func (address *AccountAddress) UnmarshalJSON(data []byte) (err error) {
	var cached cachedAddress
	err = json.Unmarshal(data, &cached)
	if err != nil {
		return
	}

	address_ := NewFromBytes32(cached.PublicKey)
	address.PublicKey = address_.PublicKey
	address.base58 = address_.base58

	return
}

func (address *AccountAddress) UnmarshalWithDecoder(dec *bin.Decoder) (err error) {
	err = dec.Decode(&address.PublicKey)
	if err != nil {
		err = gerror.Wrapf(err, "反序列化失败")
		return
	}

	address.base58 = address.PublicKey.String()

	return
}

func NewFromBase58(base58 string) (address AccountAddress) {
	address = AccountAddress{
		PublicKey: solana.MustPublicKeyFromBase58(base58),
		base58:    base58,
	}

	return
}

func NewsFromBase58(base58s []string) (addresses []AccountAddress) {
	for _, base58 := range base58s {
		addresses = append(addresses, AccountAddress{
			PublicKey: solana.MustPublicKeyFromBase58(base58),
			base58:    base58,
		})
	}

	return
}

func NewFromBytes32(bytes32 solana.PublicKey) (address AccountAddress) {
	address = AccountAddress{
		PublicKey: bytes32,
		base58:    bytes32.String(),
	}

	return
}

func NewsFromBytes32(bytes32s []solana.PublicKey) (addresses []AccountAddress) {
	for _, bytes32 := range bytes32s {
		addresses = append(addresses, AccountAddress{
			PublicKey: bytes32,
			base58:    bytes32.String(),
		})
	}

	return
}

func (address AccountAddress) String() string {
	return address.base58
}

func (address AccountAddress) ShortString() string {
	return address.String()[:5] + "..." + address.String()[len(address.String())-5:]
}

func (address AccountAddress) AsTokenAccountAddress() TokenAccountAddress {
	return TokenAccountAddress{
		AccountAddress: address,
	}
}

func (address AccountAddress) AsTokenAddress() TokenAddress {
	return TokenAddress{
		AccountAddress: address,
	}
}

func (address AccountAddress) AsTokenMetadataAddress() TokenMetadataAddress {
	return TokenMetadataAddress{
		AccountAddress: address,
	}
}

func (address AccountAddress) AsProgramAddress() ProgramAddress {
	return ProgramAddress{
		AccountAddress: address,
	}
}

func (address AccountAddress) Bytes() []byte {
	return address.PublicKey.Bytes()
}

func (address AccountAddress) Meta() *solana.AccountMeta {
	return solana.Meta(address.PublicKey)
}

func (address TokenAddress) FindTokenMetadataAddress() (tokenMetadataAddress TokenMetadataAddress, err error) {
	var tokenMetadataAddressPublicKey solana.PublicKey
	tokenMetadataAddressPublicKey, _, err = solana.FindTokenMetadataAddress(address.PublicKey)
	if err != nil {
		err = gerror.Wrapf(err, "计算代币信息地址失败")
		return
	}

	tokenMetadataAddress = TokenMetadataAddress{
		AccountAddress{
			PublicKey: tokenMetadataAddressPublicKey,
			base58:    tokenMetadataAddressPublicKey.String(),
		},
	}

	return
}

func (address ProgramAddress) FindProgramDerivedAddress(seeds [][]byte) (programAddress AccountAddress, bumpSeed uint8, err error) {
	var programAddressPublicKey solana.PublicKey
	programAddressPublicKey, bumpSeed, err = solana.FindProgramAddress(seeds, address.PublicKey)
	if err != nil {
		err = gerror.Wrapf(err, "计算程序派生地址失败")
		return
	}

	programAddress = AccountAddress{
		PublicKey: programAddressPublicKey,
		base58:    programAddressPublicKey.String(),
	}

	return
}

func (address ProgramAddress) CreateProgramDerivedAddress(seeds [][]byte, bump uint8) (programAddress AccountAddress, err error) {
	return address.CreateProgramAddress(append(seeds, []byte{bump}))
}

func (address ProgramAddress) CreateProgramAddress(seeds [][]byte) (programAddress AccountAddress, err error) {
	var programAddressPublicKey solana.PublicKey
	programAddressPublicKey, err = solana.CreateProgramAddress(seeds, address.PublicKey)
	if err != nil {
		err = gerror.Wrapf(err, "计算程序派生地址失败")
		return
	}

	programAddress = AccountAddress{
		PublicKey: programAddressPublicKey,
		base58:    programAddressPublicKey.String(),
	}

	return
}

func (address AccountAddress) FindAssociatedTokenAccountAddress(tokenAddress TokenAddress) (tokenAccountAddress TokenAccountAddress, err error) {
	var tokenAccountAddressPublicKey solana.PublicKey
	tokenAccountAddressPublicKey, _, err = solana.FindProgramAddress([][]byte{
		address.Bytes(),
		solana.TokenProgramID[:], // todo 常量
		tokenAddress.Bytes(),
	}, solana.SPLAssociatedTokenAccountProgramID) // todo 常量
	if err != nil {
		err = gerror.Wrapf(err, "计算代币账户派生地址失败")
		return
	}

	tokenAccountAddress = TokenAccountAddress{
		AccountAddress{
			PublicKey: tokenAccountAddressPublicKey,
			base58:    tokenAccountAddressPublicKey.String(),
		},
	}

	return
}

func (address AccountAddress) FindAssociatedToken2022AccountAddress(tokenAddress TokenAddress) (tokenAccountAddress TokenAccountAddress, err error) {
	var tokenAccountAddressPublicKey solana.PublicKey
	tokenAccountAddressPublicKey, _, err = solana.FindProgramAddress([][]byte{
		address.Bytes(),
		NewFromBase58("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb").AsProgramAddress().Bytes(), // todo 常量
		tokenAddress.Bytes(),
	}, solana.SPLAssociatedTokenAccountProgramID) // todo 常量
	if err != nil {
		err = gerror.Wrapf(err, "计算代币账户派生地址失败")
		return
	}

	tokenAccountAddress = TokenAccountAddress{
		AccountAddress{
			PublicKey: tokenAccountAddressPublicKey,
			base58:    tokenAccountAddressPublicKey.String(),
		},
	}

	return
}

func (address ProgramAddress) GetDiscriminator(namespace string, name string) (discriminator []byte) {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", namespace, name)))
	discriminator = hash[0:8]
	return
}
