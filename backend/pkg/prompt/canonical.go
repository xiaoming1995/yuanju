// Package prompt centralizes AI prompt template canonical definitions.
//
// Canonical 注册表是代码侧权威 prompt 源；启动期 SyncCanonical 把它对齐到
// DB 的 ai_prompts 表（详见 sync.go）。新增模块需在 canonical_xxx.go 文件的
// init() 中调用 Register(module, Definition{...})。
package prompt

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

// Definition 描述一个 AI prompt 模板及其代码侧权威版本。
type Definition struct {
	// Version 是描述性版本标识符（free-form string），比较时仅做相等判定。
	// 改 prompt 时应同步升级此字段：它会作为「出厂版本号」展示在后台漂移徽标
	// （outdated 时显示「出厂已更新到 vX」）。SyncCanonical 不再据此覆盖任何已存在行。
	Version string

	// Description 简述本 prompt 用途，用于 ai_prompts.description 列。
	Description string

	// Content 是完整的 prompt 文本（Go template 语法，渲染时配合 model.XxxPromptData）。
	Content string

	// Hash 是 sha256(Content) 的 hex 字符串，Register 时自动计算。
	Hash string
}

// canonical 是模块名 → Definition 的全局注册表。
// 由各 canonical_xxx.go 的 init() 写入；运行时只读。
// 外部调用者请使用 Lookup / Has / MustGet。
var canonical = map[string]Definition{}

var registerMu sync.Mutex

// Register 把一个模块的 canonical Definition 写入注册表。
// 自动对 def.Content 计算 sha256 填充 def.Hash。
//
// 同一 module 重复 Register 会 panic —— 这是代码 bug 而非运行时错误，
// 因为各模块的 canonical_xxx.go init() 只应执行一次。
func Register(module string, def Definition) {
	registerMu.Lock()
	defer registerMu.Unlock()
	if _, exists := canonical[module]; exists {
		panic(fmt.Sprintf("prompt: module %q already registered", module))
	}
	def.Hash = HashContent(def.Content)
	canonical[module] = def
}

// HashContent 返回 content 的 sha256 hex 字符串，与 canonical Hash 同一算法。
func HashContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// MustGet 返回指定 module 的 canonical Definition；
// 未注册的 module 直接 panic（不静默回退空字符串）。
func MustGet(module string) Definition {
	def, ok := canonical[module]
	if !ok {
		panic(fmt.Sprintf("prompt: module %q not registered in canonical map", module))
	}
	return def
}

// Lookup returns the Definition for module if registered, plus a boolean ok.
// Use this when the caller can handle the unregistered case (e.g., admin reset endpoint).
func Lookup(module string) (Definition, bool) {
	def, ok := canonical[module]
	return def, ok
}

// Has reports whether module is registered. Equivalent to `_, ok := Lookup(module)`.
func Has(module string) bool {
	_, ok := canonical[module]
	return ok
}
