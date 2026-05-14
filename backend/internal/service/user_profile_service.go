package service

import (
	"fmt"
	"yuanju/internal/model"
	"yuanju/internal/repository"
)

func GetUserProfileOverview(userID string) (*model.UserProfileOverview, error) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	chartCount, err := repository.CountChartsByUserID(userID)
	if err != nil {
		return nil, err
	}
	aiReportCount, err := repository.CountAIReportsByUserID(userID)
	if err != nil {
		return nil, err
	}
	compatibilityCount, err := repository.CountCompatibilityReadingsByUserID(userID)
	if err != nil {
		return nil, err
	}
	recentCharts, err := repository.ListRecentChartsForProfile(userID, 5)
	if err != nil {
		return nil, err
	}
	recentCompatibility, err := repository.ListRecentCompatibilityForProfile(userID, 3)
	if err != nil {
		return nil, err
	}

	return &model.UserProfileOverview{
		User: user,
		Stats: model.UserProfileStats{
			ChartCount:         chartCount,
			AIReportCount:      aiReportCount,
			CompatibilityCount: compatibilityCount,
		},
		RecentCharts:        recentCharts,
		RecentCompatibility: recentCompatibility,
		Features: []model.UserProfileFeatureEntry{
			{
				Key:         "wallet",
				Title:       "充值与点数",
				Description: "后续用于充值、点数余额和消费流水管理。",
				Status:      "coming_soon",
			},
			{
				Key:         "pdf_templates",
				Title:       "PDF 模板定制",
				Description: "后续用于设置命书封面、报告模板和导出偏好。",
				Status:      "coming_soon",
			},
		},
	}, nil
}
