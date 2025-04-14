package rpcinterface

import (
	"time"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gmutex"
	"github.com/gogf/gf/v2/os/gtime"

	"git.wkr.moe/web3/solana-helper/errcode"
)

type RPCInterface interface {
	Name() string
	Fingerprint() string
	Success(spendTime time.Duration)
	Fail(err error)
	Weight() float64
	IsCoolDown() bool
	AddCoolDown(cooldownIntervalMill int64)
	RunMoreThread() bool
	EndThread()
}

type RPC struct {
	name                 string
	cooldownIntervalMill int64
	maxRunningThreads    uint

	statusMutex         *gmutex.RWMutex
	spendTime           uint64
	succeedTimes        uint64
	failedTimes         uint64
	cooldown            int64
	cooldownExtra       int64
	weight              float64
	runningThreads      uint
	runningThreadsMutex *gmutex.Mutex
}

func New(name string, cooldownIntervalMill int64, maxRunningThreads uint) (node *RPC) {
	node = &RPC{
		name:                 name,
		cooldownIntervalMill: cooldownIntervalMill,
		maxRunningThreads:    maxRunningThreads,

		statusMutex:         &gmutex.RWMutex{},
		runningThreadsMutex: &gmutex.Mutex{},
	}
	return
}

func (node *RPC) Name() string {
	return node.name
}

func (node *RPC) Fingerprint() string {
	return gmd5.MustEncryptString(node.name)
}

func (node *RPC) Success(spendTime time.Duration) {
	node.statusMutex.Lock()
	defer node.statusMutex.Unlock()

	node.cooldown = 0
	node.spendTime += uint64(spendTime.Milliseconds())
	node.succeedTimes++
	node.calcWeight()
}

func (node *RPC) Fail(err error) {
	node.statusMutex.Lock()
	defer node.statusMutex.Unlock()

	if gerror.HasCode(err, errcode.CoolDownLessError) {
		return
	}
	node.cooldown = max(node.cooldown, gtime.TimestampMilli()+node.cooldownIntervalMill)
	node.failedTimes++
	node.calcWeight()
}

func (node *RPC) Weight() float64 {
	node.statusMutex.RLock()
	defer node.statusMutex.RUnlock()

	return node.weight
}

func (node *RPC) calcWeight() {
	spendTime := float64(node.spendTime) / float64(node.succeedTimes)                       // 单次成功请求平均耗时 ms/t
	succeedRate := float64(node.succeedTimes) / float64(node.succeedTimes+node.failedTimes) // 成功率
	node.weight = spendTime / succeedRate
}

func (node *RPC) IsCoolDown() bool {
	node.statusMutex.RLock()
	defer node.statusMutex.RUnlock()

	return gtime.TimestampMilli() <= node.cooldown || gtime.TimestampMilli() <= node.cooldownExtra
}

func (node *RPC) AddCoolDown(cooldownIntervalMill int64) {
	node.statusMutex.Lock()
	defer node.statusMutex.Unlock()

	node.cooldownExtra = max(node.cooldownExtra, gtime.TimestampMilli()+cooldownIntervalMill)
}

func (node *RPC) RunMoreThread() bool {
	node.runningThreadsMutex.Lock()
	defer node.runningThreadsMutex.Unlock()

	if node.runningThreads < node.maxRunningThreads {
		node.runningThreads++
		return true
	} else {
		return false
	}
}

func (node *RPC) EndThread() {
	node.runningThreadsMutex.Lock()
	defer node.runningThreadsMutex.Unlock()

	if node.runningThreads <= 0 {
		panic("存在过多的并发线程计数调用")
	}
	node.runningThreads--
}
