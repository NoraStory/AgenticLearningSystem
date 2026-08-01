# AI 助手模块 — 工作流现状审计与开源改造方案

> 聚焦 `backend/internal/api/agent.go` + `tool_planner.go` + `tool_helpers.go`(手写的"智能路由 + 工具规划 + 顺序执行"链路)与 `frontend/src/app/agent/chat/page.tsx`(工作流展示)。仅方案,不含代码改动。

## 一、现状盘点(已源码级复核)

### 1.1 一次对话的完整处理链路

```
POST /agent/chat
  ├─ routeAgent(agent.go:179)      关键词路由 6 分支 → agent_type(career/code-review/...)
  ├─ planAgentTools(tool_planner.go:58)
  │    ├─ 20+ 个 containsAny 关键词分支 → 选出工具列表
  │    ├─ 末尾 llmShouldSearch(agent.go:786) 只返回 yes/no,不规划组合
  │    └─ scheduleTools 按 phase(4 档)排序,execution 恒为串行
  ├─ 逐个 executeWithRetry(重试 5 次,1s/2s/3s 退避)
  ├─ 结果拼接进 prompt → llm.StreamChatWithImages 输出
  └─ 全程通过 SSE(agent_route/workflow_start/workflow_step/tool_call/tool_result/tool_failure/token/done)推送
```

### 1.2 核心问题

| # | 问题 | 证据 |
|---|---|---|
| A1 | **路由与工具规划纯关键词匹配**,"智能"名不副实。`routeAgent` 只查 6 个词,"我的程序出问题了"路由不到 code-review;`planAgentTools` 的 ~20 个 `containsAny` 分支是硬编码,无法泛化到未枚举的意图;且规划与执行分离,LLM 无法按真实上下文调整工具组合 | agent.go:179、tool_planner.go:58 |
| A2 | **`llmShouldSearch` 只做二值判断**,每次漏网请求还要多付一次 LLM 调用,不产出工具组合 | tool_planner.go:786 |
| A3 | **工具执行全串行、无 DAG 调度**。`scheduleTools` 只按 phase 排序列化执行,`DependsOn` 字段只声明不调度;15 个工具要么全跑要么不跑,web_search 结果无法决定是否继续 | tool_planner.go:154 |
| A4 | **对话页四个协作模式全是死按钮**(串行接力/并行合并/辩论/师徒),`collaboration_mode` 参数后端只在 workflow_start 事件里原样回显,无任何逻辑 | chat/page.tsx:386-393 |
| A5 | **UI 与后端的 agent 命名是两套体系**。前端 agents 数组是 learning/reviewer/tutor/mentor/coach/community,后端 `routeAgent` 返回 career/code-review/problem-explain/planner/project/learning-assistant,两者对不上,前端显示的全是兜底"学习助手"图标 | chat/page.tsx:11-18 vs agent.go:179 |
| A6 | **`WorkflowExecution` 表只记录运行痕迹**,`status`/`current_node`/`result_json` 写死 running→completed,前端从不查询,confirmWorkflow(agent.go:441)无人调用 | models.go:307 |
| A7 | **记忆只是截断拼 prompt**:最近 12 条消息、单条截 2400 字、总上限 12000 字,`buildConversationPrompt` 纯字符串拼接,无摘要压缩(超长会话前 12 条被粗暴丢弃)、无用户画像(agent_profile.go 有 profile 表但 agentChat 从不读取) | agent_memory.go:20 |
| A8 | **工具执行无沙箱隔离**(仅 subprocess+黑名单,audit 报告 B1 已列)、**无 API 限流**(B3 已列)、**无 MCP 协议**(工具是 Go 硬编码 switch,外部工具无法接入) | tool_helpers.go、tool_planner.go:185 |

### 1.3 已完成的先修项(上一轮已落地,不再重复)

- SSE 前端解析已换 eventsource-parser + AbortController(chat/page.tsx:4,257)
- Markdown/代码高亮/Mermaid 官方渲染(components/Markdown.tsx、Mermaid.tsx)
- LLM 已禁用 thinking、超时 180s(llm/client.go:86)

## 二、开源方案调研(能开源就用开源)

### 2.1 候选对比

| 方案 | 简介 | 许可 | 与现网契合度 | 结论 |
|---|---|---|---|---|
| **字节跳动 Eino**(cloudwego/eino) | 组件化 Go Agent 框架,Graph/Workflow/Chain 编排、流式处理、Callbacks、内置 Agent | Apache-2.0 | 高:**eino-ext 官方有 `model/ark`(ToolCallingChatModel,原生对接火山 Ark/豆包)**、`tool/mcp`、`retriever`;字节 60+ 业务线自用;LLM 供应商就是豆包 | **首选** |
| **smallnest/langgraphgo**(LangGraph Go 移植) | ReAct/CreateAgent/Supervisor、并行、human-in-the-loop、状态机,`llms/doubao` 原生支持 Ark(基于 volcengine-go-sdk) | MIT | 中:偏状态机,SSE 流式事件与现在的前端 WorkflowStep 模型需要适配层;DDD 风格与现有 gin handler 混杂 | 备选 |
| **openai/openai-go + function calling** | 官方 Go SDK,Ark 兼容 OpenAI 协议,LLM 原生 tool loop | Apache-2.0 | 中低:只解决 A1,不解决编排/A4/A6/A7 | 局部方案,不如 Eino 一次到位 |
| **mark3labs/mcp-go**(MCP 协议) | Go 官方 MCP 客户端/服务端 SDK,stdio/SSE/streamable-http | MIT | 单点替换 toolCatalog 为 MCP 工具注册(可被 Eino 的 `tool/mcp` 复用) | 并入 Eino 方案 |

### 2.2 推荐路线:Eino 原生 Ark 集成(替代手写规划链路)

```
现网(手写):            改造后(Eino 编排):
routeAgent(关键词)  →   ChatModel/agenticark(原生 ToolCallingChatModel)
planAgentTools(关键词) →  ReAct Agent:LLM 自行决策工具调用序列,支持多轮循环
executeWithRetry(串行) →  eino Tool 组件 + 失败重试回调(现有 executeWithRetry 保留在工具内部)
手拼 prompt + 工具结果 →  Eino Messages 协议(工具结果自动回填对话)
```

- **`model/ark`**(eino-ext 官方组件):直接对接现网 `ARK_API_KEY`/`ARK_BASE_URL`/`ARK_MODEL`,支持 streaming、vision(现有图片上传能力保留)、工具调用;**thinking 深度思考开关官方参数化**(替代现在手写的 `Thinking:{Type:"disabled"}` hack)。
- **工具注册**:把现有 15 个 `executeAgentTool` 的 switch 分支逐个包成 eino `Tool`(名称/描述/JSON Schema),描述来自 `toolCatalog` 已有文案,零信息损失;eino 有 `tool/mcp` 组件,后续外部工具可直接以 MCP server 接入(顺手解决 A8 的 MCP 部分)。
- **工作流事件**:eino 的 `Callbacks` 回调(pre/post run、tool start/end、stream chunk)直接映射现有 SSE 事件流,前端 `WorkflowStep[]` 渲染**零改动**;`RunInfo` 序列化后写入 `WorkflowExecution.ResultJSON`(顺手解决 A6,前端可加"工作流详情"回查)。
- **许可**:Apache-2.0,允许闭源使用,符合项目既有依赖风格(gin/gorm 同许可族)。

### 2.3 分期落地

| 阶段 | 内容 | 验证 |
|---|---|---|
| **P0(低风险先行)** | ① `chat/page.tsx`:agents 列表改为读后端 `/agent/tools`/新增 `/agent/agents` 返回真实路由结果映射,修 A5;② 删除四个假协作模式按钮,改为单选"自动(动态规划)"并去掉 `collaboration_mode` 参数传递,修 A4 | 浏览器实测路由后 agent 图标/名称正确;模式 UI 不再有死按钮 |
| **P1(核心)** | 引入 eino + `model/ark` + `tool/mcp`;把 15 个工具包成 eino Tool;`agentChat` 改为 ReAct 循环(最多 N 轮工具调用);SSE 事件经 eino Callbacks 映射保持现有协议不变;`llmShouldSearch` 删除(规划交给 LLM);保留 `executeWithRetry` 在工具内做重试;`Thinking` 用官方参数 | `go test ./...`;curl 走通"查最新 Go 版本→搜站内课程→总结"多轮工具链;SSE 事件顺序与前端兼容 |
| **P2(增强)** | 记忆升级:超长会话摘要压缩(先实现长度感知截断→摘要合并),agentChat 读取 UserProfile 注入人设;`WorkflowExecution` 写完整 RunInfo,前端工作流卡片可展开查看 | 长会话 3 轮后仍能回答指代;工作流详情可回查 |
| **P3(可选)** | 限流/沙箱(承接 audit B1/B3,本方案不展开) | — |

### 2.4 保留项(不建议动)

- **Session 模型**(agent_sessions.go):按 session_id 聚合、以 session_messages 为唯一真相源,设计正确,保留。
- **SSE 协议与 WorkflowStep 前端结构**:与 Eino Callbacks 天然对接,保留。
- **routeAgent 的 agent_type 语义**:改造后由 LLM 从 tool 描述自行决策,`agent_type` 仅作为展示标签保留(前端 agents 数组改为动态映射)。

## 三、风险与注意点

- **Eino 依赖体积**:eino + eino-ext 会引入一批组件依赖,需先 `go get` 验证与 Go 1.26 兼容;若 P1 联调卡壳,可降级为"openai-go function calling + 手写 ReAct 循环"(方案 2.1 第 3 行)先交付 A1。
- **stream 语义差异**:现有前端按 token 增量渲染,Eino 流式输出到 SSE 的映射要保证 chunk 顺序与 flush 节奏(现在 `c.Writer.Flush()` 每 token 一次,体验基线不能回退)。
- **thinking 禁用**:豆包带 reasoning 的模型在 eino `model/ark` 下仍要显式配置 thinking disabled(现网已经吃过坑,见 llm/client.go:86)。
- **工具描述即契约**:LLM 决策质量取决于 `toolCatalog` 的 desc 是否具体,落地时逐个补齐(如"联网搜索:支持中文关键词,优先返回最新技术资料")。
