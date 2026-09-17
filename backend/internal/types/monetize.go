package types

type InsuranceProductView struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Company string `json:"company"`
	CoverDesc string `json:"coverDesc"`
	Price string `json:"price"`
}

type InsuranceApplyReq struct {
	ProductID string `json:"productId"`
	Contact string `json:"contact"`
	Phone string `json:"phone"`
}

type StudServiceView struct {
	ID string `json:"id"`
	BreedName string `json:"breedName"`
	PetName string `json:"petName"`
	HealthCerts string `json:"healthCerts"`
	Price string `json:"price"`
	Description string `json:"description"`
}

type StudUpsertReq struct {
	ID string `json:"id,optional"`
	BreedName string `json:"breedName"`
	PetName string `json:"petName"`
	HealthCerts string `json:"healthCerts,optional"`
	Price string `json:"price"`
	Description string `json:"description,optional"`
}

type SegmentListResp struct {
	List []SegmentRow `json:"list"`
}

type HomeConfigResp struct {
	Config string `json:"config"`
}

type HomeConfigSaveReq struct {
	Config string `json:"config"`
}

type DistributorApplyResp struct {
	Status string `json:"status"`
	CommissionRate string `json:"commissionRate"`
}

type DistributorInfo struct {
	Balance string `json:"balance"`
	TotalCommission string `json:"totalCommission"`
	CommissionRate string `json:"commissionRate"`
	Status int `json:"status"`
}

type WithdrawalReq struct {
	Amount string `json:"amount"`
}

type CommissionView struct {
	OrderNo string `json:"orderNo"`
	Amount string `json:"amount"`
	Status int `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type ScheduledOffSaleReq struct {
	ProductID string `json:"productId"`
	OffSaleAt string `json:"offSaleAt,optional"` // RFC3339, 空=取消
}

type ProductDetailPatches struct {
	CertType string `json:"certType,optional"`
	CertNo string `json:"certNo,optional"`
	ChipNo string `json:"chipNo,optional"`
	PreSale int `json:"preSale,optional"`
	PreSalePrice string `json:"preSalePrice,optional"`
	PreSaleETA string `json:"preSaleEta,optional"`
	ScheduledOffSaleAt string `json:"scheduledOffSaleAt,optional"`
}
