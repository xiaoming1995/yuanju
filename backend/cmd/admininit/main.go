// admininit creates the first administrator from a trusted local terminal only.
// Run from backend/: ADMIN_INIT_EMAIL=... ADMIN_INIT_NAME=... go run ./cmd/admininit
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"syscall"
	"yuanju/configs"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	email := strings.TrimSpace(os.Getenv("ADMIN_INIT_EMAIL"))
	name := strings.TrimSpace(os.Getenv("ADMIN_INIT_NAME"))
	if email == "" || name == "" {
		fail("请设置 ADMIN_INIT_EMAIL 和 ADMIN_INIT_NAME")
	}
	fmt.Fprint(os.Stderr, "请输入初始管理员密码：")
	line, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		fail("读取密码失败")
	}
	password := strings.TrimSpace(string(line))
	if len(password) < 12 {
		fail("管理员密码至少需要 12 位")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fail("密码加密失败")
	}
	configs.Load()
	db, err := sql.Open("postgres", configs.AppConfig.DatabaseURL)
	if err != nil {
		fail("数据库连接失败")
	}
	defer db.Close()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		fail("数据库事务失败")
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`LOCK TABLE admins IN EXCLUSIVE MODE`); err != nil {
		fail("锁定管理员表失败")
	}
	var count int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM admins`).Scan(&count); err != nil || count != 0 {
		fail("已有管理员或无法检查管理员表，初始化已中止")
	}
	if _, err := tx.Exec(`INSERT INTO admins (email, password_hash, name) VALUES ($1, $2, $3)`, email, string(hash), name); err != nil {
		fail("创建管理员失败")
	}
	if err := tx.Commit(); err != nil {
		fail("提交管理员失败")
	}
	fmt.Println("首个管理员已创建")
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
