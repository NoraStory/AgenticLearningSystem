<div align="center">

# CodeForge Academy

### 全栈 AI 智能学习平台

面向开发者的 Python · C++ · 数据库 · 算法 · AI Agent 一体化学习系统

Go 1.26 · Gin · GORM · PostgreSQL + pgvector · Redis · MinIO · SearXNG
Next.js 16 · React 19 · TypeScript 5 · Tailwind CSS 4 · shadcn/ui

</div>

---

## 项目概览

CodeForge Academy 是一个全栈技术学习平台，核心是 **AI Agent 驱动的智能学习助手**。它将课程学习、算法练习、代码沙箱、简历/项目分析、AI 对话等功能整合在一个系统中，后端提供 SSE 流式 Agent 工作流，前端以工作流可视化面板实时展示工具调度与执行过程。

### 核心能力

| 模块 | 功能 |
|------|------|
| **AI Agent 对话** | SSE 流式输出 · 智能路由 · 动态工具规划 · 对话记忆 · 历史会话 · 工作流可视化 |
| **联网搜索** | SearXNG 本地搜索 · 关键词自动提取 · Wikipedia 回退 · AI 判断搜索意图 |
| **工具系统** | 15 个内置工具 · 关键词+AI 双重规划 · 最多 5 次重试 · 失败降级到本地知识 |
| **课程系统** | 方向课程流 · 标签筛选 · 推荐课程 · 详情 · 点赞 · 收藏 · 评论 · 阅读进度 |
| **算法题库** | 题目列表 · 每日一题 · 代码模板 · 在线运行 · 提交评分 |
| **代码沙箱** | Python / JavaScript / C++ / Rust 受限执行 · 系统级沙箱隔离 |
| **简历分析** | 模板 · 上传 · AI 分析 · 优化建议 · HTML/DOCX/PDF 导出 |
| **项目实战** | 自建项目 · 任务生成 · 源码上传 · AI 分析建议 |
| **笔试模拟** | AI 出题 · 题目运行 · 提交评分 · 历史记录 |
| **学习追踪** | 学习进度 · 时长统计 · 连续打卡 · 活动流 · 成就 |
| **用户画像** | 等级 · 聚焦领域 · 薄弱环节 · 学习风格 · 知识图谱 |

## 技术栈

### 后端（`backend/`）

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.26 | 运行时 |
| Gin | 1.10 | HTTP 框架 |
| GORM | 1.25 | ORM |
| PostgreSQL | 16 + pgvector | 主数据库 + 向量检索 |
| Redis | 7.4 | 缓存 |
| MinIO | 对象存储 | 文件/附件 |
| SearXNG | 联网搜索 | Agent 工具 |

### 前端（`frontend/`）

| 技术 | 版本 | 用途 |
|------|------|------|
| Next.js | 16 | 全栈框架 |
| React | 19 | UI |
| TypeScript | 5 | 类型安全 |
| Tailwind CSS | 4 | 样式 |
| shadcn/ui | latest | 组件库 |
| pnpm | 9 | 包管理 |

### 基础设施（`docker-compose.yml`）

| 服务 | 镜像 | 端口 |
|------|------|------|
| PostgreSQL + pgvector | `pgvector/pgvector:pg16` | 5432 |
| Redis | `redis:7.4-alpine` | 6379 |
| SearXNG | `searxng/searxng:latest` | 8081 |
| MinIO | `minio/minio` | 9000 / 9001 |

## 快速开始

### 环境要求

- Go ≥ 1.26
- Node.js ≥ 20 + pnpm ≥ 9（通过 corepack）
- Docker Desktop（用于基础设施）
- PowerShell（启停脚本基于 PS）

### 1. 克隆仓库

```bash
git clone https://github.com/NoraStory/AgenticLearningSystem.git
cd AgenticLearningSystem
```

### 2. 配置环境变量

```powershell
Copy-Item .env.example .env
```

编辑 `.env`，按需填写 Ark（豆包）模型配置。留空时 Agent 使用本地降级回退，系统仍可运行。

```env
# 基础设施
POSTGRES_DB=codeforge
POSTGRES_USER=codeforge
POSTGRES_PASSWORD=codeforge

# 后端
PORT=8080
DATABASE_URL=postgres://codeforge:codeforge@localhost:5432/codeforge?sslmode=disable
REDIS_ADDR=localhost:6379
JWT_SECRET=local-development-change-this-secret
CORS_ORIGINS=http://localhost:5000,http://127.0.0.1:5000

# 可选：Ark 模型（留空时使用本地降级回退）
ARK_API_KEY=
ARK_MODEL=doubao-seed-2-1-pro-260628
ARK_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
LLM_WIRE_API=responses

# 联网搜索
SEARCH_PROVIDER=searxng
SEARCH_BASE_URL=http://localhost:8081
SEARCH_TIMEOUT_SECONDS=15
```

### 3. 启动

```powershell
# 首次：初始化依赖 + 启动基础设施 + 启动前后端
.\scripts\setup.ps1
.\scripts\start.ps1

# 已安装过依赖，跳过前端安装
.\scripts\setup.ps1 -SkipFrontendInstall
```

### 4. 访问

| 入口 | 地址 |
|------|------|
| 前端 | http://localhost:5000 |
| 后端健康检查 | http://localhost:8080/health |
| MinIO 控制台 | http://localhost:9001（codeforge / codeforge123） |

### 默认账号

| 邮箱 | 密码 |
|------|------|
| `demo@codeforge.local` | `Demo123!` |

### 状态管理

```powershell
.\scripts\status.ps1          # 查看运行状态
.\scripts\stop.ps1            # 停止前后端
.\scripts\stop.ps1 -Infrastructure  # 同时停止 Docker 基础设施
```

## 项目结构

```
AgenticLearningSystem/
├── backend/                        # Go 后端
│   ├── cmd/server/main.go          # 入口
│   ├── internal/
│   │   ├── api/                    # HTTP 路由 + 处理器
│   │   │   ├── agent.go            # Agent 对话 + SSE 工作流
│   │   │   ├── agent_memory.go     # 对话记忆 + 上下文构建
│   │   │   ├── agent_sessions.go   # 会话列表 + 删除
│   │   │   ├── tool_planner.go     # 工具规划 + 执行器
│   │   │   ├── tool_helpers.go     # 重试 + 关键词提取 + 提示词
│   │   │   ├── server.go           # 路由注册 + 中间件
│   │   │   ├── auth_user.go        # 认证 + 用户
│   │   │   ├── course_problem.go   # 课程 + 题目
│   │   │   ├── learning.go         # 学习路径 + 进度
│   │   │   ├── apps.go             # 简历 + 项目 + 笔试
│   │   │   └── ...
│   │   ├── config/config.go        # 配置加载
│   │   ├── database/               # 数据库连接 + 种子数据
│   │   ├── llm/client.go           # LLM 客户端（SSE 流式）
│   │   ├── model/models.go         # GORM 模型（32 张表）
│   │   ├── sandbox/runner.go       # 代码沙箱执行器
│   │   └── storage/store.go        # MinIO 对象存储
│   ├── migrations/                 # SQL 迁移
│   ├── prompts/
│   │   └── system_prompt.md        # 可编辑系统提示词（运行时热加载）
│   ├── data/uploads/               # 上传目录
│   ├── go.mod / go.sum
│   └── .gomodcache/ .gocache/      # 项目内 Go 缓存（不入全局）
│
├── frontend/                       # Next.js 前端
│   ├── src/
│   │   ├── app/                    # App Router 页面
│   │   │   ├── agent/chat/         # AI 对话页（工作流可视化）
│   │   │   ├── agent/profile/      # 用户画像 + 知识图谱
│   │   │   ├── agent/tools/        # 工具管理
│   │   │   ├── python/ cpp/ database/ algorithm/
│   │   │   ├── course/[id]/        # 课程详情
│   │   │   ├── practice/           # 在线练习
│   │   │   ├── learning-path/      # 学习路径
│   │   │   ├── interview/          # 笔试模拟
│   │   │   ├── project/            # 项目实战
│   │   │   ├── resume/             # 简历分析
│   │   │   └── profile/            # 个人中心
│   │   ├── components/layout/      # Header + Sidebar
│   │   ├── lib/api.ts             # API 客户端
│   │   └── hooks/
│   ├── documents/                   # 设计文档
│   │   ├── API_SPEC.md
│   │   ├── DATABASE_DESIGN.md
│   │   └── AGENT_DESIGN.md
│   ├── public/
│   ├── package.json
│   └── pnpm-lock.yaml
│
├── scripts/                        # PowerShell 启停脚本
│   ├── setup.ps1                   # 初始化依赖
│   ├── start.ps1                   # 启动
│   ├── status.ps1                  # 状态
│   └── stop.ps1                    # 停止
│
├── searxng/settings.yml            # SearXNG 配置
├── tests/fixtures/                 # 测试夹具
├── docker-compose.yml              # 基础设施编排
├── .env.example                    # 环境变量模板
├── .gitignore
├── AGENTS.md                       # AI 编码约定
└── README.md
```

## AI Agent 架构

### 工作流

```
用户提问
  │
  ▼
智能路由（routeAgent）── 根据关键词 + 对话记忆判断 Agent 类型
  │
  ▼
工具规划（planAgentTools）── 关键词匹配 + LLM 判断双重机制
  │
  ├── 联网搜索（SearXNG → Wikipedia 回退）
  ├── 文档阅读（图片 / 文本 / 代码附件）
  ├── 代码检索 / 执行 / 自修复
  ├── SQL 分析 / 图表生成 / 测验生成
  ├── 课程检索 / 进度查询
  ├── 简历审阅 / 项目审阅
  └── Git 助手 / 思维导图
  │
  ▼
工具执行（executeWithRetry）── 最多 5 次重试，间隔递增
  │
  ▼
上下文构建（buildConversationPrompt）── 记忆 + 工具结果 + 时间注入
  │
  ▼
LLM 流式回答（StreamChatWithImages）── SSE token 实时推送
  │
  ▼
持久化 ── 用户消息 + 助手回答 + 工作流 JSON 全部入库
```

### 对话记忆

- 每轮对话的用户消息和助手回答都持久化到 `session_messages` 表
- `loadConversationMemory` 从数据库读取最近 12 条作为工作记忆
- `buildConversationPrompt` 把记忆拼入模型提示，解析"这个/它/上面"等指代
- `contextualToolMessage` 在追问时把最近 4 轮记忆注入工具调用
- 重启或刷新后从持久化数据完整恢复上下文

### 系统提示词

系统提示词从 `backend/prompts/system_prompt.md` 读取，带文件缓存和修改时间检测——**编辑文件后自动热加载，无需重启**。每次请求自动注入当前时间（Asia/Shanghai）。

### 工具重试机制

- 工具调用失败时自动重试，最多 5 次，间隔递增（1s→2s→3s→4s→5s）
- 前端工作流面板实时显示失败次数（已重试 X/5 次）、错误详情、累计失败次数
- 超过上限后停止重试，注入失败上下文让模型用本地知识回答

## 验证

```powershell
# 后端测试
cd backend
$env:GOMODCACHE="$PWD\.gomodcache"; $env:GOCACHE="$PWD\.gocache"; $env:GOPATH="$PWD\.gopath"
go test ./...

# 前端
cd ..\frontend
$env:COREPACK_HOME="$PWD\.corepack"
corepack pnpm ts-check       # TypeScript 类型检查
corepack pnpm lint           # ESLint + Stylelint
corepack pnpm exec next build  # 生产构建
```

## 依赖环境约定

> **所有包均使用项目内 venv / 本地缓存，绝不写入全局目录。**

| 运行时 | 缓存目录 |
|--------|----------|
| Python | `backend/.venv/` |
| Go | `backend/.gomodcache/` `backend/.gocache/` `backend/.gopath/` |
| 前端 | `frontend/.corepack/` `frontend/.pnpm-store/` |

上述目录均已写入 `.gitignore`，不会提交到仓库。

## 设计文档

| 文档 | 路径 |
|------|------|
| 后端 API 接口 | `frontend/documents/API_SPEC.md` |
| 数据库设计 | `frontend/documents/DATABASE_DESIGN.md` |
| AI Agent 设计 | `frontend/documents/AGENT_DESIGN.md` |
| AI 编码约定 | `AGENTS.md` |

## License

本项目仅用于学习和个人使用。
