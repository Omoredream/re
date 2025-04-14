package jitoTypes

import (
	"github.com/gagliardetto/solana-go/rpc"
)

type BundleStatus string

const (
	BundleInvalid BundleStatus = "Invalid" // 捆绑 ID 不在我们的系统中 (5 分钟回顾)
	BundlePending BundleStatus = "Pending" // 未失败、未落地、未无效
	BundleFailed  BundleStatus = "Failed"  // 所有收到该捆绑包的区域均已将其标记为失败, 且尚未转发
	BundleLanded  BundleStatus = "Landed"  // 已上链
)

type GetBundleStatusesResponse struct {
	BundleId           string             `json:"bundle_id"`
	TxHashes           []string           `json:"transactions"`
	Slot               uint64             `json:"slot"`
	ConfirmationStatus rpc.CommitmentType `json:"confirmation_status"`
	Err                any                `json:"err"`
}

type GetInflightBundleStatusesResponse struct {
	BundleId   string       `json:"bundle_id"`
	Status     BundleStatus `json:"status"`
	LandedSlot *uint64      `json:"landed_slot"`
}
