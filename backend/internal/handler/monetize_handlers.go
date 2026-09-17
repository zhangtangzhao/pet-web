package handler

import (
	"net/http"
	"time"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/manage"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

func memberM[T any](sc *svc.ServiceContext, needBody bool, fn func(int64, *T) (any, error)) http.HandlerFunc {
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

func adminM[T any](sc *svc.ServiceContext, needBody bool, fn func(*T) (any, error)) http.HandlerFunc {
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

func InsuranceList(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := manage.InsuranceList(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	}
}

func InsuranceApply(sc *svc.ServiceContext) http.HandlerFunc {
	return memberM[types.InsuranceApplyReq](sc, true, func(uid int64, req *types.InsuranceApplyReq) (any, error) {
		return nil, manage.InsuranceApplyCreate(sc, uid, req)
	})
}

func AdminInsuranceUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminM[types.InsuranceProductView](sc, true, func(req *types.InsuranceProductView) (any, error) {
		return nil, manage.AdminInsuranceUpsert(sc, req)
	})
}

func AdminInsuranceApplyList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminM[struct{}](sc, false, func(_ *struct{}) (any, error) { return manage.AdminInsuranceApplyList(sc) })
}

func StudList(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := manage.StudList(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, list)
	}
}

func AdminStudUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminM[types.StudUpsertReq](sc, true, func(req *types.StudUpsertReq) (any, error) {
		return nil, manage.AdminStudUpsert(sc, req)
	})
}

func HomeConfigGet(sc *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := manage.HomeConfigGet(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, map[string]string{"config": cfg})
	}
}

func AdminHomeConfigGet(sc *svc.ServiceContext) http.HandlerFunc {
	return adminM[struct{}](sc, false, func(_ *struct{}) (any, error) { return manage.HomeConfigGet(sc) })
}

func AdminHomeConfigSet(sc *svc.ServiceContext) http.HandlerFunc {
	return adminM[types.HomeConfigSaveReq](sc, true, func(req *types.HomeConfigSaveReq) (any, error) {
		return nil, manage.HomeConfigSet(sc, req.Config)
	})
}

func DistributorApply(sc *svc.ServiceContext) http.HandlerFunc {
	return memberM[struct{}](sc, false, func(uid int64, _ *struct{}) (any, error) {
		return nil, manage.DistributorApply(sc, uid)
	})
}

func DistributorInfo(sc *svc.ServiceContext) http.HandlerFunc {
	return memberM[struct{}](sc, false, func(uid int64, _ *struct{}) (any, error) { return manage.DistributorInfo(sc, uid) })
}

func DistributorWithdrawal(sc *svc.ServiceContext) http.HandlerFunc {
	return memberM[types.WithdrawalReq](sc, true, func(uid int64, req *types.WithdrawalReq) (any, error) {
		return nil, manage.DistributorWithdrawal(sc, uid, req.Amount)
	})
}

func ScheduledOffSale(sc *svc.ServiceContext) http.HandlerFunc {
	return adminM[types.ScheduledOffSaleReq](sc, true, func(req *types.ScheduledOffSaleReq) (any, error) {
		id, err := parseID64(req.ProductID)
		if err != nil {
			return nil, err
		}
		var at *time.Time
		if req.OffSaleAt != "" {
			t, e := time.Parse(time.RFC3339, req.OffSaleAt)
			if e != nil {
				return nil, common.ErrParam
			}
			at = &t
		}
		return nil, manage.ScheduledOffSale(sc, id, at)
	})
}
