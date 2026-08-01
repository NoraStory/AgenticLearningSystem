'use client';

import { useEffect, useRef } from 'react';

// echarts 按需加载：动态 import 触发代码分割，只在画像页懒加载；core 注册仅 graph 系列。
async function loadECharts() {
  const echarts = await import('echarts/core');
  const { GraphChart } = await import('echarts/charts');
  const { TooltipComponent } = await import('echarts/components');
  echarts.use([GraphChart, TooltipComponent]);
  return echarts;
}

export interface GraphNode {
  id: string;
  name: string;
  category: string;
  mastery: number;
  attempts: number;
}

interface Props {
  nodes: GraphNode[];
  height?: number;
}

const masteryColors = [
  { min: 0.85, color: '#22c55e' },
  { min: 0.6, color: '#3b82f6' },
  { min: 0.3, color: '#f97316' },
  { min: 0, color: '#ef4444' },
];

function nodeColor(m: number) {
  return masteryColors.find((c) => m >= c.min)?.color ?? '#ef4444';
}

// 前端从 knowledge_states 构建图谱：节点=知识点，边=同分类相邻连接 + 分类中心节点
export function buildGraph(states: GraphNode[]) {
  const nodes: Record<string, unknown>[] = [];
  const links: Record<string, unknown>[] = [];
  const byCategory = new Map<string, GraphNode[]>();
  for (const st of states) {
    const list = byCategory.get(st.category) || [];
    list.push(st);
    byCategory.set(st.category, list);
  }
  for (const [category, list] of byCategory) {
    const sorted = [...list].sort((a, b) => a.name.localeCompare(b.name));
    // 分类中心节点
    const centerId = `cat-${category}`;
    nodes.push({
      id: centerId,
      name: category,
      category,
      symbolSize: 26,
      itemStyle: { color: '#6366f1' },
      label: { show: true, fontWeight: 'bold' },
    });
    for (let i = 0; i < sorted.length; i++) {
      const st = sorted[i];
      nodes.push({
        id: st.id,
        name: st.name,
        category,
        symbolSize: Math.max(12, Math.min(28, 10 + st.attempts * 2)),
        value: st.mastery,
        itemStyle: { color: nodeColor(st.mastery) },
      });
      links.push({ source: centerId, target: st.id });
      if (i > 0) {
        links.push({ source: sorted[i - 1].id, target: st.id });
      }
    }
  }
  return { nodes, links };
}

export default function KnowledgeGraph({ nodes, height = 320 }: Props) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let cancelled = false;
    let chart: { dispose: () => void } | null = null;
    (async () => {
      if (!ref.current) return;
      const echarts = await loadECharts();
      if (cancelled || !ref.current) return;
      const { nodes: gNodes, links } = buildGraph(nodes);
      if (gNodes.length === 0) return;
      const inst = echarts.init(ref.current);
      chart = inst;
      inst.setOption({
        tooltip: {
          formatter: (params: { dataType?: string; data?: { name?: string; value?: number } }) =>
            params.dataType === 'edge'
              ? ''
              : `${params.data?.name ?? ''}<br/>掌握度: ${Math.round((params.data?.value ?? 0) * 100)}%`,
        },
        series: [
          {
            type: 'graph',
            layout: 'force',
            roam: true,
            draggable: true,
            data: gNodes,
            links,
            force: { repulsion: 120, edgeLength: [40, 100], gravity: 0.1 },
            label: { show: true, fontSize: 10, position: 'right', overflow: 'truncate', width: 60 },
            emphasis: { focus: 'adjacency' },
            lineStyle: { color: '#94a3b8', opacity: 0.5, width: 1 },
          },
        ],
      });
    })();
    return () => {
      cancelled = true;
      chart?.dispose();
    };
  }, [nodes]);

  if (nodes.length === 0) {
    return <p className="text-sm text-muted-foreground/60 py-2">完成课程或算法练习后，这里会生成你的知识图谱</p>;
  }
  return <div ref={ref} style={{ height }} className="w-full" />;
}
