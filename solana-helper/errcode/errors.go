package errcode

import (
	"github.com/gogf/gf/v2/errors/gcode"
)

const (
	FatalErrorCode = iota + 1001
	PreemptedErrorCode
	NeedReconnectErrorCode
	IgnoreErrorCode
)

var (
	FatalError         = gcode.New(FatalErrorCode, "不可重试的错误", nil)
	PreemptedError     = gcode.New(PreemptedErrorCode, "被抢占的错误", nil)
	NeedReconnectError = gcode.New(NeedReconnectErrorCode, "需要发起重连的错误", nil)
	IgnoreError        = gcode.New(IgnoreErrorCode, "无需输出的错误", nil)
	CoolDownError      = gcode.New(IgnoreErrorCode, "需要冷却的错误", nil)
)
