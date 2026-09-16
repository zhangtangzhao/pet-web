// finance_handlers.go 平台端财务对账 / CSV 导出 / 供货商管理。
package handler

import (
	"encoding/csv"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"pet/backend/internal/common"
	"pet/backend/internal/logic/manage"
	"pet/backend/internal/svc"
	"pet/backend/internal/types"
)

// AdminFinancePayments 支付流水
func AdminFinancePayments(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.FinancePaymentListReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := manage.AdminFinancePayments(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// AdminFinanceDaily 按日收支汇总
func AdminFinanceDaily(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.FinanceDailyReq
		if err := httpx.ParseForm(r, &req); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		resp, err := manage.AdminFinanceDaily(sc, req.Days)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

func writeCSV(w http.ResponseWriter, filename string, rows [][]string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	_, _ = w.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM，Excel 直开不乱码
	cw := csv.NewWriter(w)
	_ = cw.WriteAll(rows)
	cw.Flush()
}

// AdminExportOrders 订单导出 CSV
func AdminExportOrders(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		rows, err := manage.ExportOrders(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		writeCSV(w, "orders.csv", rows)
	})
}

// AdminExportMembers 会员导出 CSV
func AdminExportMembers(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		rows, err := manage.ExportMembers(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		writeCSV(w, "members.csv", rows)
	})
}

// AdminExportPoints 积分流水导出 CSV
func AdminExportPoints(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		rows, err := manage.ExportPoints(sc)
		if err != nil {
			common.Err(w, err)
			return
		}
		writeCSV(w, "points.csv", rows)
	})
}

// AdminSupplierList 供货商列表（含停用）
func AdminSupplierList(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.SupplierListReq
		_ = httpx.ParseForm(r, &req)
		resp, err := manage.AdminSupplierList(sc, &req)
		if err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, resp)
	})
}

// AdminSupplierUpsert 新增 / 编辑供货商
func AdminSupplierUpsert(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var req types.SupplierUpsertReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.AdminSupplierUpsert(sc, &req); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// AdminSupplierStatus 启用 / 停用
func AdminSupplierStatus(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		var req types.BannerStatusReq
		if err := common.ParseBody(r, &req); err != nil {
			common.Err(w, err)
			return
		}
		id, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.AdminSupplierStatus(sc, id, req.Status); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}

// AdminSupplierDelete 删除（关联商品时拒绝）
func AdminSupplierDelete(sc *svc.ServiceContext) http.HandlerFunc {
	return adminAuth(sc, func(w http.ResponseWriter, r *http.Request) {
		var path types.IDPathReq
		if err := httpx.ParsePath(r, &path); err != nil {
			common.Err(w, common.ErrParam)
			return
		}
		id, err := parseID64(path.ID)
		if err != nil {
			common.Err(w, err)
			return
		}
		if err := manage.AdminSupplierDelete(sc, id); err != nil {
			common.Err(w, err)
			return
		}
		common.OK(w, nil)
	})
}
