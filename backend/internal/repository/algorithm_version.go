package repository

// CurrentAlgorithmVersion 当前生成 AI 报告 / 大运总结 所用的算法版本。
//
// 写入新行时使用此常量；老行 algorithm_version IS NULL 被视为 v1 baseline。
// 版本变化时（如 Phase 2 落地）需同步更新此常量并新增 migration。
const CurrentAlgorithmVersion = "v2-yongshen-shishen"
