package review

import (
	"strconv"
	"strings"

	"pet/backend/internal/common"
)

func strconvI64(v int64) string { return strconv.FormatInt(v, 10) }

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, common.ErrParam
	}
	return id, nil
}

func trimSpace(s string) string { return strings.TrimSpace(s) }

func trimFloat(f float64) string { return strconv.FormatFloat(f, 'f', 1, 64) }
