# AGENTS.md — CodeForge Academy / AgenticLearningSystem

本文件用于约束 AI 编码代理在本仓库中的行为，同时作为开发约定的速查表。

## 仓库总览

```
.
├── backend/                  # Go 1.26 + Gin + GORM 后端
│   ├── cmd/server/           # 程序入口
│   ├── internal/
│   │   ├── api/              # 路由 + 处理器（agent / auth / course / apps / tool_planner / tool_helpers）
│   │   ├── config/           # 配置加载（.env）
│   │   ├── database/         # 连接 + 迁移 + 种子
│   │   ├── llm/              # LLM 客户端（SSE 流式）
│   │   ├── model/            # GORM 模型（32 张表）
│   │   ├── sandbox/          # 代码沙箱执行器
│   │   └── storage/          # MinIO 对象存储
│   ├── migrations/           # SQL 迁移
│   ├── prompts/              # 可编辑系统提示词（运行时热加载）
│   └── data/uploads/         # 上传目录（内容不入库，仅保留 .gitkeep）
├── frontend/                 # Next.js 16 + React 19 + TypeScript 5 + Tailwind 4
│   ├── src/app/              # App Router 页面
│   ├── src/components/       # 组件（含 shadcn/ui）
│   ├── documents/            # 后端 API / 数据库 / Agent 设计文档
│   └── lib/api.ts            # API 客户端
├── scripts/                  # PowerShell 启停脚本
├── searxng/                  # SearXNG 联网搜索配置
├── tests/                    # 测试夹具
├── docker-compose.yml        # 基础设施编排（PostgreSQL+pgvector / Redis / MinIO / SearXNG）
├── .env.example              # 环境变量模板
└── .gitignore
```

## 依赖与环境约定（重要）

- **所有需要下载包的环境一律使用项目内 venv / 本地缓存目录，绝不写入全局目录。**
  - Python：使用 `backend/.venv/`（`python -m venv backend/.venv`），不要全局 `pip install`。
  - Go：使用项目内 `GOMODCACHE=backend/.gomodcache`、`GOCACHE=backend/.gocache`、`GOPATH=backend/.gopath`。
  - 前端：使用项目内 `COREPACK_HOME=frontend/.corepack`、`pnpm-store=frontend/.pnpm-store`，通过 `corepack pnpm` 安装。
- 上述缓存目录均已写入 `.gitignore`，不得提交。
- `.env` 含真实密钥，禁止提交；只提交 `.env.example`。

## 启动方式

```powershell
.\scripts\setup.ps1                  # 首次初始化（可加 -SkipFrontendInstall 跳过前端依赖安装）
.\scripts\start.ps1                  # 启动基础设施 + 前后端
.\scripts\status.ps1                 # 查看状态
.\scripts\stop.ps1                   # 停止（加 -Infrastructure 一并停止 Docker）
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
corepack pnpm lint
corepack pnpm exec next build
```

## AI Agent 架构

### 工作流

1. **智能路由**：`routeAgent` 根据关键词 + 对话记忆判断 Agent 类型
2. **工具规划**：`planAgentTools` 使用关键词匹配 + LLM 判断双重机制
3. **工具执行**：`executeWithRetry` 最多 5 次重试，间隔递增；失败后注入失败上下文，让模型用本地知识回答
4. **上下文构建**：`buildConversationPrompt` 拼入对话记忆 + 工具结果
5. **LLM 流式回答**：`StreamChatWithImages` SSE token 实时推送
6. **持久化**：用户消息 + 助手回答 + 工作流 JSON 全部入库

### 对话记忆

- 以 `session_messages` 表为唯一真相来源，不另建内存缓存
- `loadConversationMemory` 读取最近 12 条
- `contextualToolMessage` 在追问时把最近 4 轮记忆注入工具调用
- 系统提示词从 `backend/prompts/system_prompt.md` 读取，热加载，每次注入当前时间

### 系统提示词

编辑 `backend/prompts/system_prompt.md` 即可修改系统提示词，无需重启。文件修改时间变更时自动重载。

## 模型配置

- 默认未配置模型时，Agent 使用本地降级回退，系统仍可运行。
- 配置 Ark（豆包）模型：复制 `.env.example` 为 `.env`，填入 `ARK_API_KEY` / `ARK_MODEL` / `ARK_BASE_URL`。
- `backend/internal/llm/client.go` 在配置完成后自动调用 Ark Chat Completions / Responses API。
- `MaxOutputTokens` 为 4096（图片 2048），temperature 0.7，确保回答充分详细。

## 文档索引

| 你想做什么 | 去看哪个文档 |
|-----------|-------------|
| 了解后端需要实现哪些 API | `frontend/documents/API_SPEC.md` |
| 了解数据库怎么设计 | `frontend/documents/DATABASE_DESIGN.md` |
| 了解 AI Agent 怎么集成 | `frontend/documents/AGENT_DESIGN.md` |
| 了解前端结构与页面 | `frontend/AGENTS.md` |
| 了解完整项目概览 | `README.md` |
