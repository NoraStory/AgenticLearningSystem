// 轻量埋点 SDK:批量上报 + sendBeacon + localStorage 离线兜底。
// 不引入第三方分析平台;事件写入后端 UserActivity 表,画像页"学习洞察"消费。

const EVENT_QUEUE_KEY = 'codeforge_event_queue';
const FLUSH_BATCH = 5;
const FLUSH_INTERVAL_MS = 3000;
const QUEUE_MAX = 200;

type TrackProps = Record<string, string | number | boolean | null | undefined>;

interface TrackEvent {
  name: string;
  props?: TrackProps;
  at: number;
}

function loadQueue(): TrackEvent[] {
  try {
    const raw = localStorage.getItem(EVENT_QUEUE_KEY);
    return raw ? (JSON.parse(raw) as TrackEvent[]) : [];
  } catch {
    return [];
  }
}

function saveQueue(queue: TrackEvent[]) {
  try {
    const trimmed = queue.slice(-QUEUE_MAX);
    if (trimmed.length === 0) {
      localStorage.removeItem(EVENT_QUEUE_KEY);
    } else {
      localStorage.setItem(EVENT_QUEUE_KEY, JSON.stringify(trimmed));
    }
  } catch {
    // localStorage 不可用(隐私模式)时静默丢弃
  }
}

let timer: ReturnType<typeof setTimeout> | null = null;

/** 上报埋点事件,自动批量 + 离线缓存,失败静默不阻塞业务。 */
export function track(name: string, props?: TrackProps) {
  if (typeof window === 'undefined') return;
  const queue = [...loadQueue(), { name, props, at: Date.now() }];
  saveQueue(queue);
  if (queue.length >= FLUSH_BATCH) {
    flush();
  } else if (!timer) {
    timer = setTimeout(() => {
      timer = null;
      flush();
    }, FLUSH_INTERVAL_MS);
  }
}

function flush() {
  const queue = loadQueue();
  if (queue.length === 0) return;
  saveQueue([]);
  try {
    const body = JSON.stringify({
      events: queue.map(({ name, props }) => ({ name, props })),
    });
    const blob = new Blob([body], { type: 'application/json' });
    if (!navigator.sendBeacon('/api/v1/events', blob)) {
      // beacon 不可用(如 http 非安全上下文)时回退 fetch,失败再放回队列
      fetch('/api/v1/events', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body,
        keepalive: true,
      }).catch(() => {
        const current = loadQueue();
        saveQueue([...queue, ...current]);
      });
    }
  } catch {
    // 序列化失败:丢弃本批,避免死循环
  }
}

/** 页面浏览埋点(由 Tracker 组件在路由变化时调用)。 */
export function trackPageView(pathname: string) {
  track('page_view', { path: pathname });
}
