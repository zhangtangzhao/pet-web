package common

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Body struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func JSON(w http.ResponseWriter, httpStatus, code int, msg string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(Body{Code: code, Msg: msg, Data: data})
}

func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, 0, "ok", data)
}

func Err(w http.ResponseWriter, err error) {
	var ae *ApiError
	if errors.As(err, &ae) {
		JSON(w, ae.HTTP, ae.Code, ae.Msg, nil)
		return
	}
	logx.Errorf("internal error: %v", err)
	JSON(w, ErrInternal.HTTP, ErrInternal.Code, ErrInternal.Msg, nil)
}

// ParseBody 解析 JSON 请求体
func ParseBody(r *http.Request, v any) error {
	if err := httpx.ParseJsonBody(r, v); err != nil {
		return ErrParam
	}
	return nil
}
