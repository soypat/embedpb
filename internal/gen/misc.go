package gen

import (
	"go/constant"
	"go/types"
)

func constToInt64(c *types.Const) (int64, bool) {
	return constant.Int64Val(c.Val())
}
