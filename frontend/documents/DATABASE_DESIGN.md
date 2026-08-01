# 数据库设计文档

> CodeForge Academy 数据库选型与表结构设计
>
> 本文档定义了后端数据库的选型理由和表结构设计。

## 目录

- [1. 数据库选型](#1-数据库选型)
- [2. 表结构设计](#2-表结构设计)
- [3. 索引设计](#3-索引设计)
- [4. 数据迁移策略](#4-数据迁移策略)

---

## 1. 数据库选型

### 推荐方案：PostgreSQL

| 维度 | 选择 | 理由 |
|------|------|------|
| 主数据库 | PostgreSQL | 开源、功能丰富、支持 JSONB、全文搜索、扩展性强 |
| 缓存 | Redis | 高性能缓存、会话管理、排行榜 |
| 搜索引擎 | PostgreSQL FTS / Meilisearch | 课程/题目全文搜索 |
| 向量数据库 | pgvector (PostgreSQL 扩展) | 存储课程/题目的向量嵌入，支持语义搜索 |

### 选型理由

#### 为什么选择 PostgreSQL？

1. **JSONB 支持**：课程章节内容、代码示例等半结构化数据可用 JSONB 存储，灵活且支持索引
2. **全文搜索**：内置全文搜索功能，满足课程/题目搜索需求
3. **pgvector 扩展**：支持向量存储和相似度搜索，为 AI 推荐功能提供基础
4. **事务支持**：ACID 事务保证数据一致性（如学习进度更新、代码提交）
5. **社区生态**：丰富的 ORM 支持（Prisma、TypeORM、SQLAlchemy）
6. **可扩展性**：支持分区表、读写分离、连接池

#### 备选方案对比

| 数据库 | 优点 | 缺点 | 适用场景 |
|--------|------|------|----------|
| MySQL | 成熟稳定、生态好 | JSON 支持弱、无向量扩展 | 简单 CRUD 应用 |
| MongoDB | 文档模型灵活 | 事务支持弱、无 SQL | 内容管理系统 |
| SQLite | 轻量、零配置 | 并发能力差 | 本地开发/小型应用 |

---

## 2. 表结构设计

### 2.1 用户表 (users)

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    avatar VARCHAR(10),
    bio TEXT,
    level INT DEFAULT 1,
    level_title VARCHAR(50) DEFAULT '初学者',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);

-- 索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
```

### 2.2 课程表 (courses)

```sql
CREATE TABLE courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    category VARCHAR(50) NOT NULL,  -- python / cpp / database / algorithm / agent
    difficulty VARCHAR(20) NOT NULL,  -- beginner / intermediate / advanced
    cover_image TEXT,
    summary TEXT,
    content JSONB,  -- 章节内容，结构：[{section_id, title, content, code_examples}]
    author VARCHAR(100),
    publish_date DATE,
    read_time VARCHAR(20),
    views INT DEFAULT 0,
    tags TEXT[],  -- 标签数组
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_courses_category ON courses(category);
CREATE INDEX idx_courses_difficulty ON courses(difficulty);
CREATE INDEX idx_courses_tags ON courses USING GIN(tags);
CREATE INDEX idx_courses_publish_date ON courses(publish_date DESC);

-- 全文搜索索引
CREATE INDEX idx_courses_search ON courses USING GIN(to_tsvector('chinese', title || ' ' || summary));
```

### 2.3 算法题目表 (problems)

```sql
CREATE TABLE problems (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    category VARCHAR(50) NOT NULL,  -- array / linked_list / tree / dp / ...
    difficulty VARCHAR(20) NOT NULL,  -- easy / medium / hard
    description TEXT NOT NULL,
    examples JSONB,  -- [{input, output, explanation}]
    constraints TEXT[],
    tags TEXT[],
    pass_rate DECIMAL(5,2) DEFAULT 0,
    total_submissions INT DEFAULT 0,
    accepted_submissions INT DEFAULT 0,
    test_cases JSONB,  -- 测试用例
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_problems_category ON problems(category);
CREATE INDEX idx_problems_difficulty ON problems(difficulty);
CREATE INDEX idx_problems_tags ON problems USING GIN(tags);
```

### 2.4 学习进度表 (learning_progress)

```sql
CREATE TABLE learning_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id UUID REFERENCES courses(id) ON DELETE CASCADE,
    problem_id INT REFERENCES problems(id) ON DELETE CASCADE,
    progress INT DEFAULT 0,  -- 0-100
    status VARCHAR(20) DEFAULT 'not_started',  -- not_started / in_progress / completed / solved
    last_section_id VARCHAR(100),
    last_active_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_id, course_id),
    UNIQUE(user_id, problem_id)
);

-- 索引
CREATE INDEX idx_progress_user ON learning_progress(user_id);
CREATE INDEX idx_progress_status ON learning_progress(status);
```

### 2.5 学习记录表 (learning_records)

```sql
CREATE TABLE learning_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id UUID REFERENCES courses(id) ON DELETE SET NULL,
    activity_type VARCHAR(50) NOT NULL,  -- complete_course / solve_problem / start_course / ...
    description TEXT,
    duration_minutes INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_records_user ON learning_records(user_id);
CREATE INDEX idx_records_created ON learning_records(created_at DESC);
```

### 2.6 代码提交表 (submissions)

```sql
CREATE TABLE submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id INT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    language VARCHAR(20) NOT NULL,  -- python / cpp / javascript
    code TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,  -- accepted / wrong_answer / time_limit / ...
    execution_time_ms INT,
    memory_kb INT,
    test_results JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_submissions_user ON submissions(user_id);
CREATE INDEX idx_submissions_problem ON submissions(problem_id);
CREATE INDEX idx_submissions_status ON submissions(status);
```

### 2.7 评论表 (comments)

```sql
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    likes INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_comments_course ON comments(course_id);
CREATE INDEX idx_comments_created ON comments(created_at DESC);
```

### 2.8 收藏表 (favorites)

```sql
CREATE TABLE favorites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_id, course_id)
);

-- 索引
CREATE INDEX idx_favorites_user ON favorites(user_id);
```

### 2.9 笔记表 (notes)

```sql
CREATE TABLE notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id UUID REFERENCES courses(id) ON DELETE SET NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_notes_user ON notes(user_id);
CREATE INDEX idx_notes_course ON notes(course_id);
```

### 2.10 成就表 (achievements)

```sql
CREATE TABLE achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(50),
    condition_type VARCHAR(50),  -- courses_completed / problems_solved / streak_days / ...
    condition_value INT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE user_achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    unlocked_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_id, achievement_id)
);
```

### 2.11 学习路径表 (learning_paths)

```sql
CREATE TABLE learning_paths (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,  -- python / cpp / database / algorithm / agent
    description TEXT,
    total_courses INT DEFAULT 0,
    difficulty_range VARCHAR(50),  -- "入门 → 高级"
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE learning_path_stages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    path_id UUID NOT NULL REFERENCES learning_paths(id) ON DELETE CASCADE,
    stage_number INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    goal TEXT,
    course_ids UUID[],
    total_hours INT DEFAULT 0,
    prerequisite VARCHAR(100)
);
```

### 10. resume_templates - 简历模板表

```sql
CREATE TABLE resume_templates (
    id VARCHAR(60) PRIMARY KEY,           -- 内置模板为固定 id(如 tech-standard)；自定义模板为 uuid
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL, -- tech/pm/designer/marketing/general/academic/freshgraduate/custom
    description TEXT,
    preview VARCHAR(20),                  -- 模板预览 emoji
    sections JSONB NOT NULL,              -- 章节名数组(字符串)；自定义模板注册时写入确认后的章节名
    style TEXT,                           -- 风格描述
    object_key VARCHAR(500),              -- MinIO 中模板原文件 key(自定义模板)
    registered_path VARCHAR(500),         -- 注册后(已注入占位符)模板本地路径 backend/data/templates/{id}.docx
    status VARCHAR(20) DEFAULT '',        -- draft(已上传未确认)/ ready(确认注册完成)；内置模板为空
    owner_id VARCHAR(60) DEFAULT '',      -- 上传者；空串 = 系统内置模板
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_resume_templates_category ON resume_templates(category);
CREATE INDEX idx_resume_templates_owner ON resume_templates(owner_id);
```

> 变更说明(2026-08-01)：新增 `object_key` / `registered_path` / `status` / `owner_id` 四列，支持用户上传自己的 DOCX 模板并注册。GORM AutoMigrate 自动迁移。
> 自定义模板注册流程：上传(`status=draft`) → 用户确认章节(`status=ready`) → `registered_path` 指向已注入 `{{ sections[i]['items'][j] }}` 占位符的 docx 文件。

### 11. resumes - 简历表

```sql
CREATE TABLE resumes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    template_id UUID REFERENCES resume_templates(id),
    title VARCHAR(200) NOT NULL,
    original_file_url VARCHAR(500), -- 原始简历文件
    analysis_result JSONB, -- AI 分析结果
    optimized_content JSONB, -- 优化后的内容
    export_format VARCHAR(20), -- pdf/docx/html
    status VARCHAR(20) DEFAULT 'draft', -- draft/analyzed/optimized/exported
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_resumes_user ON resumes(user_id);
CREATE INDEX idx_resumes_status ON resumes(status);
```

### 12. projects - 项目实战表

```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    tech_stack TEXT,
    tasks JSONB NOT NULL, -- [{id, title, description, status}]
    status VARCHAR(20) DEFAULT 'active', -- active/completed/archived
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_projects_user ON projects(user_id);
CREATE INDEX idx_projects_status ON projects(status);
```

### 13. project_submissions - 项目提交表

```sql
CREATE TABLE project_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    files JSONB NOT NULL, -- [{name, size, url}]
    task_id VARCHAR(50), -- 上传任务 ID
    analysis_result JSONB, -- AI 分析结果
    completed_task_ids UUID[], -- 已完成的任务 ID
    pending_task_ids UUID[], -- 待完成的任务 ID
    suggestions TEXT[], -- AI 建议
    status VARCHAR(20) DEFAULT 'uploaded', -- uploaded/analyzed
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_project_submissions_project ON project_submissions(project_id);
CREATE INDEX idx_project_submissions_user ON project_submissions(user_id);
```

---

## 3. 索引设计

### 核心查询场景与索引

| 查询场景 | 涉及表 | 推荐索引 |
|----------|--------|----------|
| 按分类浏览课程 | courses | `idx_courses_category` |
| 按难度筛选课程 | courses | `idx_courses_difficulty` |
| 按标签搜索课程 | courses | `idx_courses_tags` (GIN) |
| 全文搜索课程 | courses | `idx_courses_search` (GIN) |
| 用户学习进度 | learning_progress | `idx_progress_user` |
| 题目列表按分类 | problems | `idx_problems_category` |
| 代码提交记录 | submissions | `idx_submissions_user`, `idx_submissions_problem` |

---

## 4. 数据迁移策略

### 推荐工具：Prisma Migrate

```prisma
// prisma/schema.prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id            String   @id @default(uuid())
  username      String   @unique
  email         String   @unique
  passwordHash  String   @map("password_hash")
  avatar        String?
  level         Int      @default(1)
  createdAt     DateTime @default(now()) @map("created_at")
  
  @@map("users")
}

// ... 其他模型
```

### 迁移命令

```bash
# 生成迁移
npx prisma migrate dev --name init

# 应用迁移
npx prisma migrate deploy

# 重置数据库
npx prisma migrate reset
```

---

## 附录：Redis 缓存设计

### 缓存 Key 规范

| Key | 类型 | TTL | 说明 |
|-----|------|-----|------|
| `session:{user_id}` | String | 24h | 用户会话 |
| `course:detail:{id}` | String | 1h | 课程详情缓存 |
| `problem:detail:{id}` | String | 1h | 题目详情缓存 |
| `user:progress:{user_id}` | Hash | 30min | 用户学习进度 |
| `daily:problem` | String | 24h | 每日一题 |
| `leaderboard:daily` | ZSet | 24h | 每日排行榜 |
