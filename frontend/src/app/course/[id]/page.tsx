'use client';

import Link from 'next/link';
import { useParams } from 'next/navigation';
import { ChevronLeft, ChevronRight, Heart, Bookmark, Share2 } from 'lucide-react';
import { useState, useEffect } from 'react';
import { apiFetch } from '@/lib/api';
import { track } from '@/lib/track';

// 模拟课程数据
const fallbackCourseData: Record<string, {
  title: string;
  category: string;
  author: string;
  date: string;
  readTime: string;
  views: string;
  sections: { id: string; title: string; content: string; code?: string }[];
  tags: string[];
}> = {
  '1': {
    title: 'Rust 所有权机制深度解析',
    category: 'Rust',
    author: '陈教授',
    date: '2024 年 1 月 15 日',
    readTime: '15 分钟',
    views: '2,847',
    sections: [
      {
        id: 'section-1',
        title: '1. 栈与堆内存',
        content: '理解栈和堆的区别是掌握所有权的基础。栈用于存储大小固定、生命周期短的数据，而堆用于存储大小动态变化或生命周期较长的数据。',
        code: `// 栈上分配 - 大小在编译时已知
let x = 42;           // i32, 4字节
let y = true;         // bool, 1字节

// 堆上分配 - 动态大小
let s = String::from("hello");  // 堆上分配
let v = vec![1, 2, 3];          // 堆上分配`,
      },
      {
        id: 'section-2',
        title: '2. 所有权规则',
        content: 'Rust 的所有权系统有三条核心规则：每个值都有且仅有一个所有者；当所有者离开作用域，值被丢弃；同一时刻只能有一个可变引用或多个不可变引用。',
        code: `// 所有权转移
let s1 = String::from("hello");
let s2 = s1;  // s1 的所有权移动到 s2
// println!("{}", s1);  // 错误！s1 已无效

// 克隆避免转移
let s3 = s2.clone();  // 深拷贝
println!("s2 = {}, s3 = {}", s2, s3);  // 都有效`,
      },
      {
        id: 'section-3',
        title: '3. 变量作用域',
        content: '变量的作用域决定了它在哪些代码区域有效。当变量离开作用域时，Rust 会自动释放其占用的资源。',
        code: `{
    let s = String::from("hello"); // s 进入作用域
    // 使用 s...
}  // s 离开作用域，资源自动释放`,
      },
      {
        id: 'section-4',
        title: '4. 移动语义',
        content: 'Rust 默认使用移动语义，赋值或传参时所有权会转移。如果需要保留原值，可以使用 clone() 或实现 Copy trait。',
        code: `fn take_ownership(s: String) {
    println!("{}", s);
} // s 离开作用域，内存释放

let my_string = String::from("hello");
take_ownership(my_string);
// my_string 已移动，不能再使用`,
      },
      {
        id: 'section-5',
        title: '5. 引用与借用',
        content: '引用允许你使用值而不获取所有权，这称为"借用"。不可变引用 (&T) 可以有多个，可变引用 (&mut T) 同一时刻只能有一个。',
        code: `fn calculate_length(s: &String) -> usize {
    s.len()  // 借用，不获取所有权
}

let s = String::from("hello");
let len = calculate_length(&s);
println!("'{}' 的长度是 {}", s, len);  // s 仍然有效`,
      },
    ],
    tags: ['Rust', '所有权', '内存管理', '借用', '引用'],
  },
};

export default function CourseDetailPage() {
  const params = useParams();
  const courseId = (params?.id as string) || '1';
  const fallbackCourse = fallbackCourseData[courseId] || fallbackCourseData['1'];
  const [course, setCourse] = useState(fallbackCourse);
  const [activeSection, setActiveSection] = useState(fallbackCourse.sections[0]?.id || '');
  const [readProgress, setReadProgress] = useState(0);
  const [liked, setLiked] = useState(false);
  const [bookmarked, setBookmarked] = useState(false);
  const [comments, setComments] = useState<Array<{ comment_id: string; user: { username: string }; content: string; created_at: string }>>([]);
  const [commentText, setCommentText] = useState('');

  useEffect(() => {
    Promise.all([
      apiFetch<Record<string, unknown>>('/api/v1/courses/' + courseId),
      apiFetch<{ items: Array<{ comment_id: string; user: { username: string }; content: string; created_at: string }> }>('/api/v1/courses/' + courseId + '/comments'),
    ]).then(([detail, commentData]) => {
      setCourse({
        ...fallbackCourse, title: String(detail.title), category: String(detail.category_label || detail.category),
        author: String(detail.author), date: String(detail.publish_date), readTime: String(detail.read_time),
        views: Number(detail.views).toLocaleString(), sections: detail.sections as typeof fallbackCourse.sections,
        tags: detail.tags as string[],
      });
      setComments(commentData.items);
      track('course_view', { course_id: courseId, category: String(detail.category_label || detail.category) });
    }).catch(() => undefined);
  }, [courseId]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (readProgress > 0) apiFetch('/api/v1/progress/courses/' + courseId, {
        method: 'PUT', body: JSON.stringify({ progress: readProgress, last_section_id: activeSection }),
      }).catch(() => undefined);
    }, 800);
    return () => window.clearTimeout(timer);
  }, [readProgress, activeSection, courseId]);

  const submitComment = async () => {
    if (!commentText.trim()) return;
    const created = await apiFetch<{ comment_id: string; content: string; created_at: string }>('/api/v1/courses/' + courseId + '/comments', {
      method: 'POST', body: JSON.stringify({ content: commentText }),
    });
    setComments((current) => [{ ...created, user: { username: '小初' } }, ...current]);
    setCommentText('');
  };

  useEffect(() => {
    const handleScroll = () => {
      const sections = course.sections.map((s) =>
        document.getElementById(s.id)
      ).filter(Boolean) as HTMLElement[];

      const scrollPos = window.scrollY + 100;

      for (let i = sections.length - 1; i >= 0; i--) {
        if (sections[i].offsetTop <= scrollPos) {
          setActiveSection(sections[i].id);
          break;
        }
      }

      // 计算阅读进度
      const docHeight = document.documentElement.scrollHeight - window.innerHeight;
      const progress = Math.min(100, Math.round((scrollPos / docHeight) * 100));
      setReadProgress(progress);
    };

    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, [course.sections]);

  return (
    <>
      {/* 面包屑 */}
      <div className="px-8 pt-6 pb-2">
        <nav className="flex items-center gap-2 text-sm text-muted-foreground">
          <Link href="/" className="hover:text-primary">首页</Link>
          <span>/</span>
          <Link href="/rust" className="hover:text-primary">Rust 编程</Link>
          <span>/</span>
          <span className="text-foreground">第3章：所有权系统</span>
        </nav>
      </div>

      {/* 文章标题 */}
      <div className="px-8 py-6 border-b border-outline/10">
        <h1 className="text-3xl font-bold text-foreground mb-4">
          {course.title}
        </h1>
        <div className="flex items-center gap-4 text-sm text-muted-foreground">
          <span className="px-2 py-0.5 bg-primary/10 text-primary rounded-sm text-xs font-medium">
            {course.category}
          </span>
          <span>作者：{course.author}</span>
          <span>{course.date}</span>
          <span>阅读 {course.readTime}</span>
          <span>{course.views} 次阅读</span>
        </div>
      </div>

      {/* 内容区 */}
      <div className="flex gap-8 px-8 py-8">
        {/* 正文 */}
        <article className="flex-1 min-w-0 max-w-3xl">
          {/* 阅读进度条 */}
          <div className="fixed top-0 left-0 right-0 h-1 bg-surface-container z-50">
            <div
              className="h-full bg-primary transition-all"
              style={{ width: `${readProgress}%` }}
            />
          </div>

          {/* 文章内容 */}
          <div className="space-y-8">
            {course.sections.map((section) => (
              <section key={section.id} id={section.id} className="scroll-mt-20">
                <h2 className="text-xl font-bold text-foreground mb-4 pb-2 border-b border-outline/10">
                  {section.title}
                </h2>
                <p className="text-foreground leading-relaxed mb-4">
                  {section.content}
                </p>
                {section.code && (
                  <div className="relative">
                    <pre className="bg-[#1e1e2e] text-[#cdd6f4] rounded-lg p-4 overflow-x-auto text-sm font-mono leading-relaxed">
                      <code>{section.code}</code>
                    </pre>
                  </div>
                )}
              </section>
            ))}
          </div>

          {/* 互动栏 */}
          <div className="flex items-center gap-4 py-6 mt-8 border-t border-outline/10">
            <button onClick={async () => { const result = await apiFetch<{ liked: boolean }>('/api/v1/courses/' + courseId + '/like', { method: 'POST' }); setLiked(result.liked); }} className="flex items-center gap-2 px-4 py-2 text-sm text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-md transition-colors">
              <Heart className={`w-4 h-4 ${liked ? 'fill-current text-destructive' : ''}`} />
              {liked ? '已点赞' : '点赞'}
            </button>
            <button onClick={async () => { const result = await apiFetch<{ bookmarked: boolean }>('/api/v1/courses/' + courseId + '/bookmark', { method: 'POST' }); setBookmarked(result.bookmarked); }} className="flex items-center gap-2 px-4 py-2 text-sm text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-md transition-colors">
              <Bookmark className={`w-4 h-4 ${bookmarked ? 'fill-current text-primary' : ''}`} />
              {bookmarked ? '已收藏' : '收藏'}
            </button>
            <button className="flex items-center gap-2 px-4 py-2 text-sm text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-md transition-colors">
              <Share2 className="w-4 h-4" />
              分享
            </button>
          </div>

          {/* 上下章导航 */}
          <div className="flex items-center justify-between py-6 border-t border-outline/10">
            <Link
              href="/course/2"
              className="flex items-center gap-2 text-sm text-muted-foreground hover:text-primary transition-colors"
            >
              <ChevronLeft className="w-4 h-4" />
              上一章：变量与类型
            </Link>
            <Link
              href="/course/4"
              className="flex items-center gap-2 text-sm text-muted-foreground hover:text-primary transition-colors"
            >
              下一章：枚举与模式匹配
              <ChevronRight className="w-4 h-4" />
            </Link>
          </div>

          {/* 标签 */}
          <div className="flex flex-wrap gap-2 py-4 border-t border-outline/10">
            {course.tags.map((tag) => (
              <span
                key={tag}
                className="text-xs text-muted-foreground bg-surface-container px-2.5 py-1 rounded-sm"
              >
                {tag}
              </span>
            ))}
          </div>

          {/* 评论区 */}
          <div className="pt-8 border-t border-outline/10">
            <h3 className="text-lg font-semibold text-foreground mb-6">
              评论 ({comments.length})
            </h3>
            <div className="flex gap-3 mb-6">
              <input value={commentText} onChange={(event) => setCommentText(event.target.value)} onKeyDown={(event) => event.key === 'Enter' && submitComment()} placeholder="写下你的评论..." className="flex-1 px-3 py-2 text-sm bg-surface-container-lowest border border-outline/20 rounded-md" />
              <button onClick={submitComment} className="px-4 py-2 text-sm bg-primary text-on-primary rounded-md">发表</button>
            </div>
            <div className="space-y-6">
              {comments.map((comment) => (
                <div key={comment.comment_id} className="flex gap-3">
                  <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium text-primary shrink-0">{comment.user.username[0]}</div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1"><span className="text-sm font-medium text-foreground">{comment.user.username}</span><span className="text-xs text-muted-foreground">{new Date(comment.created_at).toLocaleString()}</span></div>
                    <p className="text-sm text-muted-foreground leading-relaxed">{comment.content}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </article>

        {/* 右侧目录 */}
        <div className="w-64 shrink-0">
          <div className="sticky top-4">
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h4 className="text-sm font-semibold text-foreground mb-4">
                目录
              </h4>
              <nav className="space-y-2">
                {course.sections.map((section) => (
                  <a
                    key={section.id}
                    href={`#${section.id}`}
                    className={`block text-sm py-1 transition-colors ${
                      activeSection === section.id
                        ? 'text-primary font-medium'
                        : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    {section.title}
                  </a>
                ))}
              </nav>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
