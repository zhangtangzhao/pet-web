// play_handlers.go 砍价/竞拍/任务/VIP/预约/发票/分群/定价建议。
package handler

import (
	"net/http"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/auth"
	"pet/backend/internal/logic/growth"
	"pet/backend/internal/logic/manage"
	"pet/backend/internal/logic/marketing"
	"pet/backend/internal/logic/trade"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func memberH[T any](sc *svc.ServiceContext, needBody bool, fn func(int64, *T) (any, error)) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req T
		if needBody {
			if err := common.ParseBody(r, &req); err != nil {
				common.Err(w, err)
				return
			}
		}
		data, err := fn(memberID(r), &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, data)
	})
}

func adminH[T any](sc *svc.ServiceContext, needBody bool, fn func(*T) (any, error)) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req T
		if needBody {
			if err := common.ParseBody(r, &req); err != nil {
				common.Err(w, err)
				return
			}
		}
		data, err := fn(&req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, data)
	})
}

func pathID(sc *svc.ServiceContext, w http.ResponseWriter, r *http.Request) (int64, bool) {
	var req struct {
		ID string `path:"id"`
	}
	if err := httpx.ParsePath(r, &req); err != nil {
		common.Err(w, common.ErrParam)
		return 0, false
	}
	id, err := parseID64(req.ID)
	if err != nil {
		common.Err(w, err)
		return 0, false
	}
	return id, true
}

func BargainLaunches(sc *svc.ServiceContext) http.HandlerFunc {
	return memberH[struct{}](sc, false, func(uid int64, _ *struct{}) (any, error) { return marketing.BargainLaunches(sc, uid) })
}

func BargainLaunch(sc *svc.ServiceContext) http.HandlerFunc {
	return memberH[types.BargainLaunchReq](sc, true, func(uid int64, req *types.BargainLaunchReq) (any, error) {
		id, err := parseID64(req.ActivityID)
		if err != nil {
			return nil, err
		}
		return marketing.BargainLaunch(sc, uid, id)
	})
}

func BargainHelp(sc *svc.ServiceContext) http.HandlerFunc {
	return memberH[types.BargainHelpReq](sc, true, func(uid int64, req *types.BargainHelpReq) (any, error) {
		id, err := parseID64(req.LaunchID)
		if err != nil {
			return nil, err
		}
		return marketing.BargainHelp(sc, uid, id)
	})
}

func AdminBargainUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminH[types.BargainActivityUpsert](sc, true, func(req *types.BargainActivityUpsert) (any, error) { return marketing.AdminBargainUpsert(sc, req) })
}

func AdminBargainList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminH[struct{}](sc, false, func(_ *struct{}) (any, error) { return marketing.AdminBargainList(sc) })
}

func Auctions(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := marketing.Auctions(sc, optionalMemberID(sc, r))
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	}
}

func AuctionDepositPay(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `path:"id"`
		}
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		id, err := parseID64(req.ID)
		if err != nil {
		}
		resp, err := marketing.AuctionDepositPay(sc, memberID(r), id)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func AuctionBid(sc *svc.ServiceContext) http.HandlerFunc {
	return memberH[types.AuctionBidReq](sc, true, func(uid int64, req *types.AuctionBidReq) (any, error) {
		id, err := parseID64(req.ID)
		if err != nil {
			return nil, err
		}
		d, _ := decimal.NewFromString(req.Price)
		return nil, marketing.AuctionBid(sc, uid, id, d)
	})
}

func AdminAuctionUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminH[types.AuctionUpsertReq](sc, true, func(req *types.AuctionUpsertReq) (any, error) { return marketing.AdminAuctionUpsert(sc, req) })
}

func AdminAuctionList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminH[struct{}](sc, false, func(_ *struct{}) (any, error) { return marketing.AdminAuctionList(sc) })
}

func Tasks(sc *svc.ServiceContext) http.HandlerFunc {
	return memberH[struct{}](sc, false, func(uid int64, _ *struct{}) (any, error) { return growth.Tasks(sc, uid) })
}

func TaskClaim(sc *svc.ServiceContext) http.HandlerFunc {
	return memberH[types.TaskClaimReq](sc, true, func(uid int64, req *types.TaskClaimReq) (any, error) {
		return nil, growth.ClaimTask(sc, uid, req.Key)
	})
}

func VipBuy(sc *svc.ServiceContext) http.HandlerFunc {
	return memberH[struct{}](sc, false, func(uid int64, _ *struct{}) (any, error) { return auth.VipBuy(sc, uid) })
}

func BookingCreate(sc *svc.ServiceContext) http.HandlerFunc {
	return memberH[types.BookingReq](sc, true, func(uid int64, req *types.BookingReq) (any, error) { return marketing.CreateBooking(sc, uid, req) })
}

func MyBookings(sc *svc.ServiceContext) http.HandlerFunc {
	return memberH[struct{}](sc, false, func(uid int64, _ *struct{}) (any, error) { return marketing.MyBookings(sc, uid) })
}

func BookingCancel(sc *svc.ServiceContext) http.HandlerFunc {
	return memberAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			No string `path:"no"`
		}
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		if err := marketing.CancelBooking(sc, memberID(r), req.No); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func AdminBookings(sc *svc.ServiceContext) http.HandlerFunc {
	return adminH[types.PageReq](sc, false, func(req *types.PageReq) (any, error) { return marketing.AdminBookings(sc, req) })
}

func AdminBookingVerify(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			BookingNo string `json:"bookingNo"`
			Code      string `json:"code"`
		}
		if err := common.ParseBody(r, &body); err != nil {
			common.Err(w, err)
			return
		}
		if err := marketing.VerifyBooking(sc, body.BookingNo, body.Code); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

func InvoiceApply(sc *svc.ServiceContext) http.HandlerFunc {
	return memberH[types.InvoiceApplyReq](sc, true, func(uid int64, req *types.InvoiceApplyReq) (any, error) {
		return nil, trade.InvoiceApply(sc, uid, req)
	})
}

func MyInvoices(sc *svc.ServiceContext) http.HandlerFunc {
	return memberH[struct{}](sc, false, func(uid int64, _ *struct{}) (any, error) { return trade.MyInvoices(sc, uid) })
}

func AdminInvoices(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Status int `form:"status,optional"`
		}
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		list, err := trade.AdminInvoices(sc, req.Status)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	})
}

func AdminInvoiceIssue(sc *svc.ServiceContext) http.HandlerFunc {
	return adminH[types.InvoiceIssueReq](sc, true, func(req *types.InvoiceIssueReq) (any, error) {
		id, err := parseID64(req.ID)
		if err != nil {
			return nil, err
		}
		return nil, trade.AdminInvoiceIssue(sc, id, req.Link)
	})
}

func MemberSegments(sc *svc.ServiceContext) http.HandlerFunc {
	return adminH[struct{}](sc, false, func(_ *struct{}) (any, error) { return manage.MemberSegments(sc) })
}

func SegmentMembers(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Segment string `form:"segment"`
		}
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		list, err := manage.SegmentMembers(sc, req.Segment)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	})
}

func SegmentIssue(sc *svc.ServiceContext) http.HandlerFunc {
	return adminH[types.SegmentIssueReq](sc, true, func(req *types.SegmentIssueReq) (any, error) {
		tpl, err := parseID64(req.CouponTemplateID)
		if err != nil {
			return nil, err
		}
		granted, err := manage.SegmentIssue(sc, req.Segment, tpl)
		return map[string]any{"granted": granted}, err
	})
}

func PriceSuggest(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `path:"id"`
		}
		if err := httpx.ParsePath(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		id, err := parseID64(req.ID)
		if err != nil {
		}
		resp, err := manage.PriceSuggest(sc, id)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}
