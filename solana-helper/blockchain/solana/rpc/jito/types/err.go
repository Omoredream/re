package jitoTypes

import (
	"github.com/gogf/gf/v2/errors/gcode"
)

var (
	ErrBlockhashExpired               = gcode.New(-32602, "签名区块过期", nil)
	ErrTransactionAlreadyProcessed    = gcode.New(-32602, "交易已处理过", nil)
	ErrTransactionAccountLocksTooMany = gcode.New(-32603, "交易锁定过多账户", nil)
	ErrRateLimited                    = gcode.New(-32097, "频率限制", nil)
)
