# API 接口文档

> CodeForge Academy 后端 API 接口规范
> 
> 本文档定义了前端与后端之间的 API 契约，供后端开发参考实现。

## 目录

- [1. 概述](#1-概述)
- [2. 认证接口](#2-认证接口)
- [3. 用户接口](#3-用户接口)
- [4. 课程接口](#4-课程接口)
- [5. 算法题目接口](#5-算法题目接口)
- [6. 学习进度接口](#6-学习进度接口)
- [7. 代码执行接口](#7-代码执行接口)
- [8. 评论接口](#8-评论接口)
- [9. 收藏与笔记接口](#9-收藏与笔记接口)

---

## 1. 概述

### 基础信息

| 项目 | 值 |
|------|-----|
| Base URL | `/api/v1` |
| 协议 | HTTPS |
| 数据格式 | JSON |
| 认证方式 | Bearer Token (JWT) |

### 通用响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 2. 认证接口

### 2.1 用户注册

```
POST /api/v1/auth/register
```

**请求体**：

```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

**响应**：

```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": "uuid",
    "username": "string",
    "token": "jwt_token"
  }
}
```

### 2.2 用户登录

```
POST /api/v1/auth/login
```

**请求体**：

```json
{
  "email": "string",
  "password": "string"
}
```

**响应**：

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user_id": "uuid",
    "username": "string",
    "token": "jwt_token",
    "expires_in": 86400
  }
}
```

---

## 3. 用户接口

### 3.1 获取用户信息

```
GET /api/v1/users/me
```

**Headers**: `Authorization: Bearer <token>`

**响应**：

```json
{
  "code": 200,
  "data": {
    "user_id": "uuid",
    "username": "小初",
    "avatar": "初",
    "level": 5,
    "level_title": "进阶学习者",
    "join_date": "2024-03-15",
    "learning_days": 128,
    "stats": {
      "total_hours": 256,
      "completed_courses": 23,
      "solved_problems": 81,
      "current_streak": 15
    }
  }
}
```

### 3.2 更新用户信息

```
PUT /api/v1/users/me
```

**请求体**：

```json
{
  "username": "string",
  "avatar": "string",
  "bio": "string"
}
```

---

## 4. 课程接口

### 4.1 获取课程列表

```
GET /api/v1/courses
```

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| category | string | 否 | 分类：python / cpp / database / algorithm / agent |
| difficulty | string | 否 | 难度：beginner / intermediate / advanced |
| status | string | 否 | 状态：completed / in_progress / not_started |
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 20 |

**响应**：

```json
{
  "code": 200,
  "data": {
    "total": 25,
    "page": 1,
    "items": [
      {
        "course_id": "uuid",
        "title": "Python 装饰器与元编程详解",
        "category": "python",
        "difficulty": "intermediate",
        "cover_image": "https://...",
        "summary": "深入理解 Rust 的所有权、借用和生命周期机制...",
        "author": "陈教授",
        "publish_date": "2024-01-15",
        "read_time": "15分钟",
        "views": 2847,
        "tags": ["Rust", "所有权", "内存管理"],
        "status": "in_progress",
        "progress": 60
      }
    ]
  }
}
```

### 4.2 获取课程详情

```
GET /api/v1/courses/:id
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "course_id": "uuid",
    "title": "Python 装饰器与元编程详解",
    "category": "python",
    "difficulty": "intermediate",
    "cover_image": "https://...",
    "author": "陈教授",
    "publish_date": "2024-01-15",
    "read_time": "15分钟",
    "views": 2847,
    "tags": ["Python", "装饰器", "元编程"],
    "sections": [
      {
        "section_id": "uuid",
        "title": "装饰器基础",
        "content": "理解 Python 装饰器的原理...",
        "code_examples": [
          {
            "language": "python",
            "code": "def decorator(func):\n    def wrapper(*args):\n        return func(*args)\n    return wrapper"
          }
        ]
      }
    ],
    "prev_course": {
      "id": "uuid",
      "title": "变量与类型"
    },
    "next_course": {
      "id": "uuid",
      "title": "枚举与模式匹配"
    }
  }
}
```

---

## 5. 算法题目接口

### 5.1 获取题目列表

```
GET /api/v1/problems
```

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| category | string | 否 | 分类：array / linked_list / tree / dp / ... |
| difficulty | string | 否 | 难度：easy / medium / hard |
| status | string | 否 | 状态：solved / attempted / not_started |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应**：

```json
{
  "code": 200,
  "data": {
    "total": 360,
    "solved": 81,
    "items": [
      {
        "problem_id": 1,
        "title": "两数之和",
        "category": "array",
        "difficulty": "easy",
        "pass_rate": 48.5,
        "submissions": 3847291,
        "status": "solved",
        "tags": ["数组", "哈希表"]
      }
    ]
  }
}
```

### 5.2 获取题目详情

```
GET /api/v1/problems/:id
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "problem_id": 42,
    "title": "两数之和",
    "difficulty": "easy",
    "description": "给定一个整数数组 nums 和一个整数目标值 target...",
    "examples": [
      {
        "input": "nums = [2,7,11,15], target = 9",
        "output": "[0,1]",
        "explanation": "因为 nums[0] + nums[1] == 9"
      }
    ],
    "constraints": [
      "2 <= nums.length <= 10^4",
      "-10^9 <= nums[i] <= 10^9"
    ],
    "tags": ["数组", "哈希表"],
    "related_problems": [15, 167]
  }
}
```

### 5.3 获取每日一题

```
GET /api/v1/problems/daily
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "problem_id": 200,
    "title": "岛屿数量",
    "difficulty": "medium",
    "date": "2024-01-20"
  }
}
```

---

## 6. 学习进度接口

### 6.1 获取学习进度

```
GET /api/v1/progress
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "python": {
      "progress": 45,
      "completed_chapters": 4,
      "total_chapters": 8
    },
    "cpp": {
      "progress": 32,
      "completed_chapters": 3,
      "total_chapters": 9
    },
    "database": {
      "progress": 28,
      "completed_chapters": 3,
      "total_chapters": 10
    },
    "algorithm": {
      "progress": 32,
      "solved_problems": 81,
      "total_problems": 360,
      "by_difficulty": {
        "easy": { "solved": 45, "total": 120 },
        "medium": { "solved": 28, "total": 180 },
        "hard": { "solved": 8, "total": 60 }
      }
    },
    "agent": {
      "progress": 18,
      "completed_modules": 3,
      "total_modules": 12
    }
  }
}
```

### 6.2 更新课程进度

```
PUT /api/v1/progress/courses/:id
```

**请求体**：

```json
{
  "progress": 60,
  "last_section_id": "uuid"
}
```

### 6.3 记录学习时长

```
POST /api/v1/progress/time
```

**请求体**：

```json
{
  "duration_minutes": 45,
  "course_id": "uuid"
}
```

---

## 7. 代码执行接口

### 7.1 提交代码执行

```
POST /api/v1/code/run
```

**请求体**：

```json
{
  "problem_id": 42,
  "language": "python",
  "code": "def two_sum(nums, target):\n    seen = {}\n    for i, n in enumerate(nums):\n        if target - n in seen:\n            return [seen[target-n], i]\n    return []"
}
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "status": "accepted",
    "test_results": [
      {
        "test_case": 1,
        "input": "[2,7,11,15], 9",
        "expected": "[0,1]",
        "actual": "[0,1]",
        "passed": true
      }
    ],
    "execution_time_ms": 52,
    "memory_kb": 16800,
    "stdout": ""
  }
}
```

### 7.2 提交答案

```
POST /api/v1/code/submit
```

**请求体**：同 7.1

**响应**：

```json
{
  "code": 200,
  "data": {
    "status": "accepted",
    "passed_cases": 45,
    "total_cases": 45,
    "execution_time_ms": 52,
    "memory_kb": 16800,
    "ranking_percentile": 78.5
  }
}
```

---

## 8. 评论接口

### 8.1 获取评论列表

```
GET /api/v1/courses/:id/comments
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "total": 3,
    "items": [
      {
        "comment_id": "uuid",
        "user": {
          "user_id": "uuid",
          "username": "张明远",
          "avatar": "张"
        },
        "content": "这篇文章讲解得非常清晰...",
        "created_at": "2024-01-18T10:30:00Z",
        "likes": 12
      }
    ]
  }
}
```

### 8.2 发表评论

```
POST /api/v1/courses/:id/comments
```

**请求体**：

```json
{
  "content": "评论内容..."
}
```

---

## 9. 收藏与笔记接口

### 9.1 收藏课程

```
POST /api/v1/favorites
```

**请求体**：

```json
{
  "course_id": "uuid"
}
```

### 9.2 取消收藏

```
DELETE /api/v1/favorites/:course_id
```

### 9.3 获取收藏列表

```
GET /api/v1/favorites
```

### 9.4 创建笔记

```
POST /api/v1/notes
```

**请求体**：

```json
{
  "course_id": "uuid",
  "title": "所有权规则总结",
  "content": "笔记内容..."
}
```

### 9.5 获取笔记列表

```
GET /api/v1/notes
```

---

## 10. 简历相关接口

### 10.1 获取简历模板列表

```
GET /api/v1/resume/templates
```

**查询参数**：
- `category`: 模板分类（可选）

**响应示例**：

```json
{
  "templates": [
    {
      "id": "template-001",
      "name": "技术岗位模板",
      "category": "tech",
      "description": "适合软件工程师、数据科学家等技术岗位",
      "sections": ["个人信息", "教育背景", "技能清单", "项目经历", "工作经历"],
      "style": "简洁专业，突出技术栈和项目经验",
      "preview_url": "/templates/tech-preview.png"
    }
  ]
}
```

### 10.2 上传简历

```
POST /api/v1/resume/upload
Content-Type: multipart/form-data
```

**请求参数**：
- `file`: 简历文件（PDF/DOC/DOCX）

**响应示例**：

```json
{
  "success": true,
  "file_id": "file-uuid",
  "filename": "resume.pdf",
  "file_size": 102400
}
```

### 10.3 AI 分析简历

```
POST /api/v1/resume/analyze
```

**请求体**：

```json
{
  "file_id": "file-uuid"
}
```

**响应示例**：

```json
{
  "success": true,
  "analysis": {
    "overall_score": 85,
    "dimensions": {
      "content_completeness": 90,
      "format_layout": 80,
      "keyword_match": 85,
      "professionalism": 88,
      "expression_clarity": 82
    },
    "highlights": ["技术栈描述清晰", "项目经验量化"],
    "suggestions": ["增加更多数据指标", "优化技能排序"]
  }
}
```

### 10.4 AI 优化简历

```
POST /api/v1/resume/optimize
```

**请求体**：

```json
{
  "file_id": "file-uuid",
  "template_id": "template-001",
  "optimization_directions": ["精简内容", "突出技能"]
}
```

**响应示例**：

```json
{
  "success": true,
  "optimized_content": {
    "sections": [
      {
        "title": "个人信息",
        "content": "优化后的内容..."
      }
    ],
    "template_used": "技术岗位模板"
  }
}
```

### 10.5 导出简历

```
POST /api/v1/resume/export
```

**请求体**：

```json
{
  "format": "pdf",
  "content": { ... }
}
```

**响应**：文件下载流

---

## 11. 项目实战接口

### 11.1 创建项目

```
POST /api/v1/projects
```

**请求体**：

```json
{
  "name": "我的项目",
  "description": "项目描述",
  "tech_stack": ["Python", "FastAPI", "PostgreSQL"]
}
```

**响应示例**：

```json
{
  "success": true,
  "project": {
    "id": "project-uuid",
    "name": "我的项目",
    "tasks": [
      { "id": "task-001", "title": "需求分析", "description": "...", "status": "pending" }
    ]
  }
}
```

### 11.2 上传项目源码

```
POST /api/v1/projects/upload
Content-Type: multipart/form-data
```

**请求参数**：
- `project_id`: 项目 ID
- `files`: 源码文件（支持多个文件）

**响应示例**：

```json
{
  "success": true,
  "task_id": "task-uuid",
  "files": [
    { "name": "main.py", "size": 2048 }
  ]
}
```

### 11.3 AI 分析源码

```
POST /api/v1/projects/analyze
```

**请求体**：

```json
{
  "project_id": "project-uuid",
  "task_id": "task-uuid"
}
```

**响应示例**：

```json
{
  "success": true,
  "analysis": {
    "files": [
      { "name": "main.py", "size": 2048 }
    ],
    "completed_tasks": ["task-001", "task-002"],
    "pending_tasks": ["task-003"],
    "suggestions": [
      "建议添加单元测试",
      "代码结构可以进一步优化"
    ]
  }
}
```

---

## 附录：学习资源来源

### Python 学习资源

| 资源 | 来源 | 链接 |
|------|------|------|
| Python 官方文档 | 官方 | https://docs.python.org/zh-cn/3/ |
| Real Python | 教程网站 | https://realpython.com/ |
| Python Cookbook | 经典书籍 | O'Reilly Media |
| PyPI | 包管理 | https://pypi.org/ |

### C++ 学习资源

| 资源 | 来源 | 链接 |
|------|------|------|
| C++ Primer | 经典教材 | Addison-Wesley |
| cppreference.com | 在线文档 | https://en.cppreference.com/ |
| C++ Core Guidelines | 官方指南 | https://isocpp.github.io/CppCoreGuidelines/ |
| LearnCpp.com | 教程网站 | https://www.learncpp.com/ |

### 数据库学习资源

| 资源 | 来源 | 链接 |
|------|------|------|
| PostgreSQL 官方文档 | 官方 | https://www.postgresql.org/docs/ |
| SQLZoo | 在线练习 | https://sqlzoo.net/ |
| Redis 官方文档 | 官方 | https://redis.io/docs/ |
| MongoDB University | 官方课程 | https://learn.mongodb.com/ |
| 《数据库系统概念》 | 经典教材 | 机械工业出版社 |

### 算法学习资源

| 资源 | 来源 | 链接 |
|------|------|------|
| LeetCode | 在线题库 | https://leetcode.cn/ |
| 力扣 | 中文站 | https://leetcode.cn/ |
| 《算法导论》 | 经典教材 | MIT Press |
| Visualgo | 可视化 | https://visualgo.net/ |

### AI Agent 学习资源

| 资源 | 来源 | 链接 |
|------|------|------|
| LangChain 文档 | 官方 | https://python.langchain.com/ |
| LangGraph | 官方 | https://langchain-ai.github.io/langgraph/ |
| OpenAI API | 官方 | https://platform.openai.com/docs |
| Hugging Face | 模型库 | https://huggingface.co/ |

## 11. AI Agent 接口

### 11.1 智能对话（SSE 流式）

`POST /api/v1/agent/chat`

请求体：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| message | string | 是 | 用户消息（仅附件时可省略） |
| session_id | string | 否 | 会话 ID，留空时后端自动生成；同一会话会被用于加载对话记忆 |
| agent_type | string | 否 | 指定 Agent，留空时智能路由 |
| collaboration_mode | string | 否 | 协作模式，默认 dynamic |
| attachments | string[] | 否 | 已上传的附件 URL |
| context | object | 否 | 页面上下文 |

响应：`text/event-stream`，事件包括 `agent_route`、`workflow_start`、`workflow_step`、`tool_call`、`tool_result`、`token`、`done`。每个 turn 的用户消息与助手回答都会持久化到 `session_messages`，作为对话记忆的唯一来源。

### 11.2 会话列表（对话历史）

`GET /api/v1/agent/sessions`

返回当前用户的全部会话，按最近活跃倒序。无需单独维护会话表——直接从 `session_messages` 按 `session_id` 聚合：

| 字段 | 类型 | 说明 |
|------|------|------|
| session_id | string | 会话 ID |
| title | string | 会话标题（取第一条用户消息，截断 60 字） |
| agent | string | 最近一次使用的 Agent |
| message_count | int | 消息总数 |
| created_at | string | 会话创建时间 |
| updated_at | string | 会话最近更新时间 |

### 11.3 会话消息历史

`GET /api/v1/agent/history?session_id=xxx`

返回指定会话的全部消息（按时间正序，上限 200 条）。

### 11.4 删除会话

`DELETE /api/v1/agent/sessions/:id`

删除指定会话下的全部消息。返回 `{ deleted: true, session_id, deleted_count }`。

### 11.5 对话记忆说明

后端 `loadConversationMemory(uid, sessionID)` 从 `session_messages` 读取最近 12 条作为工作记忆，`buildConversationPrompt` 把记忆拼入模型提示，`contextualToolMessage` 在追问时把最近 4 轮记忆注入工具调用，确保指代（“这个/它/上面”）能被正确解析。记忆完全基于持久化数据，重启或刷新后仍可恢复。
