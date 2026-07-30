# CodeForge Academy 全栈系统

已根据 `C:\Users\story\Desktop\完整文档\Agent.md`、模块设计文档和现有前端完成可运行版本。

## 项目结构

- `frontend/`：现成 Next.js 16 + React 19 前端，已接入后端 API 与 SSE。
- `backend/`：Go 1.26 + Gin + GORM 后端。
- `docker-compose.yml`：只部署外部基础设施：PostgreSQL 16 + pgvector、Redis 7、MinIO。
- `backend/.venv/`：项目专用 Python 虚拟环境。
- `backend/.gomodcache/`、`frontend/.pnpm-store/`：项目内依赖缓存，不写入全局项目外目录。

## 默认账号

- 邮箱：`demo@codeforge.local`
- 密码：`Demo123!`

## 启动

在 PowerShell 中执行：

```powershell
cd C:\Users\story\Desktop\学习计划
.\scripts\setup.ps1
.\scripts\start.ps1
```

打开：

- 前端：`http://localhost:5000`
- 后端健康检查：`http://localhost:8080/health`
- MinIO 控制台：`http://localhost:9001`
  - 用户：`codeforge`
  - 密码：`codeforge123`

如果前端依赖已经安装，可以跳过重复安装：

```powershell
.\scripts\setup.ps1 -SkipFrontendInstall
```

查看状态/停止：

```powershell
.\scripts\status.ps1
.\scripts\stop.ps1
# 同时停止 Docker 基础设施
.\scripts\stop.ps1 -Infrastructure
```

## 已完成的功能

- 用户注册、登录、JWT 刷新、个人信息、活动、成就、收藏、笔记。
- 首页课程流、方向课程列表、推荐课程、标签、课程详情、点赞、收藏、评论、阅读进度。
- 算法题库、每日一题、题目详情、代码模板。
- Python / JavaScript / C++ / Rust 代码沙箱运行和提交接口；默认限制系统、网络和危险操作。
- 学习进度、学习时长、学习路径和阶段。
- Agent SSE 流式对话、智能路由、工作流事件、历史、文件上传、Tool 开关、用户画像、知识图谱。
- 简历模板、简历上传、分析、优化和 HTML/DOCX/PDF 导出。
- 项目创建、任务生成、源码上传和 AI 分析建议。
- 笔试生成、历史、题目运行和提交评分。
- 全局搜索接口。

## 当前已配置模型

已从本机 Codex 配置中读取并写入项目 `.env`：

- 模型服务地址：`https://ark.cn-beijing.volces.com/api/v3`
- 模型：`doubao-seed-2-1-pro-260628`
- 协议：Responses
- API Key：已写入 `.env`，不会在文档中明文展示

已经通过真实接口验证，`/api/v1/agent/chat` 可以返回真实模型的 SSE 流式回答。
## AI 配置

默认未配置模型时，Agent 使用本地降级回答，系统仍然可运行。

复制 `.env.example` 为 `.env` 后配置：

```env
ARK_API_KEY=你的 Ark API Key
ARK_MODEL=你的模型推理接入点
ARK_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
```

`backend/internal/llm/client.go` 会在配置完成后自动调用 Ark Chat Completions。

## 验证命令

```powershell
# 后端
cd backend
$env:GOMODCACHE="$PWD\.gomodcache"
$env:GOCACHE="$PWD\.gocache"
$env:GOPATH="$PWD\.gopath"
go test ./...

# 前端
cd ..\frontend
$env:COREPACK_HOME="$PWD\.corepack"
corepack pnpm ts-check
corepack pnpm lint:build
corepack pnpm exec next build
```

当前已验证：后端测试、前端 TypeScript、ESLint、Next.js 生产构建、Docker 健康检查、前端代理、课程/题库/代码运行/简历/项目/SSE 联调。
