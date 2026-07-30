# CodeForge Academy - 技术学习平台

> 面向开发者的 Python、C++、数据库、数据结构与算法、AI Agent 学习平台

## 项目概览

这是一个技术学习平台的前端项目，采用 Next.js 16 + React 19 + TypeScript 5 + Tailwind CSS 4 构建。

**五大学习方向**：
- **Python 编程**：从入门到精通，掌握 Python 核心语法与实战应用
- **C++ 编程**：系统级编程利器，从基础语法到高性能开发
- **数据库**：掌握关系型与非关系型数据库，从 SQL 到数据库设计
- **数据结构与算法**：LeetCode 风格的题目练习，覆盖数组、链表、树、动态规划等
- **AI Agent 开发**：LangChain、LangGraph、RAG 等 Agent 开发技术

## 快速开始

```bash
# 安装依赖
pnpm install

# 开发模式
pnpm dev

# 构建
pnpm build

# 生产模式
pnpm start
```

## 目录结构

```
├── documents/                    # 项目文档（AI 可读）
│   ├── API_SPEC.md              # 后端 API 接口文档
│   ├── DATABASE_DESIGN.md       # 数据库选型与表结构设计
│   └── AGENT_DESIGN.md          # AI Agent 集成设计
├── src/
│   ├── app/                     # Next.js App Router 页面
│   │   ├── page.tsx             # 首页（博客文章流）
│   │   ├── python/page.tsx      # Python 学习页
│   │   ├── cpp/page.tsx         # C++ 学习页
│   │   ├── database/page.tsx    # 数据库学习页
│   │   ├── algorithm/page.tsx   # 算法学习页
│   │   ├── agent/page.tsx       # Agent 学习页
│   │   ├── agent/chat/page.tsx  # Agent 智能对话页（含工作流展示）
│   │   ├── agent/tools/page.tsx # Tool 插件管理页
│   │   ├── agent/profile/page.tsx # 用户画像页（含知识图谱）
│   │   ├── interview/page.tsx   # 笔试模拟页
│   │   ├── project/page.tsx     # 项目实战页（自建项目+AI分析）
│   │   ├── resume/page.tsx      # 简历分析与优化页
│   │   ├── course/[id]/page.tsx # 课程详情页
│   │   ├── practice/page.tsx    # 在线练习页
│   │   ├── learning-path/page.tsx # 学习路径页
│   │   ├── profile/page.tsx     # 个人中心页
│   │   ├── globals.css          # 全局样式（Design Token）
│   │   └── layout.tsx           # 根布局
│   ├── components/
│   │   └── layout/
│   │       ├── Header.tsx       # 顶栏组件
│   │       └── Sidebar.tsx      # 侧边栏组件
│   └── lib/
│       └── utils.ts             # 工具函数
├── .cozeproj/
│   └── prototype/web/           # 原型文件（HTML）
├── package.json
└── tsconfig.json
```

## 文档索引

### 给 AI 的快速导航

| 你想做什么 | 去看哪个文档 |
|-----------|-------------|
| 了解后端需要实现哪些 API | `documents/API_SPEC.md` |
| 了解数据库怎么设计 | `documents/DATABASE_DESIGN.md` |
| 了解 AI Agent 怎么集成 | `documents/AGENT_DESIGN.md` |
| 了解页面结构和样式 | `.cozeproj/prototype/web/*.html` |
| 了解 Design Token | `src/app/globals.css` |

### 文档详细说明

#### `documents/API_SPEC.md`
- 定义了所有后端 API 接口
- 包含：认证、用户、课程、题目、进度、代码执行、评论、收藏、笔记
- 每个接口都有请求/响应示例
- 包含学习资源来源链接

#### `documents/DATABASE_DESIGN.md`
- 数据库选型：PostgreSQL + Redis + pgvector
- 完整的表结构设计（11 张表）
- 索引设计建议
- Prisma Schema 示例
- Redis 缓存设计

#### `documents/AGENT_DESIGN.md`
- AI Agent 功能场景
- Agent 架构设计图
- 三个 Agent 模块：学习助手、代码审查、题目讲解
- 技术选型：LangChain + LangGraph
- 实现代码示例
- 知识库来源

## 页面说明

| 页面 | 路径 | 说明 |
|------|------|------|
| 首页 | `/` | 博客文章流展示最新课程，右侧边栏含学习进度、每日一题、标签云 |
| Python 学习 | `/python` | Python 课程列表，博客卡片风格，支持按难度/状态筛选 |
| C++ 学习 | `/cpp` | C++ 课程列表，博客卡片风格，支持按难度/状态筛选 |
| 数据库学习 | `/database` | 数据库课程列表，博客卡片风格，支持按难度/状态筛选 |
| 算法学习 | `/algorithm` | 算法题目列表，LeetCode 风格，支持按分类/难度筛选 |
| Agent 学习 | `/agent` | AI Agent 课程模块，含实战项目推荐 |
| Agent 对话 | `/agent/chat` | AI 智能对话页，SSE 流式输出，6个Agent智能路由，工作流执行状态展示，支持上传图片/文件 |
| Tool 管理 | `/agent/tools` | 15 个 Tool 插件开关管理面板 |
| 用户画像 | `/agent/profile` | 学习画像、能力雷达图、知识图谱可视化、掌握状态 |
| 笔试模拟 | `/interview` | 代码题（LeetCode风格编辑器）+ 问答题（文本框），AI评分讲解 |
| 项目实战 | `/project` | 自建项目、AI生成任务清单、上传源码、AI分析完成情况 |
| 简历分析 | `/resume` | 上传简历AI分析、选择模板AI优化、导出PDF/DOCX/HTML |
| 课程详情 | `/course/[id]` | 文章详情页，含代码块、TOC 目录、评论区 |
| 在线练习 | `/practice` | 代码练习页，左右分栏（题目描述 + 代码编辑器） |
| 学习路径 | `/learning-path` | 学习路线图，时间轴展示各阶段 |
| 个人中心 | `/profile` | 用户信息、学习统计、成就徽章、收藏、笔记 |

## 设计风格

采用**技术博客风格**，参考 lololowe.com 的设计：
- 内容优先、排版松散、留白充足
- 课程卡片有封面大图、分类标签、摘要
- 右侧边栏含小组件（进度、每日一题、标签云）
- 配色：蓝色主色（#2563EB），白色背景

## 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Next.js | 16 | React 框架（App Router） |
| React | 19 | UI 库 |
| TypeScript | 5 | 类型安全 |
| Tailwind CSS | 4 | 样式 |
| Lucide React | - | 图标 |

## 后端开发指引

如果后端开发者要接入，请按以下顺序阅读文档：

1. **先看 API 文档**：`documents/API_SPEC.md` - 了解所有接口定义
2. **再看数据库文档**：`documents/DATABASE_DESIGN.md` - 了解表结构
3. **最后看 Agent 文档**：`documents/AGENT_DESIGN.md` - 了解 AI 功能

前端调用 API 的示例代码在各页面组件中，搜索 `fetch('/api/` 即可找到。

## 学习资源来源

### Python
- Python 官方文档: https://docs.python.org/zh-cn/3/
- Real Python: https://realpython.com/
- Python Tutorial: https://www.pythontutorial.net/

### C++
- C++ Primer: https://www.cppprimer.com/
- cppreference.com: https://en.cppreference.com/
- C++ Core Guidelines: https://isocpp.github.io/CppCoreGuidelines/

### 数据库
- PostgreSQL 官方文档: https://www.postgresql.org/docs/
- Redis 官方文档: https://redis.io/docs/
- MongoDB 官方文档: https://www.mongodb.com/docs/
- SQLZoo: https://www.sqlzoo.net/

### 算法
- LeetCode: https://leetcode.cn/
- Visualgo: https://visualgo.net/

### AI Agent
- LangChain: https://python.langchain.com/
- LangGraph: https://langchain-ai.github.io/langgraph/
- OpenAI API: https://platform.openai.com/docs
