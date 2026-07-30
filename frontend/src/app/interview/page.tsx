'use client';

import { useState } from 'react';
import { apiFetch } from '@/lib/api';
import { Clock, Target, Trophy, Code, CheckCircle, Lightbulb, Play, RotateCcw } from 'lucide-react';

// 模拟笔试题目
const fallbackExamQuestions = [
  {
    id: 1,
    type: 'code', // 代码题
    category: '算法',
    difficulty: '中等',
    title: '两数之和',
    description: '给定一个整数数组 nums 和一个整数目标值 target，请你在该数组中找出和为目标值的那两个整数，并返回它们的数组下标。',
    constraints: ['2 <= nums.length <= 10^4', '-10^9 <= nums[i] <= 10^9', '-10^9 <= target <= 10^9', '只会存在一个有效答案'],
    timeLimit: 20,
    score: 100,
    example: {
      input: 'nums = [2,7,11,15], target = 9',
      output: '[0,1]',
      explanation: '因为 nums[0] + nums[1] == 9，返回 [0, 1]',
    },
  },
  {
    id: 2,
    type: 'text', // 文字题
    category: '系统设计',
    difficulty: '中等',
    title: '设计一个短链接系统',
    description: '设计一个 URL 短链接服务，类似于 bit.ly。用户可以输入长 URL，系统生成短 URL，访问短 URL 时重定向到原始长 URL。',
    constraints: ['支持每天生成 1000 万短链接', '短链接长度为 7 个字符', '支持链接过期和统计功能'],
    timeLimit: 30,
    score: 100,
  },
  {
    id: 3,
    type: 'text', // 文字题
    category: 'Python',
    difficulty: '简单',
    title: '解释 Python 的 GIL',
    description: '请解释 Python 的全局解释器锁（GIL）是什么，它对多线程有什么影响，以及如何绕过这个限制。',
    constraints: ['需要涵盖原理', '需要给出实际例子', '需要说明解决方案'],
    timeLimit: 10,
    score: 100,
  },
  {
    id: 4,
    type: 'code', // 代码题
    category: '数据结构',
    difficulty: '简单',
    title: '反转链表',
    description: '给你单链表的头节点 head，请你反转链表，并返回反转后的链表。',
    constraints: ['链表中节点的数目范围是 [0, 5000]', '-5000 <= Node.val <= 5000'],
    timeLimit: 15,
    score: 100,
    example: {
      input: 'head = [1,2,3,4,5]',
      output: '[5,4,3,2,1]',
    },
  },
];

export default function ExamPage() {
  const [currentQuestion, setCurrentQuestion] = useState(0);
  const [isStarted, setIsStarted] = useState(false);
  const [timeLeft, setTimeLeft] = useState(0);
  const [answers, setAnswers] = useState<string[]>([]);
  const [currentAnswer, setCurrentAnswer] = useState('');
  const [language, setLanguage] = useState('python');
  const [examQuestions, setExamQuestions] = useState(fallbackExamQuestions);
  const [examId, setExamId] = useState('');

  const question = examQuestions[currentQuestion];
  const isCodeQuestion = question?.type === 'code';

  const startExam = async () => {
    try {
      const data = await apiFetch<{ exam_id: string; questions: Array<Record<string, unknown>> }>('/api/v1/interview/exams/generate', {
        method: 'POST', body: JSON.stringify({ direction: '全栈开发', difficulty: '中等', question_count: 4 }),
      });
      const questions = data.questions.map((item, index) => ({
        ...fallbackExamQuestions[index % fallbackExamQuestions.length],
        id: index + 1,
        type: String(item.type), category: String(item.category), difficulty: String(item.difficulty),
        title: String(item.title), description: String(item.description), constraints: item.constraints as string[],
        timeLimit: Number(item.timeLimit || item.time_limit), score: Number(item.score),
        example: item.example as typeof fallbackExamQuestions[number]['example'],
      }));
      setExamId(data.exam_id);
      setExamQuestions(questions);
      setCurrentQuestion(0);
      setAnswers([]);
      setIsStarted(true);
      setTimeLeft(questions[0].timeLimit * 60);
    } catch (error) {
      console.error(error);
    }
  };

  const submitAnswer = async () => {
    const nextAnswers = [...answers, currentAnswer];
    setAnswers(nextAnswers);
    setCurrentAnswer('');
    if (currentQuestion < examQuestions.length - 1) {
      setCurrentQuestion(currentQuestion + 1);
      setTimeLeft(examQuestions[currentQuestion + 1].timeLimit * 60);
    } else {
      if (examId) {
        const result = await apiFetch<{ score: number }>('/api/v1/interview/exams/' + examId + '/submit', {
          method: 'POST', body: JSON.stringify({ answers: nextAnswers.map((answer, index) => ({ question_id: String(index + 1), answer })) }),
        });
        window.alert('本次笔试得分：' + result.score);
      }
      setIsStarted(false);
    }
  };
  const codeTemplates: Record<string, string> = {
    python: `class Solution:
    def solve(self, nums: list[int], target: int) -> list[int]:
        # 在这里写你的代码
        pass`,
    cpp: `class Solution {
public:
    vector<int> solve(vector<int>& nums, int target) {
        // 在这里写你的代码
        
    }
};`,
    javascript: `/**
 * @param {number[]} nums
 * @param {number} target
 * @return {number[]}
 */
function solve(nums, target) {
    // 在这里写你的代码
    
}`,
  };

  return (
    <div className="max-w-6xl mx-auto">
      {/* 页面标题 */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground">笔试模拟</h1>
        <p className="text-muted-foreground mt-2">
          AI 出题、自动评分、代码题支持在线编程
        </p>
      </div>

      {!isStarted ? (
        /* 笔试准备 */
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="p-6 bg-surface border border-border rounded-xl">
            <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
              <Target className="w-4 h-4 text-primary" />
              笔试设置
            </h3>
            <div className="space-y-4">
              <div>
                <label className="text-sm text-muted-foreground">笔试方向</label>
                <select className="w-full mt-1 px-3 py-2 bg-surface-container border-none rounded-lg text-sm text-foreground">
                  <option>Python 后端开发</option>
                  <option>C++ 系统开发</option>
                  <option>算法与数据结构</option>
                  <option>AI Agent 开发</option>
                  <option>全栈综合</option>
                </select>
              </div>
              <div>
                <label className="text-sm text-muted-foreground">难度级别</label>
                <select className="w-full mt-1 px-3 py-2 bg-surface-container border-none rounded-lg text-sm text-foreground">
                  <option>初级（1-3年经验）</option>
                  <option>中级（3-5年经验）</option>
                  <option>高级（5年以上经验）</option>
                </select>
              </div>
              <div>
                <label className="text-sm text-muted-foreground">题目数量</label>
                <select className="w-full mt-1 px-3 py-2 bg-surface-container border-none rounded-lg text-sm text-foreground">
                  <option>3 题（快速模拟）</option>
                  <option>5 题（标准模拟）</option>
                  <option>8 题（完整模拟）</option>
                </select>
              </div>
            </div>
            <button
              onClick={startExam}
              className="w-full mt-6 py-3 bg-primary text-primary-foreground rounded-xl font-medium hover:opacity-90 transition-opacity"
            >
              开始笔试
            </button>
          </div>

          <div className="p-6 bg-surface border border-border rounded-xl">
            <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
              <Trophy className="w-4 h-4 text-primary" />
              笔试历史
            </h3>
            <div className="space-y-3">
              {[
                { date: '2024-01-15', score: 85, direction: 'Python 后端' },
                { date: '2024-01-10', score: 72, direction: '算法' },
                { date: '2024-01-05', score: 90, direction: '系统设计' },
              ].map((record, i) => (
                <div key={i} className="flex items-center justify-between p-3 bg-surface-container rounded-lg">
                  <div>
                    <div className="text-sm font-medium text-foreground">{record.direction}</div>
                    <div className="text-xs text-muted-foreground">{record.date}</div>
                  </div>
                  <div className="text-lg font-bold text-primary">{record.score}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      ) : (
        /* 笔试进行中 */
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* 题目区域 */}
          <div className={`lg:col-span-${isCodeQuestion ? '1' : '2'}`}>
            <div className="bg-surface border border-border rounded-xl p-6">
              {/* 题目头部 */}
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-3">
                  <span className="px-2 py-1 text-xs font-medium bg-primary/10 text-primary rounded">
                    第 {currentQuestion + 1} / {examQuestions.length} 题
                  </span>
                  <span className="px-2 py-1 text-xs bg-blue-100 text-blue-700 rounded">
                    {question.category}
                  </span>
                  <span className={`px-2 py-1 text-xs rounded ${
                    question.difficulty === '简单' ? 'bg-green-100 text-green-700' :
                    question.difficulty === '中等' ? 'bg-orange-100 text-orange-700' :
                    'bg-red-100 text-red-700'
                  }`}>
                    {question.difficulty}
                  </span>
                  {isCodeQuestion && (
                    <span className="px-2 py-1 text-xs bg-purple-100 text-purple-700 rounded flex items-center gap-1">
                      <Code className="w-3 h-3" />
                      代码题
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Clock className="w-4 h-4" />
                  <span>{Math.floor(timeLeft / 60)}:{(timeLeft % 60).toString().padStart(2, '0')}</span>
                </div>
              </div>

              {/* 题目内容 */}
              <h2 className="text-xl font-bold text-foreground mb-3">{question.title}</h2>
              <p className="text-foreground mb-4">{question.description}</p>
              
              {/* 示例（代码题） */}
              {isCodeQuestion && question.example && (
                <div className="mb-4 p-4 bg-surface-container rounded-xl">
                  <h4 className="text-sm font-medium text-muted-foreground mb-2">示例：</h4>
                  <div className="space-y-1 text-sm">
                    <div><span className="text-muted-foreground">输入：</span><code className="text-foreground">{question.example.input}</code></div>
                    <div><span className="text-muted-foreground">输出：</span><code className="text-foreground">{question.example.output}</code></div>
                    {question.example.explanation && (
                      <div><span className="text-muted-foreground">解释：</span><span className="text-foreground">{question.example.explanation}</span></div>
                    )}
                  </div>
                </div>
              )}

              <div className="mb-4">
                <h4 className="text-sm font-medium text-muted-foreground mb-2">约束条件：</h4>
                <ul className="space-y-1">
                  {question.constraints.map((c, i) => (
                    <li key={i} className="text-sm text-foreground flex items-center gap-2">
                      <span className="w-1 h-1 bg-muted-foreground rounded-full" />
                      <code className="text-xs bg-surface-container px-1.5 py-0.5 rounded">{c}</code>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>

          {/* 答题区域 */}
          <div className={`lg:col-span-${isCodeQuestion ? '2' : '1'}`}>
            <div className="bg-surface border border-border rounded-xl overflow-hidden">
              {isCodeQuestion ? (
                /* 代码编辑器（LeetCode 风格） */
                <div className="h-full flex flex-col">
                  {/* 编辑器头部 */}
                  <div className="flex items-center justify-between px-4 py-2 bg-[#1e1e2e] border-b border-[#313244]">
                    <div className="flex items-center gap-2">
                      <Code className="w-4 h-4 text-[#cdd6f4]" />
                      <span className="text-sm text-[#cdd6f4]">代码编辑器</span>
                    </div>
                    <select
                      value={language}
                      onChange={(e) => setLanguage(e.target.value)}
                      className="px-2 py-1 text-xs bg-[#313244] text-[#cdd6f4] border-none rounded"
                    >
                      <option value="python">Python</option>
                      <option value="cpp">C++</option>
                      <option value="javascript">JavaScript</option>
                    </select>
                  </div>
                  
                  {/* 代码区域 */}
                  <div className="flex-1 relative">
                    <textarea
                      value={currentAnswer || codeTemplates[language]}
                      onChange={(e) => setCurrentAnswer(e.target.value)}
                      placeholder="在这里写你的代码..."
                      className="w-full h-64 px-4 py-3 bg-[#1e1e2e] text-[#cdd6f4] text-sm font-mono resize-none focus:outline-none"
                      style={{ tabSize: 4 }}
                    />
                  </div>

                  {/* 编辑器底部 */}
                  <div className="flex items-center justify-between px-4 py-3 bg-[#1e1e2e] border-t border-[#313244]">
                    <div className="flex items-center gap-3">
                      <button className="px-3 py-1.5 text-xs bg-[#313244] text-[#cdd6f4] rounded hover:bg-[#45475a] flex items-center gap-1.5">
                        <Play className="w-3 h-3" />
                        运行
                      </button>
                      <button className="px-3 py-1.5 text-xs bg-[#313244] text-[#cdd6f4] rounded hover:bg-[#45475a] flex items-center gap-1.5">
                        <RotateCcw className="w-3 h-3" />
                        重置
                      </button>
                    </div>
                    <button
                      onClick={submitAnswer}
                      className="px-4 py-1.5 text-xs bg-green-600 text-white rounded hover:bg-green-700 flex items-center gap-1.5"
                    >
                      <CheckCircle className="w-3 h-3" />
                      提交答案
                    </button>
                  </div>
                </div>
              ) : (
                /* 文字答案输入 */
                <div className="p-6">
                  <label className="text-sm font-medium text-foreground mb-3 block">你的回答：</label>
                  <textarea
                    value={currentAnswer}
                    onChange={(e) => setCurrentAnswer(e.target.value)}
                    placeholder="在这里输入你的答案..."
                    className="w-full h-64 px-4 py-3 bg-surface-container border-none rounded-xl text-sm text-foreground placeholder:text-muted-foreground/40 focus:outline-none focus:ring-2 focus:ring-primary/30 resize-none"
                  />
                  <div className="flex justify-between items-center mt-4">
                    <button className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground flex items-center gap-2">
                      <Lightbulb className="w-4 h-4" />
                      提示
                    </button>
                    <button
                      onClick={submitAnswer}
                      className="px-6 py-2 bg-primary text-primary-foreground rounded-xl text-sm font-medium hover:opacity-90 flex items-center gap-2"
                    >
                      提交答案
                      <CheckCircle className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* 题目导航 */}
          <div className="lg:col-span-3">
            <div className="bg-surface border border-border rounded-xl p-4">
              <h4 className="text-sm font-medium text-muted-foreground mb-3">题目导航</h4>
              <div className="flex flex-wrap gap-2">
                {examQuestions.map((q, i) => (
                  <button
                    key={i}
                    onClick={() => {
                      setCurrentQuestion(i);
                      setTimeLeft(q.timeLimit * 60);
                    }}
                    className={`w-10 h-10 rounded-lg text-sm font-medium flex items-center justify-center transition-colors ${
                      i === currentQuestion
                        ? 'bg-primary text-primary-foreground'
                        : answers[i]
                        ? 'bg-green-100 text-green-700'
                        : 'bg-surface-container text-muted-foreground hover:bg-primary/10'
                    }`}
                  >
                    {i + 1}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
