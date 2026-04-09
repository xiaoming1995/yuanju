package repository

import (
	"database/sql"
	"yuanju/pkg/database"
)

// AlgoTiaohouRow 调候规则表记录
type AlgoTiaohouRow struct {
	DayGan     string
	MonthZhi   string
	XiElements string // 逗号分隔的天干，如 "丙,癸"
	Text       string
}

// CountAlgoTiaohou 返回表中记录数（用于判断是否需要 seed）
func CountAlgoTiaohou() (int, error) {
	var count int
	err := database.DB.QueryRow(`SELECT COUNT(*) FROM algo_tiaohou`).Scan(&count)
	return count, err
}

// GetAllAlgoTiaohou 读取全部调候规则，可按 day_gan 过滤（传空字符串则不过滤）
func GetAllAlgoTiaohou(dayGan string) ([]AlgoTiaohouRow, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if dayGan != "" {
		rows, err = database.DB.Query(
			`SELECT day_gan, month_zhi, xi_elements, text FROM algo_tiaohou
			 WHERE day_gan = $1 ORDER BY day_gan, month_zhi`,
			dayGan,
		)
	} else {
		rows, err = database.DB.Query(
			`SELECT day_gan, month_zhi, xi_elements, text FROM algo_tiaohou
			 ORDER BY day_gan, month_zhi`,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AlgoTiaohouRow
	for rows.Next() {
		var r AlgoTiaohouRow
		if err := rows.Scan(&r.DayGan, &r.MonthZhi, &r.XiElements, &r.Text); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

// UpsertAlgoTiaohou 写入或更新单条调候规则
func UpsertAlgoTiaohou(dayGan, monthZhi, xiElements, text string) error {
	_, err := database.DB.Exec(
		`INSERT INTO algo_tiaohou (day_gan, month_zhi, xi_elements, text)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (day_gan, month_zhi) DO UPDATE
		 SET xi_elements = EXCLUDED.xi_elements,
		     text = EXCLUDED.text,
		     updated_at = NOW()`,
		dayGan, monthZhi, xiElements, text,
	)
	return err
}

// DeleteAlgoTiaohou 删除单条自定义调候规则（恢复默认值）
func DeleteAlgoTiaohou(dayGan, monthZhi string) error {
	_, err := database.DB.Exec(
		`DELETE FROM algo_tiaohou WHERE day_gan = $1 AND month_zhi = $2`,
		dayGan, monthZhi,
	)
	return err
}
