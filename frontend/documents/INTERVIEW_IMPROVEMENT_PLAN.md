# 笔试模拟模块改进计划

> 基于市场调研（LeetCode Interview / Codility / HackerRank / CoderPad / AI Mock Interview 平台）+ CodeForge Academy 现状分析

---

## 一、市场调研：主流笔试/面试模拟平台怎么做

### 1. LeetCode Interview

| 功能 | 说明 |
|------|------|
| 实时协作 | WebRTC + Socket，面试官和候选人同时编码 |
| 题目库 | 内置 LeetCode 全题库 |
| 代码执行 | 多语言在线运行 + 测试用例验证 |
| 评分 | 自动判题（Accepted/WA/TLE）+ 面试官主观评分 |

### 2. Codility / HackerRank

| 功能 | 说明 |
|------|------|
| AI 出题 | 根据岗位/难度/技能自动生成题目 |
| 自动评分 | 代码题用测试用例通过率评分；文字题用 AI 评分 |
| 反馈报告 | 详细的评分报告 + 改进建议 |
| 防作弊 | 代码相似度检测 + 行为监控 |

### 3. AI Mock Interview 平台（2026 趋势）

| 平台 | 核心特性 |
|------|----------|
| **HackerRank AI Mock** | AI 模拟面试官，自适应出题 |
| **InterviewBuddy** | 语音面试模拟 + AI 实时反馈 |
| **mockingly.ai** | 系统设计面试 AI 模拟 + 架构图评估 |
| **InstaMock** | DSA + 系统设计 + LLD，AI 驱动 |
| **Karat AI** | 结构化面试 + AI 评分 + 详细报告 |

### 关键趋势

1. **AI 自适应出题**：根据用户水平动态调整难度
2. **AI 评分 + 详细反馈**：不再只看通过/失败，给出结构化改进建议
3. **代码题 + 系统设计 + 行为面试**：多类型综合
4. **计时 + 倒计时自动提交**：严格模拟真实笔试环境
5. **历史记录 + 进度追踪**：可视化进步曲线

---

## 二、CodeForge 现状问题

### 后端

| # | 问题 | 现状 | 影响 |
|---|------|------|------|
| 1 | **出题是硬编码的** | `generateExam` 返回固定的 4 道题，不管方向/难度 | 无个性化 |
| 2 | **评分只看字数** | `submitExam` 按 `len(answer)` 评分（>80→75分, >250→90分） | 评分不合理 |
| 3 | **代码题不验证** | `runExamQuestion` 只运行不验证测试用例 | 代码对错无法判断 |
| 4 | **无 AI 评分** | 没有用 LLM 对文字题做语义评分 | 反馈无价值 |
| 5 | **无详细反馈** | 只返回分数和一句通用评语 | 无改进指导 |
| 6 | **submitExam 不返回每题反馈** | 前端只 `alert(score)` | 看不到逐题分析 |
| 7 | **无考试详情查看** | `examDetail` 有接口但前端不用 | 无法回顾历史 |

### 前端

| # | 问题 | 现状 | 影响 |
|---|------|------|------|
| 8 | **代码编辑器是 textarea** | 纯文本框 | 体验差 |
| 9 | **运行按钮不工作** | 按钮存在但不调用 API | 假按钮 |
| 10 | **重置按钮不工作** | 同上 | 假按钮 |
| 11 | **提示按钮不工作** | 同上 | 假按钮 |
| 12 | **倒计时不自动运行** | `setTimeLeft` 设置了但没有 timer | 时间不走 |
| 13 | **笔试历史是硬编码的** | 3 条假的记录 | 不真实 |
| 14 | **方向/难度选择不生效** | `<select>` 有 UI 但不传参数 | 选择无效 |
| 15 | **提交结果只弹 alert** | `window.alert(score)` | 体验差 |
| 16 | **无考试回顾** | 不能查看历史笔试详情 | 无法复盘 |

---

## 三、改进方案

### 阶段一：后端核心（AI 出题 + AI 评分 + 代码验证）

#### 3.1 LLM 出题

```go
// 用 LLM 根据方向/难度/题数生成真实题目
func (s *Server) generateExamWithLLM(direction, difficulty string, count int) ([]InterviewQuestion, error) {
    prompt := fmt.Sprintf("生成 %d 道 %s 难度的 %s 笔试题...", count, difficulty, direction)
    // LLM 返回 JSON 格式题目
}
```

#### 3.2 AI 评分

```go
// 文字题用 LLM 评分：0-100 分 + 结构化反馈
func (s *Server) scoreTextAnswer(question, answer string) (score int, feedback string) {
    prompt := "请对以下回答评分（0-100）并给出改进建议..."
}

// 代码题用测试用例验证（复用 ValidateWithCases）
func (s *Server) scoreCodeAnswer(language, code string, cases []TestCase) (score int, feedback string) {
    results := sandbox.ValidateWithCases(language, code, cases)
    score = results.PassedCount * 100 / results.TotalCount
}
```

#### 3.3 详细反馈报告

```json
{
  "score": 82,
  "feedback": [
    { "question_id": "q1", "title": "两数之和", "type": "code", "score": 100, "feedback": "全部测试用例通过" },
    { "question_id": "q2", "title": "设计短链接系统", "type": "text", "score": 75, "feedback": "回答涵盖了核心架构但缺少缓存策略分析..." },
    { "question_id": "q3", "title": "解释事务隔离", "type": "text", "score": 70, "feedback": "概念正确但缺少实际例子..." }
  ]
}
```

### 阶段二：前端体验升级

#### 3.4 Monaco 编辑器（复用 practice 模块的 Monaco）

#### 3.5 倒计时计时器

```tsx
useEffect(() => {
  if (!isStarted) return;
  const timer = setInterval(() => {
    setTimeLeft(prev => {
      if (prev <= 1) { submitAnswer(); return 0; }
      return prev - 1;
    });
  }, 1000);
  return () => clearInterval(timer);
}, [isStarted]);
```

#### 3.6 方向/难度/题数选择生效

- 选择绑定 state，传给 API

#### 3.7 运行/重置/提示按钮对接

- 运行 → `POST /code/run`
- 重置 → 恢复模板代码
- 提示 → `POST /agent/chat`（用 Agent 生成提示）

#### 3.8 提交结果页

- 替换 `alert` 为详细结果卡片
- 逐题显示分数 + 反馈
- 总分 + 百分位

#### 3.9 笔试历史 + 考试详情

- 从 `GET /interview/exams` 获取真实历史
- 点击可查看考试详情

---

## 四、技术选型

| 需求 | 方案 | 理由 |
|------|------|------|
| AI 出题 | LLM（已有 Ark 客户端） | 复用现有基础设施 |
| AI 评分 | LLM（文字题）+ BKT（代码题） | 文字题需语义理解，代码题需测试验证 |
| 代码编辑器 | @monaco-editor/react（已安装） | 复用 practice 模块 |
| 计时器 | React useEffect + setInterval | 轻量 |
| 结果展示 | 卡片式 UI | 无需额外库 |

---

## 五、实施路线图

### 第一期：后端 AI 出题 + 评分

| 序号 | 任务 | 文件 |
|------|------|------|
| 1 | LLM 出题（generateExamWithLLM） | `backend/internal/api/apps.go` |
| 2 | AI 文字题评分（scoreTextAnswer） | `backend/internal/api/apps.go` |
| 3 | 代码题测试验证（复用 ValidateWithCases） | `backend/internal/api/apps.go` |
| 4 | submitExam 返回详细反馈 | `backend/internal/api/apps.go` |
| 5 | runExamQuestion 支持测试用例 | `backend/internal/api/apps.go` |

### 第二期：前端体验升级

| 序号 | 任务 | 文件 |
|------|------|------|
| 6 | Monaco 编辑器替换 textarea | `frontend/src/app/interview/page.tsx` |
| 7 | 倒计时计时器（自动提交） | 同上 |
| 8 | 方向/难度/题数选择生效 | 同上 |
| 9 | 运行/重置/提示按钮对接 | 同上 |
| 10 | 提交结果页（逐题反馈） | 同上 |
| 11 | 笔试历史 + 考试详情 | 同上 |

---

## 六、API 设计

### 新增/修改的接口

```
POST /api/v1/interview/exams/generate  # 修改：用 LLM 出题
POST /api/v1/interview/exams/:id/submit # 修改：返回逐题反馈
POST /api/v1/interview/exams/:id/questions/:qid/run  # 修改：支持测试用例验证
GET  /api/v1/interview/exams            # 已有：返回真实历史
GET  /api/v1/interview/exams/:id       # 已有：考试详情
```
