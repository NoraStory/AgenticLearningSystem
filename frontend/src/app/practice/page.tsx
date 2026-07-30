'use client';

import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';
import { Heart, ChevronLeft, ChevronRight } from 'lucide-react';

// 模拟题目数据
const fallbackProblem = {
  id: 42,
  title: '两数之和',
  difficulty: '简单',
  passRate: '48.5%',
  submissions: '3,847,291',
  description: '给定一个整数数组 nums 和一个整数目标值 target，请你在该数组中找出和为目标值 target 的那两个整数，并返回它们的数组下标。',
  examples: [
    { input: 'nums = [2,7,11,15], target = 9', output: '[0,1]', explanation: '因为 nums[0] + nums[1] == 9 ，返回 [0, 1] 。' },
    { input: 'nums = [3,2,4], target = 6', output: '[1,2]', explanation: '' },
    { input: 'nums = [3,3], target = 6', output: '[0,1]', explanation: '' },
  ],
  constraints: ['2 <= nums.length <= 10^4', '-10^9 <= nums[i] <= 10^9', '-10^9 <= target <= 10^9', '只会存在一个有效答案'],
  tags: ['数组', '哈希表'],
};

const fallbackCodeTemplates: Record<string, string> = {
  rust: `use std::collections::HashMap;

impl Solution {
    pub fn two_sum(nums: Vec<i32>, target: i32) -> Vec<i32> {
        let mut map: HashMap<i32, usize> = HashMap::new();
        
        for (i, &num) in nums.iter().enumerate() {
            let complement = target - num;
            if let Some(&idx) = map.get(&complement) {
                return vec![idx as i32, i as i32];
            }
            map.insert(num, i);
        }
        
        vec![]
    }
}`,
  python: `class Solution:
    def twoSum(self, nums: List[int], target: int) -> List[int]:
        seen = {}
        for i, num in enumerate(nums):
            complement = target - num
            if complement in seen:
                return [seen[complement], i]
            seen[num] = i
        return []`,
  javascript: `/**
 * @param {number[]} nums
 * @param {number} target
 * @return {number[]}
 */
var twoSum = function(nums, target) {
    const seen = new Map();
    for (let i = 0; i < nums.length; i++) {
        const complement = target - nums[i];
        if (seen.has(complement)) {
            return [seen.get(complement), i];
        }
        seen.set(nums[i], i);
    }
    return [];
};`,
};

export default function PracticePage() {
  const [language, setLanguage] = useState<'rust' | 'python' | 'javascript'>('python');
  const [activeTab, setActiveTab] = useState<'result' | 'stdout' | 'info'>('result');
  const [isFavorite, setIsFavorite] = useState(false);
  const [problem, setProblem] = useState(fallbackProblem);
  const [codeTemplates, setCodeTemplates] = useState(fallbackCodeTemplates);
  const [code, setCode] = useState(fallbackCodeTemplates.python);
  const [runResult, setRunResult] = useState<{ status: string; stdout?: string; stderr?: string; execution_time_ms?: number; passed_cases?: number; total_cases?: number } | null>(null);
  const [isRunning, setIsRunning] = useState(false);

  useEffect(() => {
    Promise.all([
      apiFetch<Record<string, unknown>>('/api/v1/problems/1'),
      apiFetch<Record<string, string>>('/api/v1/problems/1/templates'),
    ]).then(([detail, templates]) => {
      setProblem({
        ...fallbackProblem, id: Number(detail.id), title: String(detail.title), difficulty: String(detail.difficulty),
        passRate: String(Number(detail.pass_rate).toFixed(1)) + '%', submissions: Number(detail.submissions).toLocaleString(),
        description: String(detail.description), examples: detail.examples as typeof fallbackProblem.examples,
        constraints: detail.constraints as string[], tags: detail.tags as string[],
      });
      setCodeTemplates({ ...fallbackCodeTemplates, ...templates });
      setCode(templates.python || fallbackCodeTemplates.python);
    }).catch(() => undefined);
  }, []);

  const changeLanguage = (next: 'rust' | 'python' | 'javascript') => {
    setLanguage(next);
    setCode(codeTemplates[next] || '');
  };

  const execute = async (submit: boolean) => {
    setIsRunning(true);
    setActiveTab('stdout');
    try {
      const result = await apiFetch<NonNullable<typeof runResult>>(submit ? '/api/v1/code/submit' : '/api/v1/code/run', {
        method: 'POST', body: JSON.stringify({ problem_id: problem.id, language, code }),
      });
      setRunResult(result);
      setActiveTab(submit ? 'info' : 'stdout');
    } catch (error) {
      setRunResult({ status: 'error', stderr: error instanceof Error ? error.message : '执行失败' });
    } finally {
      setIsRunning(false);
    }
  };

  return (
    <>
      {/* 顶部信息栏 */}
      <div className="px-8 py-4 border-b border-outline/10 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <button className="p-1 text-muted-foreground hover:text-foreground">
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span className="text-sm text-muted-foreground">#{problem.id}</span>
            <h1 className="text-lg font-bold text-foreground">{problem.title}</h1>
          </div>
          <span className="px-2 py-0.5 bg-green-500/10 text-green-600 text-xs font-medium rounded-sm">
            {problem.difficulty}
          </span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-muted-foreground">通过率 {problem.passRate}</span>
          <span className="text-sm text-muted-foreground">提交 {problem.submissions} 次</span>
          <button
            onClick={() => setIsFavorite(!isFavorite)}
            className={`p-1 ${isFavorite ? 'text-destructive' : 'text-muted-foreground hover:text-foreground'}`}
          >
            <Heart className={`w-4 h-4 ${isFavorite ? 'fill-current' : ''}`} />
          </button>
          <button className="p-1 text-muted-foreground hover:text-foreground">
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* 主内容区 */}
      <div className="flex h-[calc(100vh-140px)]">
        {/* 左侧：题目描述 */}
        <div className="w-[40%] border-r border-outline/10 overflow-y-auto p-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">题目描述</h2>
          <p className="text-sm text-foreground leading-relaxed mb-6">
            {problem.description}
          </p>

          <h3 className="text-sm font-semibold text-foreground mb-3">示例</h3>
          <div className="space-y-4 mb-6">
            {problem.examples.map((ex, idx) => (
              <div key={idx} className="text-sm">
                <p className="text-muted-foreground mb-1">
                  <strong>输入：</strong>{ex.input}
                </p>
                <p className="text-muted-foreground">
                  <strong>输出：</strong>{ex.output}
                </p>
                {ex.explanation && (
                  <p className="text-muted-foreground mt-1 text-xs">
                    解释：{ex.explanation}
                  </p>
                )}
              </div>
            ))}
          </div>

          <h3 className="text-sm font-semibold text-foreground mb-3">约束条件</h3>
          <ul className="text-sm text-muted-foreground space-y-1 mb-6">
            {problem.constraints.map((c, idx) => (
              <li key={idx} className="flex items-start gap-2">
                <span className="text-primary mt-1">•</span>
                <code className="text-xs bg-surface-container px-1.5 py-0.5 rounded">{c}</code>
              </li>
            ))}
          </ul>

          <div className="flex flex-wrap gap-2">
            {problem.tags.map((tag) => (
              <span
                key={tag}
                className="text-xs text-muted-foreground bg-surface-container px-2.5 py-1 rounded-sm"
              >
                {tag}
              </span>
            ))}
          </div>
        </div>

        {/* 右侧：代码编辑器 */}
        <div className="flex-1 flex flex-col">
          {/* 编辑器工具栏 */}
          <div className="flex items-center justify-between px-4 py-2 border-b border-outline/10">
            <select
              value={language}
              onChange={(e) => changeLanguage(e.target.value as 'rust' | 'python' | 'javascript')}
              className="text-sm border border-outline rounded px-2 py-1 bg-surface text-foreground"
            >
              <option value="rust">Rust</option>
              <option value="python">Python</option>
              <option value="javascript">JavaScript</option>
            </select>
            <div className="flex items-center gap-2">
              <button onClick={() => setCode(codeTemplates[language] || '')} className="px-3 py-1.5 text-xs text-muted-foreground hover:text-foreground hover:bg-surface-container rounded transition-colors">
                重置
              </button>
              <button disabled={isRunning} onClick={() => execute(false)} className="px-3 py-1.5 text-xs font-medium text-primary bg-primary/10 hover:bg-primary/20 rounded transition-colors disabled:opacity-50">
                {isRunning ? '运行中...' : '运行代码'}
              </button>
              <button disabled={isRunning} onClick={() => execute(true)} className="px-3 py-1.5 text-xs font-medium text-on-primary bg-primary hover:bg-primary/90 rounded transition-colors disabled:opacity-50">
                提交答案
              </button>
            </div>
          </div>

          {/* 代码区域 */}
          <div className="flex-1 bg-[#1e1e2e] overflow-auto p-4">
            <textarea
              value={code}
              onChange={(event) => setCode(event.target.value)}
              spellCheck={false}
              className="w-full h-full resize-none bg-transparent outline-none text-sm font-mono leading-relaxed text-[#cdd6f4]"
            />
          </div>

          {/* 输出区域 */}
          <div className="h-48 border-t border-outline/10">
            <div className="flex border-b border-outline/10">
              {[
                { key: 'result', label: '测试结果' },
                { key: 'stdout', label: 'stdout' },
                { key: 'info', label: '执行结果' },
              ].map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setActiveTab(tab.key as 'result' | 'stdout' | 'info')}
                  className={`px-4 py-2 text-xs font-medium transition-colors ${
                    activeTab === tab.key
                      ? 'text-primary border-b-2 border-primary'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
            <div className="p-4 overflow-auto h-[calc(100%-36px)]">
              {activeTab === 'result' && (
                <div className="space-y-2 text-sm">
                  {[
                    { input: '[2,7,11,15], 9', expected: '[0,1]', actual: '[0,1]', pass: true },
                    { input: '[3,2,4], 6', expected: '[1,2]', actual: '[1,2]', pass: true },
                    { input: '[3,3], 6', expected: '[0,1]', actual: '[0,1]', pass: true },
                  ].map((tc, idx) => (
                    <div key={idx} className="flex items-center gap-3">
                      <span className={tc.pass ? 'text-success' : 'text-destructive'}>
                        {tc.pass ? '✓' : '✗'}
                      </span>
                      <span className="text-muted-foreground text-xs">
                        输入: {tc.input}
                      </span>
                      <span className="text-xs">
                        输出: {tc.actual}
                      </span>
                    </div>
                  ))}
                </div>
              )}
              {activeTab === 'stdout' && (
                <pre className="text-xs text-muted-foreground font-mono">
                  {runResult?.stdout || runResult?.stderr || '点击运行代码查看输出。'}
                </pre>
              )}
              {activeTab === 'info' && (
                <div className="space-y-2 text-sm">
                  <p className={runResult?.status === 'accepted' || runResult?.status === 'success' ? 'text-success font-medium' : 'text-destructive font-medium'}>{runResult ? `状态：${runResult.status}` : '尚未提交'}</p>
                  <p className="text-muted-foreground text-xs">执行时间: {runResult?.execution_time_ms ?? 0}ms</p>
                  <p className="text-muted-foreground text-xs">通过用例: {runResult?.passed_cases ?? 0}/{runResult?.total_cases ?? 0}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
