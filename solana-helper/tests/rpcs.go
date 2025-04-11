package tests

import (
	"github.com/gogf/gf/v2/frame/g"
)

var (
	Region  = g.Cfg().MustGet(nil, "rpcs.region").String()
	自适应     = Region == ""
	荷兰阿姆斯特丹 = Region == "ams"
	德国法兰克福  = Region == "fra"
	美国纽约    = Region == "ny"
	日本东京    = Region == "tyo"
	美国盐湖城   = Region == "slc"
)
