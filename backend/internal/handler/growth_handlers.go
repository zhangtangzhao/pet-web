// growth_handlers.go 用户端新增能力：地址簿、积分/签到、邀请、秒杀列表、演示支付。
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/growth"
	"pet/backend/internal/logic/trade"
	"pet/backend/internal/metrics"
	"pet/backend/internal/model"
	"pet/backend/internal/ratelimit"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func idStr(id int64) string { return strconv.FormatInt(id, 10) }

func AddressList(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		list, err := trade.AddressList(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	})
}

func AddressSave(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.AddressSaveReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := trade.AddressSave(sc, memberID(r), &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AddressSetDefault(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.AddressIDReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := trade.AddressSetDefault(sc, memberID(r), path.ID); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AddressDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.AddressIDReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := trade.AddressDelete(sc, memberID(r), path.ID); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func SignIn(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		balance, ok, err := growth.SignIn(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, &types.SignInResp{OK: ok, Balance: balance})
	})
}

func MyPoints(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.PointsPageReq
		_ = httpx.ParseForm(r, &req)
		resp, err := growth.MyPoints(sc, memberID(r), req.Page, req.PageSize)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func Invite(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		resp, err := growth.InviteSummary(sc, memberID(r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// FlashSales 生效中的秒杀活动（公开，首页/商品页展示秒杀价）
func FlashSales(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rows []model.FlashSale
		if err := sc.DB.
			Where("status = ? AND start_at <= now() AND end_at > now() AND sold < stock", model.FlashSaleOn).
			Order("start_at ASC").Find(&rows).Error; err != nil {
			common.Err(w, err)
			return
		}
		// 主图/标题快照（首页秒杀专区直接展示，避免前端逐个查详情）
		var products []model.PetProduct
		productIDs := make([]int64, 0, len(rows))
		for _, fs := range rows {
			productIDs = append(productIDs, fs.ProductID)
		}
		prodMap := map[int64]model.PetProduct{}
		if len(productIDs) > 0 {
			if err := sc.DB.Where("id IN ?", productIDs).Find(&products).Error; err == nil {
				for _, p := range products {
					prodMap[p.ID] = p
				}
			}
		}
		list := make([]types.FlashSaleView, 0, len(rows))
		for _, fs := range rows {
			v := types.FlashSaleView{
				ID:        idStr(fs.ID),
				ProductID: idStr(fs.ProductID),
				SalePrice: fs.SalePrice.StringFixed(2),
				Stock:     fs.Stock,
				Sold:      fs.Sold,
				StartAt:   fs.StartAt.Format(time.RFC3339),
				EndAt:     fs.EndAt.Format(time.RFC3339),
				Status:    fs.Status,
			}
			if p, ok := prodMap[fs.ProductID]; ok {
				v.ProductTitle = p.Title
				v.ProductImage = p.MainImage
			}
			list = append(list, v)
		}
		common.OK(w, list)
	}
}

// MockPay 开发演示支付：未配置商户号且非生产时直接落账（生产 404）
func MockPay(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.OrderNoPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if !ratelimit.Allow(sc.Rdb, "rl:mockpay:"+idStr(memberID(r)), 30, time.Minute) {
			metrics.LimitRejected.WithLabelValues("mockpay").Inc()
			common.Err(w, common.ErrTooManyRequests)
			return
		}
		if err := trade.MockPay(sc, memberID(r), path.OrderNo); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}
