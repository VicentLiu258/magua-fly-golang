package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// ContractHandler 后台合同创建/更新/详情。
type ContractHandler struct {
	contractUC *usecase.ContractUseCase
	filesUC    *usecase.ContractFilesUseCase
}

// NewContractHandler 创建 ContractHandler。
func NewContractHandler(contractUC *usecase.ContractUseCase, filesUC *usecase.ContractFilesUseCase) *ContractHandler {
	return &ContractHandler{contractUC: contractUC, filesUC: filesUC}
}

func parseOptionalDateString(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", t)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func contractToJSON(c *domain.Contract) gin.H {
	out := gin.H{
		"id":                c.ID,
		"student_id":        c.StudentID,
		"contract_no":       c.ContractNo,
		"total_minutes":     c.TotalMinutes,
		"price_total":       c.PriceTotal,
		"currency":          c.Currency,
		"status":            c.Status,
		"payment_mode":      c.PaymentMode,
		"payment_schedule":  c.PaymentSchedule,
		"mongo_file_ids":    c.MongoFileIDs,
		"remark":            c.Remark,
		"created_at":        c.CreatedAt,
		"updated_at":        c.UpdatedAt,
	}
	if c.SignDate != nil {
		out["sign_date"] = c.SignDate.Format("2006-01-02")
	} else {
		out["sign_date"] = nil
	}
	if c.StartDate != nil {
		out["start_date"] = c.StartDate.Format("2006-01-02")
	} else {
		out["start_date"] = nil
	}
	if c.EndDate != nil {
		out["end_date"] = c.EndDate.Format("2006-01-02")
	} else {
		out["end_date"] = nil
	}
	return out
}

func writeContractErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": err.Error(), "data": nil})
	case errors.Is(err, usecase.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
	case errors.Is(err, usecase.ErrInvalidPaymentMode),
		errors.Is(err, usecase.ErrInvalidPaymentSchedule),
		errors.Is(err, usecase.ErrPaymentScheduleSumMismatch),
		errors.Is(err, usecase.ErrPriceTotalRequired):
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
	}
}

// AdminContractCreate POST /api/v1/admin/contracts/create
func (h *ContractHandler) AdminContractCreate(c *gin.Context) {
	var body struct {
		StudentID       string                       `json:"student_id" binding:"required"`
		ContractNo      string                       `json:"contract_no"`
		SignDate        *string                      `json:"sign_date"`
		StartDate       *string                      `json:"start_date"`
		EndDate         *string                      `json:"end_date"`
		TotalMinutes    *int                         `json:"total_minutes"`
		PriceTotal      *float64                     `json:"price_total"`
		Currency        string                       `json:"currency"`
		Status          string                       `json:"status"`
		Remark          string                       `json:"remark"`
		PaymentMode     string                       `json:"payment_mode"`
		PaymentSchedule []domain.PaymentInstallment  `json:"payment_schedule"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	signDate, err := parseOptionalDateString(body.SignDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid sign_date", "data": nil})
		return
	}
	startDate, err := parseOptionalDateString(body.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid start_date", "data": nil})
		return
	}
	endDate, err := parseOptionalDateString(body.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid end_date", "data": nil})
		return
	}
	out, err := h.contractUC.Create(c.Request.Context(), &usecase.ContractCreateInput{
		StudentID:       body.StudentID,
		ContractNo:      body.ContractNo,
		SignDate:        signDate,
		StartDate:       startDate,
		EndDate:         endDate,
		TotalMinutes:    body.TotalMinutes,
		PriceTotal:      body.PriceTotal,
		Currency:        body.Currency,
		Status:          body.Status,
		Remark:          body.Remark,
		PaymentMode:     body.PaymentMode,
		PaymentSchedule: body.PaymentSchedule,
	})
	if err != nil {
		writeContractErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": contractToJSON(out)})
}

// AdminContractUpdate POST /api/v1/admin/contracts/update
func (h *ContractHandler) AdminContractUpdate(c *gin.Context) {
	var body struct {
		ID              string                       `json:"id" binding:"required"`
		ContractNo      *string                      `json:"contract_no"`
		SignDate        *string                      `json:"sign_date"`
		StartDate       *string                      `json:"start_date"`
		EndDate         *string                      `json:"end_date"`
		TotalMinutes    *int                         `json:"total_minutes"`
		PriceTotal      *float64                     `json:"price_total"`
		Currency        *string                      `json:"currency"`
		Status          *string                      `json:"status"`
		Remark          *string                      `json:"remark"`
		PaymentMode     *string                      `json:"payment_mode"`
		PaymentSchedule *[]domain.PaymentInstallment `json:"payment_schedule"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	in := &usecase.ContractUpdateInput{
		ID:              body.ID,
		ContractNo:      body.ContractNo,
		TotalMinutes:    body.TotalMinutes,
		PriceTotal:      body.PriceTotal,
		Currency:        body.Currency,
		Status:          body.Status,
		Remark:          body.Remark,
		PaymentMode:     body.PaymentMode,
		PaymentSchedule: body.PaymentSchedule,
	}
	if body.SignDate != nil {
		t, err := parseOptionalDateString(body.SignDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid sign_date", "data": nil})
			return
		}
		in.SignDate = t
	}
	if body.StartDate != nil {
		t, err := parseOptionalDateString(body.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid start_date", "data": nil})
			return
		}
		in.StartDate = t
	}
	if body.EndDate != nil {
		t, err := parseOptionalDateString(body.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid end_date", "data": nil})
			return
		}
		in.EndDate = t
	}
	out, err := h.contractUC.Update(c.Request.Context(), in)
	if err != nil {
		writeContractErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": contractToJSON(out)})
}

// AdminContractDetail POST /api/v1/admin/contracts/detail
func (h *ContractHandler) AdminContractDetail(c *gin.Context) {
	var body struct {
		ID string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	ct, err := h.contractUC.GetByID(c.Request.Context(), body.ID)
	if err != nil {
		writeContractErr(c, err)
		return
	}
	var files []interface{}
	if h.filesUC != nil {
		list, err := h.filesUC.ListContractFiles(c.Request.Context(), body.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
			return
		}
		for _, f := range list {
			files = append(files, f)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{
			"contract": contractToJSON(ct),
			"files":    files,
		},
	})
}
