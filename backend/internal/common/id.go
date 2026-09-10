package common

import (
	"github.com/bwmarrin/snowflake"
)

var idNode *snowflake.Node

func InitIDNode(nodeId int64) error {
	n, err := snowflake.NewNode(nodeId)
	if err != nil {
		return err
	}
	idNode = n
	return nil
}

func NewID() int64 {
	return idNode.Generate().Int64()
}

func NewBizNo(prefix string) string {
	return prefix + idNode.Generate().String()
}

// UIDFromClaims 从 JWT claims 取用户 ID
func UIDFromClaims(claims map[string]any) int64 {
	switch v := claims["uid"].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	default:
		return 0
	}
}
