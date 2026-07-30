# 用户画像改进计划

> 基于市场调研 + CodeForge Academy 项目现状的改进方案

---

## 一、市场调研：主流平台怎么做用户画像

### 1. 在线教育平台的画像维度体系

调研来源：上海开放大学「大数据环境下在线学习者画像的构建」、阿里 DataWorks 用户画像教程、知乎「用户画像标签体系」系列文章。

主流教育平台的画像通常分为 **三层标签体系**：

| 层级 | 内容 | 数据来源 | 更新频率 |
|------|------|----------|----------|
| **事实层（静态）** | 注册信息、年龄段、地区、教育背景 | 用户填写 | 低频（月/季） |
| **行为层（动态）** | 学习时长、课程完成率、做题正确率、活跃时段、连续打卡、Agent 对话次数 | 系统自动采集 | 实时/准实时 |
| **推断层（AI）** | 知识掌握度、学习风格、薄弱知识点、推荐难度、兴趣方向 | 算法/模型推断 | 准实时（每次行为后） |

**关键发现**：大多数平台的画像不只存一份数据，而是 **事实标签 + 实时行为流 + 推断结果** 三层分离存储，推断层用算法动态更新。

### 2. 学习者模型理论

调研来源：Felder-Silverman 学习风格模型、VARK 模型、A Survey of Knowledge Tracing (arXiv 2024)、知乎「知识追踪」系列。

#### 学习风格模型

| 模型 | 维度 | 适用场景 |
|------|------|----------|
| **VARK** | 视觉/听觉/阅读/动觉 | 内容形式推荐 |
| **Felder-Silverman** | 感知-直觉 / 视觉-语言 / 活跃-沉思 / 序列-全局 | 教学路径适配 |
| **Kolb** | 发散/同化/收敛/顺应 | 学习活动类型匹配 |

> **对 CodeForge 的启示**：当前的 `LearningStyle` 只是一个字符串（"实践型"），没有维度化。可以引入 Felder-Silverman 的多维度模型，让 Agent 根据用户偏好调整回答风格。

#### 知识追踪模型

| 模型 | 类型 | 原理 | 复杂度 |
|------|------|------|--------|
| **BKT**（贝叶斯知识追踪） | 概率模型 | 用 4 参数（先验/学习/猜测/失误）建模每个知识点的掌握概率 | 低，可手写 |
| **IRT**（项目反应理论） | 统计模型 | 根据题目难度和学生能力估计正确率 | 低 |
| **DKT**（深度知识追踪） | 深度学习 | 用 RNN/LSTM 追踪知识状态序列 | 高，需训练 |
| **AKT**（注意力知识追踪） | Transformer | 用注意力机制建模知识点间的依赖关系 | 高，需训练 |

> **对 CodeForge 的启示**：我们不需要上 DKT/AKT 这种重量级方案。BKT 足够——每个知识点维护一个 0~1 的掌握概率，用户做对一题就 P(掌握)↑，做错就 P(掌握)↓。Go 里几十行就能实现。

### 3. 主流平台实践

#### Duolingo

- **Birdbrain AI**：自适应难度引擎，根据用户回答动态调整下一题难度
- **Crown 等级系统**：每个技能点有独立的 0-5 Crown 等级，基于答题正确率
- **Half-Recall 算法**：间隔重复算法，预测用户何时会忘记某个知识点
- **画像维度**：学习时段偏好、每日目标、连续天数、每个技能的掌握度

#### Khan Academy

- **Mastery System**：知识点 → 熟悉 → 熟练 → 精通三级体系
- **Skill Profile**：每个技能独立追踪，基于答题正确率和尝试次数
- **Learning Dashboard**：教师端可看到每个学生在每个知识点上的精确状态

#### Coursera

- **Content-based + Collaborative Filtering**：混合推荐
- **用户画像维度**：注册时填写的兴趣领域 + 行为推断的技能水平 + 课程完成度
- **Skill Graph**：将用户技能映射到职业路径

### 4. 数据标准

| 标准 | 说明 | 与 CodeForge 的关系 |
|------|------|---------------------|
| **xAPI (Tin Can API)** | 学习记录标准，`Actor + Verb + Object` 三元组 | 可以参考其事件格式设计行为日志表 |
| **Caliper Analytics** | IMS 全球标准，类似 xAPI | 同上 |
| **LRS (Learning Record Store)** | 学习记录存储 | 类似我们的 `user_activities` 表，但更结构化 |

---

## 二、CodeForge Academy 现状问题

### 后端

| 问题 | 现状 | 影响 |
|------|------|------|
| 画像是只读的 | `agentProfile` 只有 GET，创建后永远默认值 | 用户无法编辑偏好 |
| 学习行为不反馈 | `recordTime()` 不更新 `TotalStudyTime` | 总时长永远为 0 |
| streak 是假数据 | `streak()` 硬编码返回 `15 天` | 打卡不真实 |
| 知识图谱不更新 | Agent 对话后不提取主题写入图谱 | 图谱永远为空 |
| 没有知识追踪 | 无 BKT/IRT，掌握度全靠默认 | 无法精准推荐 |
| 没有标签体系 | `FocusAreas`/`WeakAreas` 是简单字符串数组 | 无法做细粒度分析 |

### 前端

| 问题 | 现状 | 影响 |
|------|------|------|
| 知识图谱是假的 | SVG 节点硬编码 | 与真实数据无关 |
| 学习偏好写死 | `preferences` 数组不来自 API | 无法个性化 |
| 等级写死 | "Lv.5" 不基于计算 | 没有成长感 |
| 无编辑入口 | 用户不能修改偏好 | 无法自服务 |

---

## 三、改进方案

### 阶段一：画像数据闭环（核心）

**目标**：让画像数据从"死的默认值"变成"随行为动态更新的活数据"。

#### 3.1 数据模型扩展

```go
// 新增字段到 UserProfile
type UserProfile struct {
    // ... 原有字段 ...
    PreferredTimeSlot  string   // 偏好学习时段（morning/afternoon/evening/night）
    SessionCount       int      // Agent 对话总次数
    ProblemSolvedCount int      // 解题总数
    ProblemAccuracy    float64  // 平均正确率 0~1
    LastActiveAt       time.Time // 最后活跃时间
}

// 知识追踪节点（新增表）
type KnowledgeState struct {
    ID             string  `gorm:"type:uuid;primaryKey"`
    UserID         string  `gorm:"type:uuid;uniqueIndex:idx_kstate"`
    SkillName      string  `gorm:"size:120;uniqueIndex:idx_kstate"` // 如 "python-decorator"
    Category       string  // python / cpp / database / algorithm / agent
    Mastery        float64 // BKT 后验概率 0~1
    Attempts       int     // 尝试次数
    CorrectCount   int     // 正确次数
    LastPracticedAt time.Time
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

// 行为事件日志（扩展 UserActivity）
// 现有的 UserActivity 已经有 type + text，扩展为 xAPI 风格
type UserActivity struct {
    // ... 原有字段 ...
    Verb     string // viewed / completed / practiced / chatted / submitted
    Object   string // 对象 ID（课程 ID / 题目 ID / 会话 ID）
    Metadata string `gorm:"type:jsonb"` // 结构化附加数据
}
```

#### 3.2 BKT 知识追踪实现

```
参数：
  P(L0) = 0.1    先验掌握概率
  P(T)  = 0.3    每次练习后学会的概率
  P(G)  = 0.25   猜对概率
  P(S)  = 0.1    失误概率

更新规则（做了一题后）：
  做对 → P(L|correct) = P(L) * (1-P(S)) / (P(L)*(1-P(S)) + (1-P(L))*P(G))
  做错 → P(L|wrong)   = P(L) * P(S) / (P(L)*P(S) + (1-P(L))*(1-P(G)))
  练习后 → P(L)' = P(L|result) + (1-P(L|result)) * P(T)
```

#### 3.3 行为反馈闭环

```
用户行为
  │
  ├── recordTime() → 更新 TotalStudyTime + DailyStudyTime + Streak
  ├── 完成课程 → 更新 FocusAreas + KnowledgeState(category)
  ├── 提交题目 → BKT 更新对应 KnowledgeState.Mastery
  ├── Agent 对话 → 提取主题 → 更新 RecentTopics + SessionCount
  └── 更新 LastActiveAt
```

#### 3.4 真实 Streak 计算

从 `DailyStudyTime` 表计算：
- 查最近 N 天每天是否有学习记录
- 连续有记录的天数 = streak
- 今日是否已学习 = today_minutes > 0

### 阶段二：前端可视化升级

#### 3.5 动态知识图谱

- 用 API 返回的 `KnowledgeState` 数据渲染
- 节点大小 = 掌握度（mastery 0~1）
- 节点颜色 = 掌握等级（红<30% / 橙<60% / 蓝<85% / 绿≥85%）
- 节点位置：按 category 分区，用力导向布局或预设位置
- 可点击节点查看详情

#### 3.6 学习偏好可编辑

- 新增"编辑画像"对话框（shadcn Dialog）
- 可编辑：每日目标、偏好难度、偏好时段、学习风格
- 调用 `PUT /agent/profile` 保存

#### 3.7 等级自动计算

```
等级 = f(总学习时长, 完成课程数, 解题数, 连续天数)

Lv 1-3   新手    <10h 且 <2课程
Lv 4-6   进阶    10-50h 或 2-5课程
Lv 7-9   熟练    50-200h 或 5-15课程
Lv 10+   专家    >200h 或 >15课程
```

#### 3.8 学习数据仪表盘

新增卡片：
- 本周学习热力图（7天 × 时段）
- 知识点掌握雷达图（按 5 大方向）
- 最近 7 天学习时长趋势线
- Agent 对话统计

### 阶段三：AI 自动画像

#### 3.9 Agent 对话后自动更新画像

每次 Agent 对话完成后：
1. 用 LLM 从对话内容中提取讨论的知识主题
2. 将主题写入 `RecentTopics`
3. 根据用户表现（追问深度、代码正确性）推断掌握度
4. 更新 `FocusAreas`（如果用户频繁讨论某领域）

#### 3.10 个性化推荐

基于画像数据：
- 根据薄弱知识点推荐课程
- 根据学习风格调整 Agent 回答策略
- 根据掌握度推荐合适难度的题目

---

## 四、技术选型

| 需求 | 方案 | 理由 |
|------|------|------|
| 知识追踪 | BKT（Go 手写） | 轻量、无需训练、可解释 |
| 知识图谱可视化 | SVG + 自定义布局 | 已有 SVG 基础，无需引入 D3 |
| 雷达图 | recharts（已安装） | 前端已有依赖 |
| 热力图 | CSS Grid + Tailwind | 轻量，无需额外库 |
| 行为日志 | 扩展现有 UserActivity | 不引入 Kafka/ES 等重设施 |
| AI 画像推断 | LLM（已有 Ark 客户端） | 复用现有 llm.Client |
| 画像存储 | PostgreSQL JSONB | 已有基础设施，GORM 支持 |

**不引入的技术**：
- ❌ DKT/深度学习知识追踪 → 过重，需 Python + GPU
- ❌ Neo4j 图数据库 → PostgreSQL JSONB 足够
- ❌ Kafka/消息队列 → 行为量不大，同步更新即可
- ❌ xAPI/LRS 完整标准 → 借鉴事件格式，但不引入完整标准栈

---

## 五、实施路线图

### 第一期：数据闭环（核心，1-2 天）

| 序号 | 任务 | 涉及文件 |
|------|------|----------|
| 1 | 扩展数据模型 | `backend/internal/model/models.go` |
| 2 | 新增 BKT 知识追踪 | `backend/internal/api/knowledge_tracing.go`（新） |
| 3 | `recordTime` 反馈 TotalStudyTime + Streak | `backend/internal/api/learning.go` |
| 4 | 真实 Streak 计算 | `backend/internal/api/auth_user.go` |
| 5 | `PUT /agent/profile` 编辑接口 | `backend/internal/api/agent.go` |
| 6 | Agent 对话后更新画像 | `backend/internal/api/agent.go` |
| 7 | 提交题目后 BKT 更新 | `backend/internal/api/course_problem.go` |

### 第二期：前端可视化（1-2 天）

| 序号 | 任务 | 涉及文件 |
|------|------|----------|
| 8 | 动态知识图谱渲染 | `frontend/src/app/agent/profile/page.tsx` |
| 9 | 学习偏好可编辑（Dialog） | 同上 |
| 10 | 等级自动计算 | 同上 |
| 11 | 学习热力图 | 同上 |
| 12 | 知识点雷达图 | 同上 |
| 13 | 学习时长趋势线 | 同上 |

### 第三期：AI 画像（1 天）

| 序号 | 任务 | 涉及文件 |
|------|------|----------|
| 14 | LLM 提取对话主题 | `backend/internal/api/agent.go` |
| 15 | 自动更新 FocusAreas/WeakAreas | 同上 |
| 16 | 基于画像的课程/题目推荐 | `backend/internal/api/course_problem.go` |

---

## 六、API 设计

### 新增/修改的接口

```
PUT  /api/v1/agent/profile           # 编辑用户画像
GET  /api/v1/agent/knowledge-states  # 获取知识追踪状态列表
GET  /api/v1/agent/dashboard         # 仪表盘聚合数据（热力图+雷达+趋势）
POST /api/v1/agent/profile/analyze   # AI 分析对话历史，自动更新画像
```

### 响应格式

```json
// GET /agent/profile
{
  "level": "进阶开发者",
  "computed_level": 5,
  "focus_areas": ["Python", "AI Agent"],
  "weak_areas": ["并发编程"],
  "learning_style": "实践型",
  "preferred_difficulty": "中等",
  "preferred_time_slot": "evening",
  "daily_goal": 60,
  "total_study_time": 256,
  "streak": 15,
  "session_count": 42,
  "problem_solved_count": 128,
  "problem_accuracy": 0.73,
  "last_active_at": "2026-07-31T15:00:00+08:00"
}

// GET /agent/knowledge-states
{
  "items": [
    {
      "skill_name": "python-decorator",
      "category": "python",
      "mastery": 0.78,
      "attempts": 12,
      "correct_count": 9,
      "last_practiced_at": "2026-07-30T20:00:00+08:00"
    }
  ]
}

// GET /agent/dashboard
{
  "heatmap": [{"date":"2026-07-25","minutes":45,"sessions":2}, ...],
  "radar": [{"category":"python","mastery":0.72}, ...],
  "trend": [{"date":"2026-07-25","minutes":45}, ...],
  "stats": {"total_minutes":12340,"total_sessions":42,"total_problems":128,"avg_accuracy":0.73}
}
```

---

## 七、风险与取舍

| 风险 | 取舍 |
|------|------|
| BKT 参数需调优 | 先用经验默认值，后续根据数据微调 |
| LLM 提取主题有延迟 | 放在对话完成后异步执行，不阻塞 SSE |
| 前端知识图谱布局复杂 | 先用分区布局，不引入力导向 |
| 画像更新频率 | 行为触发时同步更新，不做批量/异步队列 |

---

## 八、实施进度

| 序号 | 任务 | 状态 | 文件 |
|------|------|------|------|
| 1 | 扩展数据模型 | ✅ 完成 | `backend/internal/model/models.go` |
| 2 | BKT 知识追踪 | ✅ 完成 | `backend/internal/api/knowledge_tracing.go`（新） |
| 3 | recordTime 反馈 + 真实 Streak | ✅ 完成 | `backend/internal/api/learning.go` `auth_user.go` |
| 4 | PUT /agent/profile 编辑接口 | ✅ 完成 | `backend/internal/api/agent.go` |
| 5 | GET /agent/knowledge-states | ✅ 完成 | `backend/internal/api/agent.go` |
| 6 | GET /agent/dashboard | ✅ 完成 | `backend/internal/api/agent.go` |
| 7 | agentProfile 返回新字段 + computed_level | ✅ 完成 | `backend/internal/api/agent.go` |
| 8 | Agent 对话后 updateProfileFromActivity | ✅ 完成 | `backend/internal/api/agent.go` |
| 9 | 提交题目后 BKT recordKnowledgeState | ✅ 完成 | `backend/internal/api/course_problem.go` |
| 10 | computeLevel 等级自动计算 | ✅ 完成 | `backend/internal/api/knowledge_tracing.go` |
| 11 | computeStreak 真实连续打卡 | ✅ 完成 | `backend/internal/api/knowledge_tracing.go` |
| 12 | 前端页面全面重写 | ✅ 完成 | `frontend/src/app/agent/profile/page.tsx` |
| 13 | 动态知识图谱（真实数据） | ✅ 完成 | 前端同上 |
| 14 | 方向掌握度雷达图 | ✅ 完成 | 前端同上 |
| 15 | 30天学习热力图 | ✅ 完成 | 前端同上 |
| 16 | 7天学习趋势柱状图 | ✅ 完成 | 前端同上 |
| 17 | 统计汇总卡片 | ✅ 完成 | 前端同上 |
| 18 | 学习偏好可编辑 Dialog | ✅ 完成 | 前端同上 |
| 19 | 等级进度条 | ✅ 完成 | 前端同上 |
| 20 | API_SPEC.md 文档更新 | ✅ 完成 | `frontend/documents/API_SPEC.md` |

### 新增文件

- `backend/internal/api/knowledge_tracing.go` — BKT 知识追踪 + 等级计算 + 画像更新 + Streak 计算
- `frontend/documents/USER_PROFILE_PLAN.md` — 本文档

### 修改文件

- `backend/internal/model/models.go` — UserProfile 新字段 + KnowledgeState 新表 + UserActivity 扩展
- `backend/internal/database/database.go` — AutoMigrate 添加 KnowledgeState
- `backend/internal/api/agent.go` — updateProfile + knowledgeStates + agentDashboard + agentProfile 扩展 + 对话后画像更新
- `backend/internal/api/learning.go` — recordTime 后调用 updateProfileFromActivity
- `backend/internal/api/auth_user.go` — streak() 改为真实数据
- `backend/internal/api/course_problem.go` — submitCode 后调用 BKT + 画像更新
- `backend/internal/api/server.go` — 注册新路由
- `frontend/src/app/agent/profile/page.tsx` — 全面重写
- `frontend/documents/API_SPEC.md` — 新增画像接口文档

### 修复记录

- **2026-07-31**: 修复用户画像页面无限转圈问题
  - 根因：`Promise.all` 中任一请求失败会阻塞全部，`catch(() => undefined)` 吞掉错误
  - 修复：三个 API 请求改为独立发起，互不阻塞
  - 新增 8 秒超时（`withTimeout`），避免网络问题导致永久等待
  - loading 状态分阶段：首次加载显示 spinner，失败显示重试按钮
  - 已有数据时后台继续加载，显示加载提示条
