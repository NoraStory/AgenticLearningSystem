'use client';

import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

export interface Insights {
  behavior_distribution: { name: string; value: number }[];
  time_slot_preference: { hour: number; minutes: number }[];
  interest_distribution: { category: string; cnt: number }[];
  activity_trend: { date: string; events: number; minutes: number }[];
  consistency: { total_minutes: number; avg_daily_minutes: number; active_days_7: number; best_weekday: string };
}

const PIE_COLORS = ['#6366f1', '#22c55e', '#3b82f6', '#f97316', '#ef4444', '#a855f7', '#14b8a6'];

export default function InsightsCards({ insights }: { insights: Insights | null }) {
  if (!insights) {
    return <p className="text-sm text-muted-foreground/60 py-2">开始学习后，这里会生成你的专属洞察</p>;
  }
  const hasData =
    insights.behavior_distribution.length > 0 ||
    insights.time_slot_preference.some((h) => h.minutes > 0) ||
    insights.interest_distribution.length > 0;
  if (!hasData) {
    return <p className="text-sm text-muted-foreground/60 py-2">开始学习后，这里会生成你的专属洞察</p>;
  }

  return (
    <div className="grid grid-cols-2 gap-4">
      {/* 行为分布 */}
      <div className="bg-surface rounded-lg p-4">
        <h4 className="text-xs font-semibold text-foreground mb-3">近 30 天行为分布</h4>
        {insights.behavior_distribution.length === 0 ? (
          <p className="text-xs text-muted-foreground/60">暂无行为数据</p>
        ) : (
          <ResponsiveContainer width="100%" height={140}>
            <PieChart>
              <Pie
                data={insights.behavior_distribution}
                dataKey="value"
                nameKey="name"
                innerRadius={30}
                outerRadius={55}
                paddingAngle={2}
              >
                {insights.behavior_distribution.map((_, i) => (
                  <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        )}
      </div>

      {/* 时段偏好 */}
      <div className="bg-surface rounded-lg p-4">
        <h4 className="text-xs font-semibold text-foreground mb-3">学习时段偏好（24h）</h4>
        <ResponsiveContainer width="100%" height={140}>
          <BarChart data={insights.time_slot_preference}>
            <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
            <XAxis dataKey="hour" tick={{ fontSize: 9 }} interval={3} />
            <YAxis tick={{ fontSize: 9 }} />
            <Tooltip />
            <Bar dataKey="minutes" fill="#6366f1" radius={[2, 2, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* 兴趣分布 */}
      <div className="bg-surface rounded-lg p-4">
        <h4 className="text-xs font-semibold text-foreground mb-3">兴趣方向（浏览/做题）</h4>
        {insights.interest_distribution.length === 0 ? (
          <p className="text-xs text-muted-foreground/60">暂无浏览记录</p>
        ) : (
          <div className="space-y-2">
            {insights.interest_distribution.slice(0, 5).map((it, i) => (
              <div key={it.category} className="flex items-center gap-2">
                <span
                  className="w-2 h-2 rounded-full shrink-0"
                  style={{ background: PIE_COLORS[i % PIE_COLORS.length] }}
                />
                <span className="text-xs text-foreground flex-1 truncate">{it.category}</span>
                <span className="text-xs text-muted-foreground">{it.cnt} 次</span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 学习一致性 */}
      <div className="bg-surface rounded-lg p-4">
        <h4 className="text-xs font-semibold text-foreground mb-3">学习一致性</h4>
        <div className="space-y-1.5 text-xs text-muted-foreground">
          <p>近 30 天累计学习 {Math.round(insights.consistency.total_minutes / 60)} 小时</p>
          <p>日均学习 {Math.round(insights.consistency.avg_daily_minutes)} 分钟</p>
          <p>近 7 天活跃 {insights.consistency.active_days_7} 天</p>
          <p>最常学习：{insights.consistency.best_weekday}</p>
        </div>
      </div>
    </div>
  );
}
