package database

import (
	"fmt"
	"time"

	"codeforge/backend/internal/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const DemoUserID = "00000000-0000-0000-0000-000000000001"

func Seed(db *gorm.DB) error {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("Demo123!"), bcrypt.DefaultCost)
		u := model.User{ID: DemoUserID, Username: "小初", Email: "demo@codeforge.local", PasswordHash: string(hash), Bio: "正在系统学习全栈开发与 AI Agent。", Level: 8, LevelTitle: "中级学习者", LearningDays: 15}
		if err := db.Create(&u).Error; err != nil {
			return err
		}
	}
	if err := seedCourses(db); err != nil {
		return err
	}
	if err := seedProblems(db); err != nil {
		return err
	}
	if err := seedPaths(db); err != nil {
		return err
	}
	if err := seedTemplates(db); err != nil {
		return err
	}
	return seedUserData(db)
}

func seedCourses(db *gorm.DB) error {
	var count int64
	db.Model(&model.Course{}).Count(&count)
	if count > 0 {
		return nil
	}
	categories := []struct {
		key, label string
		titles     []string
		image      string
	}{
		{"python", "Python", []string{"Python 基础语法与类型系统", "函数、模块与包管理", "面向对象编程实战", "异步编程与 Web API"}, "https://images.unsplash.com/photo-1526379095098-d400fd0bf935?w=800&h=400&fit=crop"},
		{"cpp", "C++", []string{"C++ 现代语法入门", "STL 容器与算法", "内存模型与智能指针", "并发与高性能编程"}, "https://images.unsplash.com/photo-1555066931-4365d14bab8c?w=800&h=400&fit=crop"},
		{"database", "数据库", []string{"SQL 查询基础", "索引与查询优化", "事务与并发控制", "Redis 与缓存设计"}, "https://images.unsplash.com/photo-1544383835-bda2bc66a55d?w=800&h=400&fit=crop"},
		{"algorithm", "算法", []string{"数组与双指针", "树与图的遍历", "动态规划方法论", "贪心与回溯实战"}, "https://images.unsplash.com/photo-1509228468518-180dd4864904?w=800&h=400&fit=crop"},
		{"agent", "Agent", []string{"AI Agent 核心架构", "Prompt 与结构化输出", "RAG 知识库实战", "多 Agent 工作流编排"}, "https://images.unsplash.com/photo-1677442136019-21780ecad995?w=800&h=400&fit=crop"},
	}
	levels := []string{"入门", "入门", "进阶", "高级"}
	statuses := []string{"completed", "in_progress", "not_started", "not_started"}
	id := uint(1)
	for _, cat := range categories {
		for i, title := range cat.titles {
			sections := []model.CourseSection{
				{ID: "section-1", Title: "1. 核心概念", Content: fmt.Sprintf("本节系统介绍《%s》的核心概念、适用场景和学习目标。通过可运行示例建立完整认知。", title), Code: sampleCode(cat.key), Language: cat.key},
				{ID: "section-2", Title: "2. 实践示例", Content: "从一个最小示例开始，逐步增加边界条件、错误处理和工程化约束。建议边阅读边在练习区运行代码。", Code: sampleCode(cat.key), Language: cat.key},
				{ID: "section-3", Title: "3. 常见问题", Content: "总结学习过程中最容易混淆的概念，并给出排查思路、测试方法和延伸练习。"},
			}
			c := model.Course{ID: id, Title: title, Category: cat.key, CategoryLabel: cat.label, Difficulty: levels[i], Level: levels[i], Status: statuses[i], CoverImage: cat.image, Summary: "面向实践的系统课程，包含概念讲解、示例代码、练习与工程建议。", Description: "循序渐进掌握核心知识，并通过项目式练习形成可迁移的解决问题能力。", Author: "CodeForge 教研组", PublishDate: fmt.Sprintf("2026-07-%02d", 29-i), ReadTime: fmt.Sprintf("%d 分钟", 12+i*4), Views: 1200 + i*487, Tags: []string{cat.label, levels[i], "实战"}, Sections: sections, LessonsCount: 8 + i*2, EstimatedHours: 4 + i*3, Progress: []int{100, 45, 0, 0}[i]}
			if err := db.Create(&c).Error; err != nil {
				return err
			}
			id++
		}
	}
	return nil
}

func sampleCode(category string) string {
	switch category {
	case "python":
		return "def greet(name: str) -> str:\n    return f\"Hello, {name}!\"\n\nprint(greet(\"CodeForge\"))"
	case "cpp":
		return "#include <iostream>\nint main() { std::cout << \"Hello CodeForge\"; }"
	case "database":
		return "SELECT category, COUNT(*)\nFROM courses\nGROUP BY category\nORDER BY COUNT(*) DESC;"
	case "algorithm":
		return "def two_sum(nums, target):\n    seen = {}\n    for i, value in enumerate(nums):\n        if target - value in seen:\n            return [seen[target - value], i]\n        seen[value] = i"
	default:
		return "const agent = { plan, act, observe };"
	}
}

func seedProblems(db *gorm.DB) error {
	var count int64
	db.Model(&model.Problem{}).Count(&count)
	if count > 0 {
		return nil
	}
	defs := []struct{ title, cat, diff, desc string }{
		{"两数之和", "数组", "简单", "给定整数数组和目标值，返回和为目标值的两个元素下标。"},
		{"反转链表", "链表", "简单", "反转一个单链表并返回新的头节点。"},
		{"有效的括号", "栈", "简单", "判断括号字符串是否有效。"},
		{"二叉树层序遍历", "树", "中等", "按层返回二叉树节点值。"},
		{"最长递增子序列", "动态规划", "中等", "求数组中最长严格递增子序列长度。"},
		{"合并区间", "排序", "中等", "合并所有重叠区间。"},
		{"最小覆盖子串", "滑动窗口", "困难", "返回覆盖目标字符的最短子串。"},
		{"单词接龙", "图", "困难", "求从起始单词到目标单词的最短转换序列长度。"},
	}
	for i, d := range defs {
		id := uint(i + 1)
		p := model.Problem{ID: id, Title: d.title, Category: d.cat, Difficulty: d.diff, Description: d.desc, Examples: []model.Example{{Input: "[2,7,11,15], 9", Output: "[0,1]", Explanation: "2 + 7 = 9"}}, Constraints: []string{"输入规模在合理范围内", "请处理空输入与边界情况"}, Tags: []string{d.cat, d.diff}, Templates: map[string]string{"python": "class Solution:\n    def solve(self, nums, target):\n        pass\n", "javascript": "class Solution {\n  solve(nums, target) {\n  }\n}\n", "cpp": "class Solution {\npublic:\n    vector<int> solve(vector<int>& nums, int target) {\n    }\n};\n", "rust": "impl Solution {\n    pub fn solve(nums: Vec<i32>, target: i32) -> Vec<i32> {\n        vec![]\n    }\n}"}, TestCases: []model.TestCase{{Input: "", Expected: ""}}, PassRate: 65.8 - float64(i)*3.1, Submissions: 2530 + i*713, Status: []string{"solved", "solved", "attempted", "not_started", "not_started", "not_started", "not_started", "not_started"}[i]}
		if err := db.Create(&p).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedPaths(db *gorm.DB) error {
	var count int64
	db.Model(&model.LearningPath{}).Count(&count)
	if count > 0 {
		return nil
	}
	paths := []model.LearningPath{{ID: "python", Name: "Python 全栈工程师", Description: "从语法基础到 Web、数据与工程实践", Courses: 12, Levels: 4, Icon: "Code", Color: "blue"}, {ID: "cpp", Name: "C++ 高性能开发", Description: "现代 C++、系统编程与性能优化", Courses: 10, Levels: 4, Icon: "Cpu", Color: "purple"}, {ID: "database", Name: "数据库工程师", Description: "SQL、事务、索引与分布式数据", Courses: 9, Levels: 4, Icon: "Database", Color: "cyan"}, {ID: "agent", Name: "AI Agent 开发者", Description: "LLM、RAG、Tool 与多 Agent 编排", Courses: 11, Levels: 4, Icon: "Sparkles", Color: "rose"}}
	for _, p := range paths {
		if err := db.Create(&p).Error; err != nil {
			return err
		}
		for i := 1; i <= 4; i++ {
			s := model.LearningPathStage{ID: fmt.Sprintf("%s-%d", p.ID, i), PathID: p.ID, Name: []string{"基础入门", "核心进阶", "项目实战", "工程深化"}[i-1], Status: []string{"completed", "in_progress", "locked", "locked"}[i-1], Hours: 10 + i*8, Goal: "完成课程、练习和阶段项目", Courses: []string{fmt.Sprintf("阶段 %d 核心课", i), fmt.Sprintf("阶段 %d 实战课", i)}, Prerequisite: func() string {
				if i == 1 {
					return "无"
				}
				return fmt.Sprintf("完成阶段 %d", i-1)
			}(), SortOrder: i}
			if err := db.Create(&s).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedTemplates(db *gorm.DB) error {
	var count int64
	db.Model(&model.ResumeTemplate{}).Count(&count)
	if count > 0 {
		return nil
	}
	items := []model.ResumeTemplate{{ID: "tech-standard", Name: "技术岗位标准版", Category: "tech", Description: "适合程序员、工程师等技术岗位", Preview: "💻", Sections: []string{"基本信息", "技能清单", "工作经历", "项目经验", "教育背景"}, Style: "简洁专业，突出技术栈和项目成果"}, {ID: "tech-modern", Name: "技术岗位现代版", Category: "tech", Description: "现代化设计，适合互联网行业", Preview: "🚀", Sections: []string{"基本信息", "技术栈", "工作经历", "项目亮点", "开源贡献"}, Style: "现代简约，强调技术深度和影响力"}, {ID: "product", Name: "产品经理版", Category: "product", Description: "适合产品经理、运营等岗位", Preview: "📊", Sections: []string{"基本信息", "个人简介", "工作经历", "项目经验", "数据成果"}, Style: "数据驱动，突出产品思维"}, {ID: "design", Name: "设计师版", Category: "design", Description: "适合 UI/UX 设计师", Preview: "🎨", Sections: []string{"基本信息", "设计技能", "作品集", "工作经历"}, Style: "视觉优先，展示设计作品"}, {ID: "general", Name: "通用版", Category: "general", Description: "适合大多数岗位", Preview: "📋", Sections: []string{"基本信息", "个人简介", "工作经历", "教育背景", "技能证书"}, Style: "经典布局，信息均衡"}, {ID: "fresh-graduate", Name: "应届生版", Category: "fresh", Description: "适合应届毕业生", Preview: "🌟", Sections: []string{"基本信息", "教育背景", "实习经历", "项目经验", "校园经历"}, Style: "突出学习能力和潜力"}}
	return db.Create(&items).Error
}

func seedUserData(db *gorm.DB) error {
	var p model.UserProfile
	if err := db.Where("user_id = ?", DemoUserID).First(&p).Error; err == gorm.ErrRecordNotFound {
		p = model.UserProfile{ID: uuid.NewString(), UserID: DemoUserID, Level: "中级开发者", FocusAreas: []string{"Python", "数据结构", "AI Agent"}, WeakAreas: []string{"并发编程", "系统设计"}, LearningStyle: "实践型", PreferredDifficulty: "中等", DailyGoal: 60, TotalStudyTime: 256, Streak: 15}
		if err := db.Create(&p).Error; err != nil {
			return err
		}
	}
	var k model.UserKnowledgeGraph
	if err := db.Where("user_id = ?", DemoUserID).First(&k).Error; err == gorm.ErrRecordNotFound {
		k = model.UserKnowledgeGraph{ID: uuid.NewString(), UserID: DemoUserID, Areas: []model.KnowledgeArea{{Name: "Python 基础", Level: 85, Color: "bg-blue-500"}, {Name: "Python 进阶", Level: 62, Color: "bg-blue-600"}, {Name: "C++ 基础", Level: 45, Color: "bg-purple-500"}, {Name: "数据结构", Level: 58, Color: "bg-green-500"}, {Name: "算法", Level: 35, Color: "bg-orange-500"}, {Name: "数据库", Level: 40, Color: "bg-cyan-500"}, {Name: "AI Agent", Level: 28, Color: "bg-rose-500"}}, RecentTopics: []model.Topic{{Name: "Python 装饰器", Connections: 5, Mastery: 78}, {Name: "二叉树遍历", Connections: 3, Mastery: 65}, {Name: "SQL 连接", Connections: 4, Mastery: 72}, {Name: "LangChain Chain", Connections: 2, Mastery: 45}, {Name: "动态规划", Connections: 6, Mastery: 38}}}
		if err := db.Create(&k).Error; err != nil {
			return err
		}
	}
	achievements := []model.Achievement{{ID: "first-course", Name: "初学乍练", Description: "完成第一门课程", Icon: "BookOpen"}, {ID: "streak-7", Name: "坚持不懈", Description: "连续学习 7 天", Icon: "Flame"}, {ID: "problem-10", Name: "算法新星", Description: "解决 10 道算法题", Icon: "Trophy"}, {ID: "agent-user", Name: "智能协作", Description: "首次使用 Agent 工作流", Icon: "Sparkles"}}
	for _, a := range achievements {
		db.FirstOrCreate(&a, model.Achievement{ID: a.ID})
	}
	var ac int64
	db.Model(&model.UserActivity{}).Where("user_id = ?", DemoUserID).Count(&ac)
	if ac == 0 {
		now := time.Now()
		acts := []model.UserActivity{{ID: uuid.NewString(), UserID: DemoUserID, Type: "course", Text: "完成了《Python 基础语法与类型系统》", CreatedAt: now.Add(-2 * time.Hour)}, {ID: uuid.NewString(), UserID: DemoUserID, Type: "problem", Text: "解决了算法题《两数之和》", CreatedAt: now.Add(-26 * time.Hour)}, {ID: uuid.NewString(), UserID: DemoUserID, Type: "agent", Text: "使用学习助手生成了一份复习计划", CreatedAt: now.Add(-48 * time.Hour)}}
		db.Create(&acts)
	}
	return nil
}
