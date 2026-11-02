# 阿里云 ECS 上用 Docker Compose 部署 HTTPS

本文记录缘聚网页与微信小程序后端的部署方案。示例使用 Ubuntu ECS、Docker Compose 和 Caddy；`www.example.com`、`api.example.com` 都需要换成实际域名。

> **当前状态**：仓库已包含下文的公共／本地／生产 Compose 文件和 Caddyfile，但尚未配置真实域名、服务器、密钥或完成线上部署。

## 部署结构

```text
浏览器 https://www.example.com ─┐
                                ├─ Caddy :80/:443 ─ 前端 Nginx :80 ─ /api、/static ─ 后端 :9002
小程序 https://api.example.com ──┘             └─────────────────────────────── 后端 :9002
                                                                          ├─ PostgreSQL
                                                                          └─ Redis
```

前端 Nginx 已处理单页路由、`/api`、`/static` 和流式接口；网页端保持同域请求即可。小程序使用独立 API 域名，地址中不附加 `/api`。Caddy 是独立容器，在入口处处理 HTTPS，并自动申请、续期证书。

## 一、准备服务器与域名

1. 准备带公网 IP 的阿里云 ECS，安装 Docker Engine 和 Compose 插件。Ubuntu 可参考 [Docker 官方安装指南](https://docs.docker.com/engine/install/ubuntu/)。
2. 将网页和 API 域名的 A 记录解析到 ECS 公网 IP。
3. ECS 安全组允许公网访问 TCP 80、443；SSH 22 仅允许可信管理 IP。5432、6379、3000、9002 无需向公网开放。若服务器启用了系统防火墙，也要允许 80、443。
4. 若 ECS 位于中国内地，网站对外开放前先办理所需的 ICP 备案；服务器区域和备案状态应在购买与解析域名前确定。参见[阿里云备案说明](https://help.aliyun.com/zh/icp-filing/basic-icp-service/support/for-the-record-process-faq)。

## 二、选择本地或生产 Compose 配置

仓库已有以下三个文件：

| 文件 | 内容 |
| --- | --- |
| `docker-compose.yml` | PostgreSQL、Redis、后端、前端及持久化卷的公共配置；前后端**不写** `ports` |
| `docker-compose.local.yml` | 本地需要的 `9002:9002`、`3000:80` 端口映射 |
| `docker-compose.prod.yml` | Caddy 服务、80/443 公网端口、证书持久化卷 |

本地文件的端口配置：

```yaml
services:
  backend:
    ports:
      - "9002:9002"
  frontend:
    ports:
      - "3000:80"
```

生产文件的入口配置：

```yaml
services:
  caddy:
    image: caddy:2
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    environment:
      WEB_DOMAIN: ${WEB_DOMAIN:?Set WEB_DOMAIN in .env.production}
      API_DOMAIN: ${API_DOMAIN:?Set API_DOMAIN in .env.production}
    volumes:
      - ./caddy:/etc/caddy:ro
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - frontend
      - backend

volumes:
  caddy_data:
  caddy_config:
```

`9002:9002` 和 `3000:80` 已从公共文件移到本地文件。Compose 合并文件时会保留并合并端口列表，因此生产文件只与公共文件合并。数据库和 Redis 保持仅在 Docker 网络内可达。

明确指定配置文件，避免本地与生产混用：

```bash
# 本地
./scripts/docker-compose-up.sh

# 生产
docker compose --env-file .env.production -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

本地脚本已显式选择公共与本地文件。正式启动前分别用 `docker compose -f docker-compose.yml -f docker-compose.local.yml config --quiet` 和 `docker compose --env-file .env.production -f docker-compose.yml -f docker-compose.prod.yml config --quiet` 校验配置，并检查生产合并结果中没有对外发布 3000、9002、5432、6379。不要将含密钥的完整 `docker compose config` 输出粘贴到工单或聊天中。

## 三、配置 Caddy

仓库中的 `caddy/Caddyfile` 从容器环境变量读取域名：

```caddyfile
{$WEB_DOMAIN} {
    reverse_proxy frontend:80
}

{$API_DOMAIN} {
    reverse_proxy backend:9002
}
```

在容器内使用 Compose 服务名 `frontend`、`backend`，不能使用 `127.0.0.1` 指向别的容器。Caddy 会在域名解析正确、80/443 可从公网访问时自动申请并续期公开可信的证书。`caddy_data` 必须保留，不能当缓存清理；`/etc/caddy` 挂载整个目录，便于后续重新加载配置。参见 [Caddy 自动 HTTPS](https://caddyserver.com/docs/automatic-https)和[Caddy 官方镜像说明](https://hub.docker.com/_/caddy)。

## 四、准备环境变量和构建上下文

在**服务器本机**准备 `backend/.env.docker`。可先执行 `./scripts/docker-compose-up.sh --prepare-only` 生成随机初始密钥，再检查并填写实际配置；已有文件不会被脚本覆盖。`POSTGRES_PASSWORD` 必须与 `DATABASE_URL` 中的密码一致，数据库主机保持 `postgres`，Redis 主机保持 `redis`。网页使用默认的同域 `/api` 请求，不设置跨域的 `VITE_API_URL`。需要微信登录时，在后端填写真实的 `WECHAT_MINIAPP_APP_ID`、`WECHAT_MINIAPP_APP_SECRET`，不要把 AppSecret 写进小程序源码。

另在服务器项目根目录创建 `.env.production`，只填写域名：

```dotenv
WEB_DOMAIN=www.example.com
API_DOMAIN=api.example.com
```

两处示例域名都要换成自己的真实域名。每台机器使用各自的环境文件，不要提交到 Git。`backend/.dockerignore` 已排除 `.env` 和 `.env.*`，避免后端 Dockerfile 的 `COPY . .` 把运行时密钥送入构建上下文。Git 的 `.gitignore` 不影响 Docker 构建上下文。参见 [Docker 构建上下文文档](https://docs.docker.com/build/concepts/context/)。

后端当前的 CORS 中间件实际返回 `Access-Control-Allow-Origin: *`；`FRONTEND_URL` 虽出现在环境变量示例中，但目前并未被用作 CORS 白名单。不要把填写该变量视为访问控制。

## 五、启动、迁移与验收

确认部署的是准备上线的提交或构建产物：本机未提交的修改不会随 `git clone` 到服务器。初次启动及升级前备份 PostgreSQL 数据卷，并保留 `brand_uploads` 上传卷。数据库、上传文件和 Caddy 证书数据都要纳入备份计划。

```bash
docker compose --env-file .env.production -f docker-compose.yml -f docker-compose.prod.yml up -d --build
docker compose --env-file .env.production -f docker-compose.yml -f docker-compose.prod.yml ps
docker compose --env-file .env.production -f docker-compose.yml -f docker-compose.prod.yml logs --tail=100 caddy backend
```

后端启动时会尝试运行迁移，但部分后续迁移失败只记录警告。因此要查看日志和待迁移清单，不能只凭健康检查判定数据库结构完整。备份数据库后，如有待执行迁移，使用项目已有的严格执行模式：

```bash
docker compose --env-file .env.production -f docker-compose.yml -f docker-compose.prod.yml run --rm backend ./server -migrate-dry-run
docker compose --env-file .env.production -f docker-compose.yml -f docker-compose.prod.yml run --rm backend ./server -migrate-apply
```

从外部网络验证：

```bash
curl -I https://www.example.com
curl https://api.example.com/health
```

第二条应返回包含 `"status":"ok"` 的 JSON。再在浏览器验证首页、登录、一个 `/api` 请求及流式报告；用真机验证小程序登录和头像上传。若证书申请失败，优先检查 DNS、备案／访问状态、80/443 的安全组与系统防火墙、Caddy 日志。

## 六、小程序发布配置

发布版将 `miniapp/config.js` 的 API 地址设为 `https://api.example.com`，不要附加 `/api`；当前文件仍使用局域网 HTTP 地址，且开发者工具分支会切到本机地址，发布前需一并核对。微信后台登记该域名的 request 和 uploadFile 合法域名，并核对小程序 AppID 与后端一致。头像存储在后端 `UPLOAD_DIR/user-avatars`，对应现有 `brand_uploads` 持久化卷。具体发布检查见 [`miniapp/README.md`](../miniapp/README.md)。
