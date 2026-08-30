package handler

import (
	"database/sql"
	"net/http"
	"strings"
	"yuanju/internal/model"
	"yuanju/internal/repository"

	"github.com/gin-gonic/gin"
)

var listPublishedMingGeHistoricalFigures = repository.ListPublishedMingGeHistoricalFigures

type mingGeHistoricalFigureRequest struct {
	MingGe               string `json:"ming_ge" binding:"required"`
	FigureName           string `json:"figure_name" binding:"required"`
	Era                  string `json:"era" binding:"required"`
	Identity             string `json:"identity" binding:"required"`
	HistoricalMemory     string `json:"historical_memory" binding:"required"`
	TurningPoint         string `json:"turning_point"`
	TurningPointYear     string `json:"turning_point_year"`
	SourceTitle          string `json:"source_title" binding:"required"`
	SourceURL            string `json:"source_url" binding:"required"`
	BirthDataPrecision   string `json:"birth_data_precision"`
	BaziVerificationNote string `json:"bazi_verification_note"`
	DayunPeriod          string `json:"dayun_period"`
	DayunExplanation     string `json:"dayun_explanation"`
	ShowDayun            bool   `json:"show_dayun"`
	DisplayOrder         int    `json:"display_order"`
	ReviewStatus         string `json:"review_status"`
}

func (request mingGeHistoricalFigureRequest) toModel() model.MingGeHistoricalFigure {
	return model.MingGeHistoricalFigure{
		MingGe: request.MingGe, FigureName: request.FigureName, Era: request.Era,
		Identity: request.Identity, HistoricalMemory: request.HistoricalMemory,
		TurningPoint: request.TurningPoint, TurningPointYear: request.TurningPointYear,
		SourceTitle: request.SourceTitle, SourceURL: request.SourceURL,
		BirthDataPrecision: request.BirthDataPrecision, BaziVerificationNote: request.BaziVerificationNote,
		DayunPeriod: request.DayunPeriod, DayunExplanation: request.DayunExplanation,
		ShowDayun: request.ShowDayun, DisplayOrder: request.DisplayOrder, ReviewStatus: request.ReviewStatus,
	}
}

// GetMingGeHistoricalFigures exposes only published, editorially approved references.
// GET /api/mingge-historical-figures/:ming_ge
func GetMingGeHistoricalFigures(c *gin.Context) {
	mingGe := strings.TrimSpace(c.Param("ming_ge"))
	if mingGe == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "命格不能为空"})
		return
	}
	figures, err := listPublishedMingGeHistoricalFigures(mingGe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取古人映照失败"})
		return
	}
	if figures == nil {
		figures = make([]model.MingGeHistoricalFigure, 0)
	}
	if len(figures) > 2 {
		figures = figures[:2]
	}
	c.JSON(http.StatusOK, gin.H{"data": figures})
}

func AdminListMingGeHistoricalFigures(c *gin.Context) {
	figures, err := repository.ListMingGeHistoricalFigures()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取古人映照列表失败"})
		return
	}
	if figures == nil {
		figures = make([]model.MingGeHistoricalFigure, 0)
	}
	c.JSON(http.StatusOK, gin.H{"data": figures})
}

func AdminCreateMingGeHistoricalFigure(c *gin.Context) {
	var request mingGeHistoricalFigureRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请补齐古人映照的必填内容"})
		return
	}
	figure, err := repository.CreateMingGeHistoricalFigure(request.toModel())
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": figure})
}

func AdminUpdateMingGeHistoricalFigure(c *gin.Context) {
	var request mingGeHistoricalFigureRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请补齐古人映照的必填内容"})
		return
	}
	figure, err := repository.UpdateMingGeHistoricalFigure(c.Param("id"), request.toModel())
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "古人映照不存在"})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": figure})
}

func AdminArchiveMingGeHistoricalFigure(c *gin.Context) {
	if err := repository.ArchiveMingGeHistoricalFigure(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "归档古人映照失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "已归档"})
}
