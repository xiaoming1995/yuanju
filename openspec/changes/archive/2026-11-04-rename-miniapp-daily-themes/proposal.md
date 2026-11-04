## Why

现有四个每日主题都以“运”命名，容易让人把基于干支规则的日常参考理解为结果预测。用户已选定更贴近日常阅读的四个名称，需要让首页、详情和接口保持一致。

## What Changes

- 按固定顺序将四个展示名称改为“人际往来、财星参考、桃花缘分、作息节奏”。
- 保留主题 ID、计算规则、等级、依据和建议；首页与每日详情使用相同名称。
- 更新对应规格、说明和展示测试。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `miniapp-daily-four-theme-guidance`: 调整四主题的用户可见名称。
- `miniapp-home-daily-guidance`: 首页与详情共用新的四主题名称。

## Impact

后端个人每日主题接口的 `title` 字段、小程序相关测试和说明文档；接口 ID 与数据结构不变。
