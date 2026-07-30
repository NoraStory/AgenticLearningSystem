# AGENTS.md — CodeForge Academy / AgenticLearningSystem

本文件用于约束 AI 编码代理在本仓库中的行为，同时作为开发约定的速查表。

## 仓库总览

```
.
├── backend/            # Go 1.26 + Gin + GORM 后端
│   ├── cmd/server/     # 程序入口
│   ├── internal/       # api / config / database / llm / model / sandbox / storage
│   ├── migrations/     # SQL 迁移
│   └── data/uploads/   # 上传目录（内容不入库，仅保留 .gitkeep）
├── frontend/           # Next.js 16 + React 19 + TypeScript 5 + Tailwind 4
│   ├── src/app/        # App Router 页面
│   ├── src/components/ # 组件（含 shadcn/ui）
│   └── documents/      # 后端 API / 数据库 / Agent 设计文档
├── scripts/            # PowerShell 启停脚本
├── searxng/            # SearXNG 联网搜索配置
├── tests/              # 测试夹具
├── docker-compose.yml  # 仅编排基础设施（PostgreSQL+pgvector / Redis / MinIO）
├── .env.example        # 环境变量模板
└── .gitignore
```

## 依赖与环境约定（重要）

- **所有需要下载包的环境一律使用项目内 venv / 本地缓存目录，绝不写入全局目录。**
  - Python：使用 `backend/.venv/`（`python -m venv backend/.venv`），不要全局 `pip install`。
  - Go：使用项目内 `GOMODCACHE=backend/.gomodcache`、`GOCACHE=backend/.gocache`、`GOPATH=backend/.gopath`。
  - 前端：使用项目内 `COREPACK_HOME=frontend/.corepack`、`pnpm-store=frontend/.pnpm-store`，通过 `corepack pnpm` 安装。
- 上述缓存目录均已写入 `.gitignore`，不得提交。
- `frontend/.env.local`、根目录 `.env` 含真实密钥，禁止提交；只提交 `.env.example`。

## 启动方式

```powershell
# 1) 初始化依赖（首次或拉取后）
.\scripts\setup.ps1                  # 可加 -SkipFrontendInstall 跳过前端依赖安装
# 2) 启动基础设施 + 前后端
.\scripts\start.ps1
# 3) 查看状态 / 停止
.\scripts\status.ps1
.\scripts\stop.ps1                    # 加 -Infrastructure 一并停止 Docker
```

默认账号：`demo@codeforge.local` / `Demo123!`
入口：前端 http://localhost:5000 ；后端健康检查 http://localhost:8080/health

## 验证

```powershell
# 后端测试
cd backend
$env:GOMODCACHE="$PWD\.gomodcache"; $env:GOCACHE="$PWD\.gocache"; $env:GOPATH="$PWD\.gopath"
go test ./...

# 前端
cd ..\frontend
$env:COREPACK_HOME="$PWD\.corepack"
corepack pnpm ts-check
corepack pnpm lint:build
corepack pnpm exec next build
```

## 模型配置

- 默认未配置模型时，Agent 使用本地降级回退，系统仍可运行。
- 配置 Ark（豆包）模型：复制 `.env.example` 为 `.env`，填入 `ARK_API_KEY` / `ARK_MODEL` / `ARK_BASE_URL`。
- `backend/internal/llm/client.go` 在配置完成后自动调用 Ark Chat Completions。

## 文档索引

| 你想做什么 | 去看哪个文档 |
|-----------|-------------|
| 了解后端需要实现哪些 API | `frontend/documents/API_SPEC.md` |
| 了解数据库怎么设计 | `frontend/documents/DATABASE_DESIGN.md` |
| 了解 AI Agent 怎么集成 | `frontend/documents/AGENT_DESIGN.md` |
| 了解前端结构与页面 | `frontend/AGENTS.md` |

README.md 由维护者自行编写，本文件不替代它。
