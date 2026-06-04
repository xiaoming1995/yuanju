// 自助重置管理员密码：用项目同款 bcrypt(DefaultCost) 生成哈希，打印 UPDATE SQL。
// 不连数据库、不自动改库——你看过 SQL 后自己执行。用完即删。
//
// 用法：
//   cd backend
//   NEWPW='你的新密码' EMAIL='admin@yuanju.com' go run ./cmd/adminpw
package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	pw := os.Getenv("NEWPW")
	email := os.Getenv("EMAIL")
	if pw == "" || email == "" {
		fmt.Fprintln(os.Stderr, "用法: NEWPW='新密码' EMAIL='管理员邮箱' go run ./cmd/adminpw")
		os.Exit(1)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bcrypt 失败:", err)
		os.Exit(1)
	}
	// 单引号转义，防 SQL 注入
	safeEmail := strings.ReplaceAll(email, "'", "''")
	fmt.Printf("UPDATE admins SET password_hash='%s' WHERE email='%s';\n", string(hash), safeEmail)
}
