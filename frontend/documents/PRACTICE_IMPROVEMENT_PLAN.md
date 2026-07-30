# 在线练习模块改进计划

> 基于市场调研（LeetCode / HackerRank / Codewars / Exercism）+ CodeForge Academy 现状分析

---

## 一、市场调研：主流代码练习平台怎么做

### 1. LeetCode

| 功能 | 说明 |
|------|------|
| 代码编辑器 | Monaco Editor（VS Code 同款），语法高亮 + 自动补全 + 多主题 |
| 测试用例 | 公开样例 + 隐藏测试用例，逐条展示通过/失败 |
| 提交结果 | Accepted / Wrong Answer / TLE / MLE / Runtime Error / Compile Error |
| 执行详情 | 执行时间 + 内存消耗 + 击败百分比 |
| 提交历史 | 每题完整提交记录，可查看历史代码 |
| 题目列表 | 难度筛选 + 标签筛选 + 状态筛选（已通过/尝试过/未做）+ 搜索 |
| 收藏 | 收藏题目 + 自定义分类 |
| 排行榜 | 全局排名 + 周赛/双周赛排名 |
| 讨论区 | 每题独立讨论 |
| 提示系统 | 逐步提示 |

### 2. HackerRank

| 功能 | 说明 |
|------|------|
| 测试用例 | Sample（可见）+ Hidden（隐藏），逐条执行结果 |
| 代码模板 | 多语言模板，支持语言切换时保留代码 |
| 执行反馈 | 编译错误高亮行号 + 运行时错误堆栈 |
| 评分 | 基于测试用例通过率 + 执行效率 |

### 3. 代码编辑器技术对比

| 编辑器 | 包大小 | 语法高亮 | 自动补全 | 移动端 | 推荐 |
|--------|--------|----------|----------|--------|------|
| **Monaco** | ~2MB | ✅ 完整 | ✅ IntelliSense | ❌ 不支持 | 桌面端首选（VS Code 同款） |
| **CodeMirror 6** | ~200KB | ✅ | ✅ 基础 | ✅ 支持 | 移动端/轻量场景 |
| **Ace** | ~500KB | ✅ | ✅ | 部分支持 | 老牌稳定 |
| **纯 textarea** | 0 | ❌ | ❌ | ✅ | **当前方案** |

> **当前 CodeForge 用的是纯 `<textarea>`** — 无语法高亮、无自动补全、无行号、无错误标记。

---

## 二、CodeForge 现状问题

### 后端

| # | 问题 | 现状 | 影响 |
|---|------|------|------|
| 1 | **测试用例不验证输出** | `ValidateSolution` 只检查代码能否运行，不比对 stdout 和 expected | 用户代码错误也能 "accepted" |
| 2 | **无逐条测试结果** | `submitCode` 返回 `passed: 3, total: 3`，但不返回每条用例的输入/期望/实际/通过 | 前端测试结果全是假的 |
| 3 | **提交结果不更新题目 passRate/submissions** | 提交后不更新题目统计 | 通过率永远不变 |
| 4 | **无提交历史查询接口** | 有 Submission 表但没有 API 查询 | 用户看不到历史提交 |
| 5 | **runCode 不带 problem_id** | 运行代码时不验证题目，只是纯执行 | 无法做题目相关的预测试 |
| 6 | **沙箱无 stdin 注入** | 测试用例的 input 没有注入到代码的 stdin | 无法做输入输出匹配验证 |
| 7 | **无排行榜** | 有 Submission 数据但没有排名接口 | 缺少竞争激励 |

### 前端

| # | 问题 | 现状 | 影响 |
|---|------|------|------|
| 8 | **代码编辑器是 textarea** | 纯文本框，无语法高亮/补全/行号 | 体验远不如 LeetCode |
| 9 | **测试结果是硬编码的** | "测试结果" tab 里的 3 条用例全是写死的 | 与实际运行无关 |
| 10 | **只支持 3 种语言** | rust/python/javascript，后端支持 cpp 但前端没加 | C++ 用户无法用 |
| 11 | **无题目列表/导航** | 只能做 problem_id=1，不能切题 | 无法浏览题库 |
| 12 | **无提交历史** | 看不到过去的提交 | 无法回溯 |
| 13 | **无收藏功能** | 有 `isFavorite` 状态但不调用 API | 收藏不生效 |
| 14 | **代码不持久化** | 刷新页面代码丢失 | 用户需要重写 |

---

## 三、改进方案

### 阶段一：后端测试验证闭环（核心）

#### 3.1 沙箱支持 stdin + 真实测试用例验证

```go
// RunWithInput 执行代码并注入 stdin，返回 stdout
func RunWithInput(language, code, stdin string) (Result, error)

// ValidateWithCases 用多个测试用例验证代码
func ValidateWithCases(language, code string, cases []TestCase) (CaseResults, error)
type CaseResult struct {
    Input     string
    Expected  string
    Actual    string
    Passed    bool
    Error     string
    TimeMS    int64
}
```

#### 3.2 submitCode 返回逐条测试结果

```json
{
  "status": "accepted",
  "passed_cases": 3,
  "total_cases": 3,
  "case_results": [
    { "input": "nums = [2,7,11,15], target = 9", "expected": "[0,1]", "actual": "[0,1]", "passed": true, "time_ms": 12 },
    { "input": "nums = [3,2,4], target = 6", "expected": "[1,2]", "actual": "[1,2]", "passed": true, "time_ms": 8 },
    { "input": "nums = [3,3], target = 6", "expected": "[0,1]", "actual": "[0,1]", "passed": true, "time_ms": 9 }
  ],
  "execution_time_ms": 29,
  "memory_kb": 2048,
  "ranking_percentile": 78
}
```

#### 3.3 提交后更新题目统计

```go
// submitCode 后：
// 1. 更新 problem.pass_rate（通过人数/总提交人数）
// 2. 更新 problem.submissions（总提交数 +1）
// 3. 如果首次通过，更新 problem.status = "solved"
```

#### 3.4 提交历史接口

```
GET /api/v1/problems/:id/submissions  — 某题的提交历史
GET /api/v1/submissions               — 全部提交历史（分页）
```

#### 3.5 排行榜接口

```
GET /api/v1/leaderboard  — 按解题数排名
```

### 阶段二：前端编辑器升级

#### 3.6 引入 Monaco Editor

- 使用 `@monaco-editor/react`（无需 webpack 配置）
- 语法高亮 + 自动补全 + 行号 + 主题
- 支持 python / javascript / cpp / rust 四种语言

#### 3.7 真实测试结果展示

- "测试结果" tab 用 `case_results` 渲染
- 每条用例：✓/✗ + 输入 + 期望输出 + 实际输出 + 耗时
- 失败用例高亮红色

#### 3.8 题目列表 + 切题导航

- 左侧题目列表（难度/标签/状态筛选）
- 上一题/下一题切换
- 题目搜索

#### 3.9 代码本地持久化

- 代码存 localStorage（按 problem_id + language 为 key）
- 刷新页面自动恢复

### 阶段三：增强功能

#### 3.10 提交历史面板

- 每题的提交记录列表
- 可查看历史代码 + 执行结果

#### 3.11 收藏功能对接

- 点击收藏调用 `POST /favorites`
- 收藏列表筛选

#### 3.12 排行榜页面

- 展示用户排名
- 解题数 + 通过率 + 连续天数

---

## 四、技术选型

| 需求 | 方案 | 理由 |
|------|------|------|
| 代码编辑器 | `@monaco-editor/react` | VS Code 同款，功能最全 |
| 测试用例验证 | 沙箱 stdin 注入 + stdout 比对 | LeetCode 标准模式 |
| 代码持久化 | localStorage | 轻量，无需后端 |
| 提交历史 | 现有 Submission 表 + 新 API | 无需新表 |
| 排行榜 | SQL 聚合查询 | 无需额外基础设施 |

**不引入的技术**：
- ❌ Docker 隔离沙箱 → 现有进程级沙箱足够
- ❌ 在线判题队列 → 同步执行即可（非高并发场景）

---

## 五、实施路线图

### 第一期：后端测试验证闭环

| 序号 | 任务 | 文件 |
|------|------|------|
| 1 | 沙箱支持 stdin 注入 | `backend/internal/sandbox/runner.go` |
| 2 | ValidateWithCases 多用例验证 | `backend/internal/sandbox/runner.go` |
| 3 | submitCode 返回逐条结果 + 更新统计 | `backend/internal/api/course_problem.go` |
| 4 | 提交历史接口 | `backend/internal/api/course_problem.go` |
| 5 | 排行榜接口 | `backend/internal/api/course_problem.go` |
| 6 | 注册路由 | `backend/internal/api/server.go` |

### 第二期：前端编辑器 + 体验升级

| 序号 | 任务 | 文件 |
|------|------|------|
| 7 | 安装 @monaco-editor/react | `frontend/` |
| 8 | 替换 textarea 为 Monaco | `frontend/src/app/practice/page.tsx` |
| 9 | 真实测试结果展示 | 同上 |
| 10 | 题目列表 + 切题 | 同上 |
| 11 | 代码 localStorage 持久化 | 同上 |
| 12 | 提交历史面板 | 同上 |
| 13 | 收藏功能对接 | 同上 |
| 14 | 支持 C++ 语言选项 | 同上 |

---

## 六、API 设计

### 新增/修改的接口

```
POST /api/v1/code/run                 # 修改：支持 stdin + problem_id
POST /api/v1/code/submit              # 修改：返回 case_results
GET  /api/v1/problems/:id/submissions # 新增：某题提交历史
GET  /api/v1/submissions              # 新增：全部提交历史（分页）
GET  /api/v1/leaderboard              # 新增：排行榜
```

---

## 七、实施进度

| 序号 | 任务 | 状态 | 文件 |
|------|------|------|------|
| 1 | 沙箱 stdin 注入（runInternal） | ✅ 完成 | `backend/internal/sandbox/runner.go` |
| 2 | ValidateWithCases 多用例验证 | ✅ 完成 | `backend/internal/sandbox/runner.go` |
| 3 | CaseResult/CaseResults 类型 | ✅ 完成 | `backend/internal/sandbox/runner.go` |
| 4 | submitCodeV2 逐条测试结果 | ✅ 完成 | `backend/internal/api/practice_v2.go`（新） |
| 5 | 提交后更新 passRate/submissions | ✅ 完成 | `backend/internal/api/practice_v2.go` |
| 6 | 提交历史接口 GET /problems/:id/submissions | ✅ 完成 | `backend/internal/api/practice_v2.go` |
| 7 | 全部提交历史 GET /submissions | ✅ 完成 | `backend/internal/api/practice_v2.go` |
| 8 | 排行榜接口 GET /leaderboard | ✅ 完成 | `backend/internal/api/practice_v2.go` |
| 9 | submitCode 委托到 submitCodeV2 | ✅ 完成 | `backend/internal/api/course_problem.go` |
| 10 | 路由注册 | ✅ 完成 | `backend/internal/api/server.go` |
| 11 | 安装 @monaco-editor/react | ✅ 完成 | `frontend/package.json` |
| 12 | Monaco 编辑器替换 textarea | ✅ 完成 | `frontend/src/app/practice/page.tsx` |
| 13 | 真实测试结果展示（逐条 ✓/✗） | ✅ 完成 | 前端同上 |
| 14 | 题目列表抽屉 + 上一题/下一题 | ✅ 完成 | 前端同上 |
| 15 | 代码 localStorage 持久化 | ✅ 完成 | 前端同上 |
| 16 | C++ 语言选项 | ✅ 完成 | 前端同上 |
| 17 | 提交历史面板 | ✅ 完成 | 前端同上 |
| 18 | 收藏功能对接 API | ✅ 完成 | 前端同上 |

### 新增文件

- `backend/internal/api/practice_v2.go` — submitCodeV2 + 提交历史 + 排行榜
- `frontend/documents/PRACTICE_IMPROVEMENT_PLAN.md` — 本文档

### 修改文件

- `backend/internal/sandbox/runner.go` — RunWithInput + ValidateWithCases + CaseResult + runInternal
- `backend/internal/api/course_problem.go` — submitCode 委托到 submitCodeV2
- `backend/internal/api/server.go` — 注册 4 条新路由
- `frontend/src/app/practice/page.tsx` — 全面重写（Monaco + 真实结果 + 题目列表 + 持久化）
- `frontend/package.json` — 新增 @monaco-editor/react 依赖
