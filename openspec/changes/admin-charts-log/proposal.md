## Why

为管理员提供洞察平台生命力的最核心报表。管理员需要实时查看平台上发生的起盘历史（用户归属、时间、八字格局、推算结果等），以此来感知平台日活、了解用户偏好并验证算法输出。

## What Changes

- 后端添加新接口 `GET /api/admin/charts` 以分页拉取关联了用户信息的八字排盘数据 (`bazi_charts` LEFT JOIN `users`)。
- 前端添加【起盘明细】(`AdminChartsPage.tsx`) 后台页面，使用 `lucide-react` 图标统一风格。
- 更新管理后台侧边栏，增加对应路由入口。

## Capabilities

### New Capabilities
- `admin-charts-log`: 平台所有起盘历史记录（包括普通用户和游客）的管理功能。

### Modified Capabilities

## Impact

将波及 Admin 后台模块的 API (`admin_handler`, `admin_repository`)，以及 Admin 前端页面路由体系。对普通 C 端用户无任何影响。
