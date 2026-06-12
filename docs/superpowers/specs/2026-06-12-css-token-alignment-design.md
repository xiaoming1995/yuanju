# CSS Token 对齐（严格不变样）设计

> Status: Approved for spec review
> Date: 2026-06-12
> Scope: frontend/src 全部组件/页面 CSS 的写死值 → 设计变量替换 + 样式不一致建议清单
> Related:
> - `docs/superpowers/specs/2026-06-03-visual-system-unification-design.md`（token 体系来源，本设计是其"清理与验收"批次的延续）
> - `docs/superpowers/specs/2026-06-01-ux-style-layout-audit-design.md`

## 1. 背景

`frontend/src/index.css` 已有完整设计变量体系（五行色、语义色、字号档、间距档、圆角档），但存量组件 CSS 中仍有大量写死值绕过了这套变量：

- 16 个 CSS 文件含写死十六进制颜色，其中存在与标准金 `#c9a84c` 近似但不同的杂色（`#c9a227`、`#c9a96e`）。
- `font-size` 写死 px 共 370+ 处，分布在 9px–32px 共 15 档，而 type scale 只定义 5 档。

本设计处理其中**零风险**的部分：值与变量定义完全相同的写死值替换为 `var()` 引用。近似值合并、档位收敛属于会改变视觉的归一化，本次不做，只产出建议清单供后续逐项决策。

## 2. 目标

1. 像素级不改变任何页面视觉表现。
2. 组件 CSS 中与设计变量值完全相同的写死值全部替换为 `var()` 引用。
3. 产出样式不一致建议清单（带 `文件:行号`），覆盖：近似杂色合并候选、字号档位收敛建议、代码可见的可疑对齐点、inline style 治理方向。

## 3. 非目标

- 不做近似值合并 / 档位收敛（任何会改变渲染结果的修改）。
- 不动 TSX 中的 inline style（写进建议清单）。
- 不新增 / 修改 / 删除任何设计变量定义。
- 不重构 CSS 结构、不改选择器、不调整代码组织。

## 4. 替换规则

1. **精确匹配才替换**：字面值与 `index.css` 某变量定义值完全一致（忽略空格、大小写、`#FFF`/`#ffffff` 等等价缩写）才替换。不做"视觉上差不多"的推断。
2. **一值多变量按语义选**：如 `#4caf7d` 同时是 `--wu-mu` 和 `--status-success`——表示五行属性的场景用 `--wu-mu`，表示成功状态的场景用 `--status-success`。每处选择记入建议清单附录。
3. **语义命名变量需上下文匹配**：值匹配只是必要条件。对语义命名的变量（`--font-size-body`、`--radius-panel` 等），还要求使用场景与变量语义一致才替换（正文文字才换 `--font-size-body`）；场景不符或拿不准的保持原样，记入建议清单。纯取值类变量（颜色、`--radius-sm/md/lg`、阴影、过渡）精确匹配即替换。
4. **跳过导出类文件**：`CompatibilityShareCard.css`、`CompatibilityPrintLayout.css`（及对应 TSX 内联样式）。这些文件特意使用浅色主题，写死颜色是有意脱离深色主题变量；且 html2canvas 等导出库对 CSS 变量解析存在兼容性风险，替换无收益。
5. **跳过 `index.css` 本身**：变量定义文件不在替换范围。

## 5. 执行步骤

每轮独立 commit，可单独回滚：

1. **颜色轮**：hex / rgba 写死值 → 颜色变量。
2. **字号轮**：`font-size` 写死 px → type scale 变量（仅精确匹配档位，如 `15px` → `var(--font-size-body)`）。
3. **形状轮**：圆角、阴影、过渡时间 → 对应变量。
4. **建议清单**：输出 `docs/superpowers/specs/2026-06-12-css-style-suggestions.md`。

## 6. 验证方式

每轮替换后全部满足：

- `cd frontend && npm run build` 通过（含 tsc 类型检查）。
- `cd frontend && npm run lint` 通过。
- 逐项核对：diff 中每处替换，`var()` 指向变量的定义值与被替换字面值一致。
- grep 计数：替换前该类写死值出现次数 − 替换后出现次数 = 替换处数。

## 7. 风险与控制

| 风险 | 控制 |
|---|---|
| 误替换语义不符的变量（值同义不同） | 规则 2/3 按语义选择并留记录；diff 逐项核对 |
| 导出图片/打印布局因变量解析问题变样 | 规则 3 整体跳过导出类文件 |
| 一轮改动过大难以审查 | 按类别分 3 轮独立提交 |
