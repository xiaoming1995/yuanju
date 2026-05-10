# Docker Compose 数据库连接覆盖事故记录

## 现象

线上通过 `docker compose up -d --build` 启动后，后端日志报错：

```text
未找到 .env 文件，使用环境变量
数据库 Ping 失败: pq: password authentication failed for user "yuanju" (28P01)
```

其中第一行在 Docker Compose 部署下不是根因。容器内通常没有 `.env` 文件，应用会正常转而读取容器环境变量。

## 根因

`docker-compose.yml` 中 `backend` 服务曾经硬编码了：

```yaml
environment:
  DATABASE_URL: postgres://yuanju:yuanju123@postgres:5432/yuanju?sslmode=disable
```

这会覆盖 `env_file: ./backend/.env` 中的 `DATABASE_URL`。

当时线上 `backend/.env` 与 `postgres` 容器实际使用的是另一套密码：

```env
DATABASE_URL=postgres://yuanju:c9e5fdec89fe2def0e7eabeb@postgres:5432/yuanju?sslmode=disable
POSTGRES_PASSWORD=c9e5fdec89fe2def0e7eabeb
```

结果变成：

- `yuanju_backend` 用 `yuanju123` 连库
- `yuanju_postgres` 使用 `c9e5fdec89fe2def0e7eabeb`

从而触发认证失败。

## 排查结论

通过 `docker compose config` 可以直接看出最终生效配置是否被覆盖：

```bash
docker compose config | sed -n '/backend:/,/frontend:/p'
```

如果输出中出现硬编码的：

```yaml
DATABASE_URL: postgres://yuanju:yuanju123@postgres:5432/yuanju?sslmode=disable
```

说明 `backend/.env` 没有按预期生效。

## 修复

删除 `docker-compose.yml` 中 `backend.environment.DATABASE_URL` 的硬编码，只保留：

```yaml
env_file:
  - ./backend/.env
environment:
  REDIS_URL: redis://redis:6379
```

然后重新部署：

```bash
docker compose down
docker compose up -d --build
```

## 防再犯规则

1. `backend` 服务不要在 `docker-compose.yml` 中硬编码 `DATABASE_URL`
2. `backend/.env` 中的 `DATABASE_URL` 密码必须与 `POSTGRES_PASSWORD` 一致
3. 任何改动数据库连接配置后，都要先运行：

```bash
docker compose config | sed -n '/backend:/,/frontend:/p'
```

确认最终渲染出的 `DATABASE_URL` 是预期值

4. 部署后必须检查：

```bash
docker logs yuanju_backend --tail 100
```

期望看到：

```text
✅ 数据库连接成功
✅ 数据库迁移完成
🚀 缘聚命理服务启动
```
