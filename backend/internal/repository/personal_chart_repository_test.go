package repository_test

import (
	"context"
	"errors"
	"testing"
	"yuanju/internal/repository"
	"yuanju/pkg/database"
)

func TestPersonalChartOwnershipDeleteAndMerge(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()
	restore := swapGlobalDB(t, db)
	defer restore()
	if _, err := database.Migrate(database.ModeApply); err != nil {
		t.Fatal(err)
	}

	var miniID, webID, otherID, miniChart, webChart, otherChart string
	for _, item := range []struct {
		email, source string
		dest          *string
	}{
		{"", "miniapp_wechat", &miniID}, {"web@example.com", "web", &webID}, {"other@example.com", "web", &otherID},
	} {
		if err := db.QueryRow(`INSERT INTO users (email, password_hash, source) VALUES (NULLIF($1,''), CASE WHEN $1='' THEN NULL ELSE 'hash' END, $2) RETURNING id`, item.email, item.source).Scan(item.dest); err != nil {
			t.Fatal(err)
		}
	}
	for _, item := range []struct {
		owner string
		dest  *string
	}{{miniID, &miniChart}, {webID, &webChart}, {otherID, &otherChart}} {
		if err := db.QueryRow(`INSERT INTO bazi_charts (user_id,birth_year,birth_month,birth_day,birth_hour,gender,chart_hash)
            VALUES ($1,1990,1,1,1,'male',gen_random_uuid()::text) RETURNING id`, item.owner).Scan(item.dest); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`UPDATE bazi_charts SET year_gan='甲',year_zhi='子',month_gan='乙',month_zhi='丑',
		day_gan='庚',day_zhi='午',hour_gan='丁',hour_zhi='卯',yongshen='金',jishen='火'`); err != nil {
		t.Fatal(err)
	}
	if err := repository.SetPersonalChart(miniID, otherChart); !errors.Is(err, repository.ErrPersonalChartNotOwned) {
		t.Fatalf("cross-user binding: %v", err)
	}
	if err := repository.SetPersonalChart(miniID, miniChart); err != nil {
		t.Fatal(err)
	}
	bound, err := repository.GetPersonalChart(miniID)
	if err != nil || bound == nil || bound.ID != miniChart {
		t.Fatalf("initial binding: %+v %v", bound, err)
	}
	if err := db.QueryRow(`INSERT INTO wechat_user_identities (app_id,open_id,user_id) VALUES ('app','open',$1) RETURNING user_id`, miniID).Scan(new(string)); err != nil {
		t.Fatal(err)
	}
	if err := repository.MergeMiniappIntoWebUser(context.Background(), miniID, webID, "app"); err != nil {
		t.Fatal(err)
	}
	bound, err = repository.GetPersonalChart(webID)
	if err != nil || bound == nil || bound.ID != miniChart {
		t.Fatalf("merged binding: %+v %v", bound, err)
	}
	var secondMiniID, secondChart string
	if err := db.QueryRow(`INSERT INTO users (source) VALUES ('miniapp_wechat') RETURNING id`).Scan(&secondMiniID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`INSERT INTO bazi_charts (user_id,birth_year,birth_month,birth_day,birth_hour,gender,chart_hash)
		VALUES ($1,1991,1,1,1,'female',gen_random_uuid()::text) RETURNING id`, secondMiniID).Scan(&secondChart); err != nil {
		t.Fatal(err)
	}
	if err := repository.SetPersonalChart(secondMiniID, secondChart); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO wechat_user_identities (app_id,open_id,user_id) VALUES ('app','open2',$1)`, secondMiniID); err != nil {
		t.Fatal(err)
	}
	// The real merge forbids linking a second WeChat identity to the same target.
	// Verify conflict preference with an independent web account instead.
	var secondWebID, secondWebChart string
	if err := db.QueryRow(`INSERT INTO users (email,password_hash,source) VALUES ('web2@example.com','hash','web') RETURNING id`).Scan(&secondWebID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`INSERT INTO bazi_charts (user_id,birth_year,birth_month,birth_day,birth_hour,gender,chart_hash)
		VALUES ($1,1992,1,1,1,'male',gen_random_uuid()::text) RETURNING id`, secondWebID).Scan(&secondWebChart); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE bazi_charts SET year_gan='甲',year_zhi='子',month_gan='乙',month_zhi='丑',
		day_gan='庚',day_zhi='午',hour_gan='丁',hour_zhi='卯',yongshen='金',jishen='火'
		WHERE id IN ($1,$2)`, secondChart, secondWebChart); err != nil {
		t.Fatal(err)
	}
	if err := repository.SetPersonalChart(secondWebID, secondWebChart); err != nil {
		t.Fatal(err)
	}
	if err := repository.MergeMiniappIntoWebUser(context.Background(), secondMiniID, secondWebID, "app"); err != nil {
		t.Fatal(err)
	}
	bound, err = repository.GetPersonalChart(secondWebID)
	if err != nil || bound == nil || bound.ID != secondWebChart {
		t.Fatalf("target binding must win: %+v %v", bound, err)
	}
	if err := repository.SetPersonalChart(webID, webChart); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.DeleteChart(webChart); err != nil {
		t.Fatal(err)
	}
	bound, err = repository.GetPersonalChart(webID)
	if err != nil || bound != nil {
		t.Fatalf("delete must clear binding: %+v %v", bound, err)
	}
	if err := repository.SetPersonalChart(webID, miniChart); err != nil {
		t.Fatal(err)
	}
	if err := repository.ClearPersonalChart(webID); err != nil {
		t.Fatal(err)
	}
	bound, err = repository.GetPersonalChart(webID)
	if err != nil || bound != nil {
		t.Fatalf("unbind: %+v %v", bound, err)
	}
}
