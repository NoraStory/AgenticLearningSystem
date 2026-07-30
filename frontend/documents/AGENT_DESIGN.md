# AI Agent 集成设计文档

> CodeForge Academy AI Agent 功能设计
>
> 本文档定义了学习平台中 AI Agent 的集成方案。

## 目录

- [1. Agent 功能概述](#1-agent-功能概述)
- [2. Agent 架构设计](#2-agent-架构设计)
- [3. Agent 功能模块](#3-agent-功能模块)
- [4. 技术选型](#4-技术选型)
- [5. 实现方案](#5-实现方案)

---

## 1. Agent 功能概述

### 平台中的 AI Agent 场景

| 场景 | 描述 | 用户价值 |
|------|------|----------|
| **学习助手** | 解答课程中的技术问题，提供代码示例 | 即时答疑，降低学习门槛 |
| **代码审查** | 分析用户提交的代码，给出优化建议 | 提升代码质量 |
| **题目讲解** | 解释算法题目思路，提供解题提示 | 辅助算法学习 |
| **学习规划** | 根据用户进度推荐学习路径 | 个性化学习体验 |
| **知识检索** | 从课程库中检索相关内容 | 快速定位知识 |

---

## 2. Agent 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                      前端 (Next.js)                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │ 聊天界面  │  │ 代码助手  │  │ 题目讲解  │              │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘              │
└───────┼──────────────┼──────────────┼───────────────────┘
        │              │              │
        ▼              ▼              ▼
┌─────────────────────────────────────────────────────────┐
│                    API Gateway                           │
└────────────────────────┬────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        ▼                ▼                ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ 学习助手 Agent│ │ 代码审查 Agent│ │ 题目讲解 Agent│
│              │ │              │ │              │
│ - RAG 检索   │ │ - 代码分析   │ │ - 思路引导   │
│ - 上下文管理  │ │ - 优化建议   │ │ - 示例代码   │
│ - 多轮对话   │ │ - 最佳实践   │ │ - 复杂度分析  │
└──────┬───────┘ └──────┬───────┘ └──────┬───────┘
       │                │                │
       ▼                ▼                ▼
┌─────────────────────────────────────────────────────────┐
│                    工具层 (Tools)                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │ 课程检索  │  │ 代码执行  │  │ 进度查询  │              │
│  └──────────┘  └──────────┘  └──────────┘              │
└─────────────────────────────────────────────────────────┘
       │                │                │
       ▼                ▼                ▼
┌─────────────────────────────────────────────────────────┐
│                    数据层                                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │PostgreSQL│  │  Redis   │  │pgvector  │              │
│  └──────────┘  └──────────┘  └──────────┘              │
└─────────────────────────────────────────────────────────┘
```

---

## 3. Agent 功能模块

### 3.1 学习助手 Agent

**功能**：
- 回答课程相关的技术问题
- 提供代码示例和解释
- 推荐相关学习内容
- 记录学习问答历史

**工具调用**：
- `search_courses(query)` - 搜索相关课程
- `get_course_content(course_id, section_id)` - 获取课程内容
- `get_user_progress(user_id)` - 获取用户学习进度
- `recommend_courses(user_id, topic)` - 推荐课程

**Prompt 模板**：
```
你是一个 Python/C++/数据库/算法/Agent 学习助手。你的任务是帮助用户理解技术概念、解答问题。

当前用户正在学习：{current_course}
用户的问题：{user_question}

请：
1. 用简洁清晰的语言解释概念
2. 提供代码示例（如果适用）
3. 推荐相关的学习资源
4. 鼓励用户继续学习
```

### 3.2 代码审查 Agent

**功能**：
- 分析用户提交的代码
- 指出潜在问题（性能、安全、风格）
- 提供优化建议
- 解释最佳实践

**工具调用**：
- `analyze_code(code, language)` - 分析代码
- `get_best_practices(language, topic)` - 获取最佳实践
- `run_tests(code, test_cases)` - 运行测试

**Prompt 模板**：
```
你是一个代码审查专家。请分析以下 {language} 代码：

```
{user_code}
```

请从以下方面给出建议：
1. 正确性：是否有逻辑错误？
2. 性能：是否有优化空间？
3. 可读性：代码是否清晰易懂？
4. 最佳实践：是否符合 {language} 的惯用写法？
```

### 3.3 题目讲解 Agent

**功能**：
- 解释算法题目的解题思路
- 提供渐进式提示（不直接给答案）
- 分析时间/空间复杂度
- 推荐相关题目

**工具调用**：
- `get_problem(problem_id)` - 获取题目信息
- `get_similar_problems(problem_id)` - 获取相关题目
- `get_solution_hints(problem_id, level)` - 获取提示

**交互模式**：
```
用户：这道题怎么做？

Agent：让我们一步步来分析这道题。

首先，题目要求我们...（复述题意）

你能想到用什么数据结构来解决吗？
提示：考虑一下查找操作的效率。

[用户回答后继续引导...]
```

---

## 4. 技术选型

### 推荐方案：LangChain + LangGraph

| 组件 | 选择 | 理由 |
|------|------|------|
| Agent 框架 | LangChain / LangGraph | 成熟的 Agent 框架，支持工具调用、记忆管理 |
| LLM | OpenAI GPT-4 / Claude | 强大的代码理解和生成能力 |
| 向量数据库 | pgvector | 与 PostgreSQL 集成，存储课程向量嵌入 |
| Embedding | OpenAI text-embedding-3-small | 课程内容的向量化 |
| 会话管理 | Redis | 存储对话历史和上下文 |

### 备选方案

| 方案 | 优点 | 缺点 |
|------|------|------|
| AutoGen | 多 Agent 协作 | 复杂度高 |
| Dify | 低代码平台 | 定制性有限 |
| 自建 Agent | 完全控制 | 开发成本高 |

---

## 5. 实现方案

### 5.1 课程向量化

```python
# 课程文档处理流程
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain_openai import OpenAIEmbeddings
from langchain_community.vectorstores import PGVector

# 1. 加载课程内容
courses = load_courses_from_db()

# 2. 文本分割
text_splitter = RecursiveCharacterTextSplitter(
    chunk_size=1000,
    chunk_overlap=200,
    separators=["\n## ", "\n### ", "\n\n", "\n", " "]
)

# 3. 创建向量存储
embeddings = OpenAIEmbeddings(model="text-embedding-3-small")
vectorstore = PGVector(
    connection_string="postgresql://...",
    collection_name="courses",
    embedding_function=embeddings
)

# 4. 添加文档
for course in courses:
    chunks = text_splitter.split_text(course.content)
    vectorstore.add_texts(
        texts=chunks,
        metadatas=[{"course_id": course.id, "title": course.title}] * len(chunks)
    )
```

### 5.2 Agent 实现示例

```python
from langchain.agents import AgentExecutor, create_openai_tools_agent
from langchain_openai import ChatOpenAI
from langchain.tools import tool

@tool
def search_courses(query: str) -> str:
    """搜索相关课程内容"""
    results = vectorstore.similarity_search(query, k=3)
    return format_results(results)

@tool
def get_user_progress(user_id: str) -> str:
    """获取用户学习进度"""
    progress = db.get_progress(user_id)
    return json.dumps(progress)

# 定义工具
tools = [search_courses, get_user_progress, ...]

# 创建 Agent
llm = ChatOpenAI(model="gpt-4", temperature=0)
agent = create_openai_tools_agent(llm, tools, prompt)
agent_executor = AgentExecutor(agent=agent, tools=tools)

# 执行
result = agent_executor.invoke({
    "input": "Python 的 GIL 是什么？",
    "chat_history": chat_history
})
```

### 5.3 前端集成

```typescript
// 前端聊天组件
'use client';

import { useState } from 'react';

export function ChatAssistant() {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');

  const sendMessage = async () => {
    const response = await fetch('/api/v1/agent/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        message: input,
        context: {
          current_page: window.location.pathname,
          course_id: getCurrentCourseId()
        }
      })
    });
    
    const data = await response.json();
    setMessages([...messages, 
      { role: 'user', content: input },
      { role: 'assistant', content: data.reply }
    ]);
    setInput('');
  };

  return (
    <div className="chat-container">
      {/* 聊天界面 */}
    </div>
  );
}
```

### 5.4 API 接口

```
POST /api/v1/agent/chat
```

**请求体**：

```json
{
  "message": "Python 的 GIL 是什么？",
  "agent_type": "learning_assistant",
  "context": {
    "current_page": "/course/1",
    "course_id": "uuid",
    "chat_history_id": "uuid"
  }
}
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "reply": "GIL（全局解释器锁）是 Python 中的一个互斥锁，它确保同一时刻只有一个线程执行 Python 字节码...",
    "sources": [
      {
        "course_id": "uuid",
        "title": "Python GIL 深度解析",
        "section": "多线程与并发"
      }
    ],
    "chat_history_id": "uuid"
  }
}
```

---

## 五、简历分析 Agent

### 功能说明

简历分析 Agent 负责分析用户上传的简历，提供评分、亮点和改进建议。

### 分析维度

| 维度 | 权重 | 说明 |
|------|------|------|
| 内容完整性 | 25% | 是否包含必要模块（教育、工作、项目、技能） |
| 排版格式 | 20% | 布局是否清晰、字体是否统一、间距是否合理 |
| 关键词匹配 | 20% | 是否包含目标岗位关键词 |
| 专业性 | 20% | 描述是否专业、数据是否量化 |
| 表达清晰度 | 15% | 语句是否通顺、逻辑是否清晰 |

### API 接口

```
POST /api/v1/resume/analyze
```

**请求**：

```json
{
  "file_url": "https://storage.example.com/resumes/xxx.pdf",
  "file_type": "pdf"
}
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "score": 78,
    "dimensions": [
      { "name": "内容完整性", "score": 85, "comment": "..." },
      { "name": "排版格式", "score": 72, "comment": "..." }
    ],
    "highlights": ["亮点1", "亮点2"],
    "suggestions": ["建议1", "建议2"]
  }
}
```

---

## 六、简历优化 Agent

### 功能说明

简历优化 Agent 根据分析结果和模板结构，生成优化后的简历内容。

### 优化流程

1. 读取用户选择的模板结构
2. 读取简历分析结果
3. 根据模板章节重新组织内容
4. 优化每个章节的描述
5. 生成优化后的简历

### API 接口

```
POST /api/v1/resume/optimize
```

**请求**：

```json
{
  "template_id": "uuid",
  "analysis_id": "uuid",
  "optimize_directions": ["simplify", "highlight_skills", "enhance_data"]
}
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "optimized_content": {
      "summary": "优化后的个人简介...",
      "sections": [
        {
          "title": "教育背景",
          "content": "优化后的内容..."
        }
      ]
    }
  }
}
```

---

## 七、项目实战 Agent

### 功能说明

项目实战 Agent 负责：
1. 根据项目描述生成任务清单
2. 分析用户上传的源码
3. 评估任务完成情况

### 任务生成

**API 接口**：

```
POST /api/v1/projects/generate-tasks
```

**请求**：

```json
{
  "name": "项目名称",
  "description": "项目描述",
  "tech_stack": ["React", "TypeScript"]
}
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "project_id": "uuid",
    "tasks": [
      {
        "id": "uuid",
        "title": "任务标题",
        "description": "任务描述",
        "priority": "high",
        "status": "pending"
      }
    ]
  }
}
```

### 源码分析

**API 接口**：

```
POST /api/v1/projects/analyze
```

**请求**：

```json
{
  "project_id": "uuid",
  "task_id": "uuid"
}
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "files": ["file1.tsx", "file2.ts"],
    "completed_tasks": ["task1", "task2"],
    "pending_tasks": ["task3"],
    "suggestions": ["建议1", "建议2"]
  }
}
```

---

## 附录：Agent 知识库来源

### Python 知识来源

| 来源 | 说明 | 链接 |
|------|------|------|
| Python 官方文档 | 官方教程 | https://docs.python.org/zh-cn/3/ |
| Real Python | 教程文章 | https://realpython.com/ |
| Python Cookbook | 实用食谱 | https://python3-cookbook.readthedocs.io/ |
| PyPI | 包管理 | https://pypi.org/ |

### C++ 知识来源

| 来源 | 说明 | 链接 |
|------|------|------|
| C++ Primer | 经典教材 | - |
| cppreference.com | 标准库文档 | https://en.cppreference.com/ |
| C++ Core Guidelines | 编码规范 | https://isocpp.github.io/CppCoreGuidelines/ |
| LearnCpp.com | 在线教程 | https://www.learncpp.com/ |

### 数据库知识来源

| 来源 | 说明 | 链接 |
|------|------|------|
| PostgreSQL 官方文档 | 官方文档 | https://www.postgresql.org/docs/ |
| SQLZoo | SQL 练习 | https://www.sqlzoo.net/ |
| Redis 官方文档 | Redis 教程 | https://redis.io/docs/ |
| MongoDB 官方文档 | MongoDB 教程 | https://www.mongodb.com/docs/ |

### 算法知识来源

| 来源 | 说明 | 链接 |
|------|------|------|
| LeetCode 题解 | 题目解析 | https://leetcode.cn/ |
| 《算法导论》 | 经典教材 | MIT Press |
| Visualgo | 可视化算法 | https://visualgo.net/ |
| GeeksforGeeks | 算法教程 | https://www.geeksforgeeks.org/ |

### AI Agent 知识来源

| 来源 | 说明 | 链接 |
|------|------|------|
| LangChain 文档 | 框架文档 | https://python.langchain.com/ |
| LangGraph 文档 | 工作流框架 | https://langchain-ai.github.io/langgraph/ |
| OpenAI Cookbook | 最佳实践 | https://cookbook.openai.com/ |
| Hugging Face | 模型库 | https://huggingface.co/ |
