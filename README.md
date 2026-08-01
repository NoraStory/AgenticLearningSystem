<p align="center">
  <img src="docs/banner.svg" width="100%" alt="CodeForge Academy Banner"/>
</p>

<h1 align="center">CodeForge Academy</h1>

<p align="center">
  <strong>面向开发者的 AI 智能学习平台</strong>
  <br/>
  <em>让 AI 成为你的学习搭档 —— 会思考、会查资料、会写代码、会帮你优化简历</em>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go"/>
  <img src="https://img.shields.io/badge/Next.js-16-black?logo=next.js&logoColor=white" alt="Next.js"/>
  <img src="https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black" alt="React"/>
  <img src="https://img.shields.io/badge/TypeScript-5-3178C6?logo=typescript&logoColor=white" alt="TypeScript"/>
  <img src="https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white" alt="PostgreSQL"/>
  <img src="https://img.shields.io/badge/Redis-7.4-DC382D?logo=redis&logoColor=white" alt="Redis"/>
  <img src="https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white" alt="Docker"/>
  <img src="https://img.shields.io/badge/Eino-0.9-4EC9B0?logo=cloud&logoColor=white" alt="Eino"/>
</p>

<p align="center">
  <a href="#✨-特性亮点">✨ 特性亮点</a> ·
  <a href="#🚀-快速开始">🚀 快速开始</a> ·
  <a href="#🤖-agent-架构">🤖 Agent 架构</a> ·
  <a href="#⚡-技术栈">⚡ 技术栈</a> ·
  <a href="#📁-项目结构">📁 项目结构</a> ·
  <a href="#✅-验证">✅ 验证</a>
</p>

---

## ✨ 特性亮点

<table>
<tr>
<td width="50%" align="center">
<h3>🤖 真·智能 Agent</h3>
<p align="left">
不是关键词模板机。<br/>
基于 <strong>Eino</strong> 工具循环，模型自主决策调用哪个工具、调几次、何时总结 —— 像真人一样"查资料 → 分析 → 回答"。
</p>
</td>
<td width="50%" align="center">
<h3>📄 简历按你的模板优化</h3>
<p align="left">
上传你自己的 DOCX 简历模板，AI 优化内容后 <strong>按原模板样式</strong> 输出 docx / pdf —— 版式一个像素都不丢（docxtpl + Gotenberg）。
</p>
</td>
</tr>
<tr>
<td width="50%" align="center">
<h3>🧪 学练闭环</h3>
<p align="left">
课程 → 练习 → 笔试 → 项目,全链路真实。
笔试 AI 出题 + AI 评分 + Monaco 在线写码；代码沙箱跑真代码,不是选择题。
</p>
</td>
<td width="50%" align="center">
<h3>📊 你的画像,AI 记得</h3>
<p align="left">
BKT 知识追踪(参数可按分类落库、离线拟合) + ECharts 知识图谱 + 雷达图 + 热力图 + 学习洞察。系统知道你会什么、弱在哪,画像驱动 Agent 回答与个性化推荐。
</p>
</td>
</tr>
</table>

---

## 🚀 快速开始

### 环境要求

> **Go ≥ 1.26** · **Node.js ≥ 20** + pnpm 9 · **Docker Desktop** · PowerShell

### 1. 克隆 & 配置

```bash
git clone https://github.com/NoraStory/AgenticLearningSystem.git
cd AgenticLearningSystem
```

```powershell
Copy-Item .env.example .env   # 填写 ARK 模型 Key（留空时系统降级运行，不影响其他功能）
```

```env
# 可选：Ark 模型（豆包）。留空时 Agent 使用本地降级回退
ARK_API_KEY=
ARK_MODEL=doubao-seed-2-1-pro-260628
ARK_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
LLM_WIRE_API=responses

# 联网搜索（SearXNG 本地实例）
SEARCH_PROVIDER=searxng
SEARCH_BASE_URL=http://localhost:8081
```

### 2. 启动

```powershell
.\scripts\setup.ps1    # 首次初始化依赖（Go / 前端 / Python）
.\scripts\start.ps1    # 启动基础设施 + 前后端
```

| 入口 | 地址 |
|------|------|
| 🖥 前端 | http://localhost:5000 |
| 🔌 后端 API | http://localhost:8080/health |
| 📦 MinIO 控制台 | http://localhost:9001 |

> **账号**:右上角「登录 / 注册」(http://localhost:5000/auth),密码 Argon2id 加密存储 · 未登录以游客身份使用(数据为 Demo 用户) · `.\scripts\status.ps1` 查看状态　·　`.\scripts\stop.ps1` 停止

---

## 🤖 Agent 架构

```
用户提问
   │
   ▼
┌──────────────────┐
│   智能路由         │  关键词 + 对话记忆 → 6 类 Agent
└──────┬───────────┘
       ▼
┌──────────────────┐
│   Eino 工具循环    │  模型自主决策工具调用序列（ReAct 风格）
│  (runAgentLoop)   │  eino AgenticModel（豆包 Ark Responses API）
│                   │  → 决策 → 执行工具 → 结果回填 → 继续决策
└──────┬───────────┘
       ▼
┌──────────────────┐
│   工具执行         │  15 个内置工具（复用 executeWithRetry 重试）
│ (executeAgentTool)│  失败后模型诚实告知并用已有知识回答
└──────┬───────────┘
       ▼
┌──────────────────┐
│   上下文构建       │  对话记忆(12条) + 当前时间
└──────┬───────────┘
       ▼
┌──────────────────┐
│   LLM 流式回答     │  SSE token 实时推送 → 前端逐字渲染
└──────┬───────────┘
       ▼
┌──────────────────┐
│   持久化           │  用户消息 + 助手回答 + 工作流 JSON → 数据库
└──────────────────┘
```

<details>
<summary><b>🛠 内置工具清单（13 个）</b></summary>

| 工具 | 分类 | 功能 |
|------|------|------|
| `web_search` | 信息获取 | SearXNG 联网搜索，Wikipedia 回退（10 分钟结果缓存） |
| `doc_reader` | 信息获取 | 图片 / 文本 / 代码附件解析 |
| `code_search` | 开发工具 | 附件或工作区中查找定义与引用 |
| `git_helper` | 开发工具 | 读取工作区状态，生成提交说明 |
| `code_execute` | 开发工具 | Piston 容器隔离执行（不可用时降级本地沙箱） |
| `self_heal` | 开发工具 | 分析错误并提出最小修复 |
| `sql_explain` | 开发工具 | PostgreSQL EXPLAIN 真执行计划分析（数据库不可达时回退静态规则） |
| `quiz_gen` | 学习工具 | 按知识点生成练习题 |
| `leetcode_fetch` | 学习工具 | 站内题库检索 |
| `course_search` | 学习工具 | 站内课程和章节检索 |
| `progress_query` | 学习工具 | 读取学习进度并生成建议 |
| `resume_review` | 职业工具 | AI 简历分析（复用简历模块真实 LLM 分析） |
| `project_review` | 职业工具 | AI 项目源码评审（LLM 结构/风险/建议） |

</details>

<details>
<summary><b>🧠 对话记忆机制</b></summary>

- **唯一真相来源**：每轮对话持久化到 PostgreSQL，重启 / 刷新后完整恢复
- **工作记忆**：读取最近 12 条，`buildConversationPrompt` 拼入提示
- **指代解析**：追问时自动注入最近 4 轮记忆，正确解析"这个 / 它 / 上面"
- **会话隔离**：每个 `session_id` 独立，切换会话互不影响
- **提示词热加载**：编辑 `backend/prompts/system_prompt.md` 即时生效，无需重启

</details>

---

## ⚡ 技术栈

| 层 | 技术 |
|------|------|
| **后端** | Go 1.26 · Gin 1.10 · GORM 1.25 · JWT v5 · MinIO Go v7 · [Eino 0.9](https://github.com/cloudwego/eino)（Agent 编排） |
| **前端** | Next.js 16 · React 19 · TypeScript 5 · Tailwind CSS 4 · shadcn/ui · recharts + ECharts（图表）· Monaco（代码编辑器） |
| **基础设施** | PostgreSQL 16 + pgvector（5432）· Redis 7.4（6379）· SearXNG（8081）· MinIO（9000/9001）· Gotenberg docx→PDF（3000）· Piston 代码执行（2000）· 全部 Docker 编排 |
| **AI** | 火山引擎 Ark / 豆包（Responses API）· docxtpl（用户模板渲染）· shiki（代码高亮）· Mermaid 11（图表） |

---

## 📁 项目结构

```
AgenticLearningSystem/
├── backend/                        # 🦫 Go 后端
│   ├── cmd/server/main.go          #    入口
│   ├── internal/
│   │   ├── api/                    #    路由 + 处理器
│   │   │   ├── agent.go            #      Agent SSE 工作流
│   │   │   ├── agent_eino.go       #      Eino 工具循环（模型自主决策）
│   │   │   ├── agent_memory.go     #      对话记忆 + 上下文
│   │   │   ├── agent_sessions.go   #      会话列表 + 删除
│   │   │   ├── tool_planner.go     #      工具执行器 + 重试
│   │   │   ├── tool_helpers.go     #      重试 + 关键词 + 提示词
│   │   │   └── server.go           #      路由 + 中间件
│   │   ├── config/                 #    配置加载
│   │   ├── database/               #    连接 + 迁移 + 种子
│   │   ├── llm/client.go           #    LLM SSE 流式客户端
│   │   ├── model/models.go         #    GORM 模型（32 表）
│   │   ├── sandbox/runner.go       #    代码沙箱
│   │   └── storage/store.go        #    MinIO 存储
│   ├── prompts/
│   │   └── system_prompt.md        #    可编辑提示词（热加载）
│   └── migrations/                 #    SQL 迁移
│
├── frontend/                       # ⚛️ Next.js 前端
│   ├── src/app/                    #    App Router 页面
│   │   ├── agent/chat/             #      AI 对话 + 工作流可视化
│   │   ├── agent/profile/          #      用户画像 + 知识图谱
│   │   ├── agent/tools/            #      工具管理
│   │   ├── python/ cpp/ database/
│   │   │   algorithm/              #      五大学习方向
│   │   ├── course/ practice/
│   │   │   learning-path/          #      课程 / 练习 / 路径
│   │   ├── interview/ project/
│   │   │   resume/                 #      笔试 / 项目 / 简历
│   │   └── profile/                #      个人中心
│   ├── documents/                  #    设计文档
│   └── lib/api.ts                  #    API 客户端
│
├── scripts/                        # 🔧 PowerShell 脚本
├── searxng/settings.yml            # 🔍 SearXNG 配置
├── tests/fixtures/                 # 🧪 测试夹具
├── docs/banner.svg                 # 🎨 README Banner
├── docker-compose.yml              # 🐳 基础设施编排
├── .env.example                    # 📋 环境变量模板
├── AGENTS.md                       # 🤖 AI 编码约定
└── README.md                       # 📖 你在这里
```

---

## ✅ 验证

```powershell
# 后端
cd backend
$env:GOMODCACHE="$PWD\.gomodcache"; $env:GOCACHE="$PWD\.gocache"; $env:GOPATH="$PWD\.gopath"
go test ./...                          # 单元测试

# 前端
cd ..\frontend
$env:COREPACK_HOME="$PWD\.corepack"
corepack pnpm ts-check                 # TypeScript 类型检查
corepack pnpm lint                     # ESLint + Stylelint
corepack pnpm exec next build          # 生产构建
```

> **依赖环境约定**：所有依赖均使用项目内缓存目录（`backend/.gomodcache/`、`backend/.venv/`、`frontend/.pnpm-store/` 等），绝不写入全局，已全部加入 `.gitignore`。

---

## 📚 设计文档

| 文档 | 路径 |
|------|------|
| 后端 API 接口 | [`frontend/documents/API_SPEC.md`](frontend/documents/API_SPEC.md) |
| 数据库设计 | [`frontend/documents/DATABASE_DESIGN.md`](frontend/documents/DATABASE_DESIGN.md) |
| AI Agent 设计 | [`frontend/documents/AGENT_DESIGN.md`](frontend/documents/AGENT_DESIGN.md) |
| AI 编码约定 | [`AGENTS.md`](AGENTS.md) |
| 功能优化与开源替代清单 | [`docs/功能优化与开源替代清单.md`](docs/功能优化与开源替代清单.md) |
| AI 助手工作流改造方案 | [`docs/AI助手工作流改造方案.md`](docs/AI助手工作流改造方案.md) |
| 用户画像优化与埋点分析 | [`docs/用户画像优化与埋点分析.md`](docs/用户画像优化与埋点分析.md) |

---

<p align="center">
  <sub>Built with ❤️ by <a href="https://github.com/NoraStory">NoraStory</a></sub><br/>
  <sub>本项目仅用于学习和个人使用</sub>
</p>
