package prompt

// Drift 状态：维护版（DB 行）相对当前出厂版（代码 canonical）的关系。
const (
	DriftAligned      = "aligned"      // 维护版内容 == 当前出厂版
	DriftCustomized   = "customized"   // 基于当前出厂版，但内容被管理员改过
	DriftOutdated     = "outdated"     // 分支点落后于当前出厂版（出厂已更新）
	DriftUnregistered = "unregistered" // 代码侧无此 canonical（历史遗留）
)

// DriftStatus 比对 DB 行（dbContent + dbCanonicalHash 分支点）与当前出厂版，
// 得出漂移状态。纯函数，不触 DB。
func DriftStatus(module, dbContent, dbCanonicalHash string) string {
	def, ok := Lookup(module)
	if !ok {
		return DriftUnregistered
	}
	if dbCanonicalHash != def.Hash {
		return DriftOutdated
	}
	if HashContent(dbContent) == def.Hash {
		return DriftAligned
	}
	return DriftCustomized
}
