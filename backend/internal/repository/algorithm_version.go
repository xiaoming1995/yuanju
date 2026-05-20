package repository

// CurrentAlgorithmVersion 当前生成 AI 报告 / 大运总结 所用的算法版本。
//
// 写入新行时使用此常量；老行 algorithm_version IS NULL 被视为 v1 baseline。
// 版本变化时（如 Phase 2 落地）需同步更新此常量并新增 migration。
//
// 版本历史：
//   v1            (NULL) — pre-yongshen-realignment baseline
//   v2-yongshen-shishen   — 喜忌十神 prompt 注入 + algorithm_version 列建立
//   v3-progressive-compressed — lazy-load dayun_indexes 过滤 + YearsData prompt 压缩
//   v3.1-narrative-guarded — AI 空 narrative / validator 清空走 template 兜底，
//                            prompt 追加弱信号年安全措辞指引
const CurrentAlgorithmVersion = "v3.1-narrative-guarded"
