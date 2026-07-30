package model

import "time"

type User struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	Username     string    `gorm:"size:80;uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"size:160;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Avatar       string    `json:"avatar"`
	Bio          string    `json:"bio"`
	Level        int       `gorm:"default:8" json:"level"`
	LevelTitle   string    `gorm:"default:中级学习者" json:"level_title"`
	LearningDays int       `gorm:"default:15" json:"learning_days"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RefreshToken struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	UserID    string `gorm:"type:uuid;index;not null"`
	TokenHash string `gorm:"size:128;uniqueIndex;not null"`
	ExpiresAt time.Time
	CreatedAt time.Time
}

type CourseSection struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Code     string `json:"code,omitempty"`
	Language string `json:"language,omitempty"`
}
type Course struct {
	ID             uint            `gorm:"primaryKey" json:"course_id"`
	Title          string          `gorm:"size:240;index;not null" json:"title"`
	Category       string          `gorm:"size:40;index" json:"category"`
	CategoryLabel  string          `gorm:"size:40" json:"category_label"`
	Difficulty     string          `gorm:"size:30" json:"difficulty"`
	Level          string          `gorm:"size:30" json:"level"`
	Status         string          `gorm:"size:30;default:not_started" json:"status"`
	CoverImage     string          `gorm:"size:500" json:"cover_image"`
	Summary        string          `gorm:"type:text" json:"summary"`
	Description    string          `gorm:"type:text" json:"description"`
	Author         string          `gorm:"size:100" json:"author"`
	PublishDate    string          `gorm:"size:50" json:"publish_date"`
	ReadTime       string          `gorm:"size:30" json:"read_time"`
	Views          int             `gorm:"default:0" json:"views"`
	Tags           []string        `gorm:"serializer:json;type:jsonb" json:"tags"`
	Sections       []CourseSection `gorm:"serializer:json;type:jsonb" json:"sections"`
	LessonsCount   int             `json:"lessons_count"`
	EstimatedHours int             `json:"estimated_hours"`
	Progress       int             `gorm:"default:0" json:"progress"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type Comment struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"comment_id"`
	CourseID  uint      `gorm:"index" json:"course_id"`
	UserID    string    `gorm:"type:uuid;index" json:"user_id"`
	Content   string    `gorm:"type:text" json:"content"`
	Likes     int       `json:"likes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type CourseLike struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	CourseID  uint   `gorm:"uniqueIndex:idx_like"`
	UserID    string `gorm:"type:uuid;uniqueIndex:idx_like"`
	CreatedAt time.Time
}
type Favorite struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	CourseID  uint      `gorm:"uniqueIndex:idx_fav" json:"course_id"`
	UserID    string    `gorm:"type:uuid;uniqueIndex:idx_fav" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
type Note struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	CourseID  uint      `gorm:"index" json:"course_id"`
	UserID    string    `gorm:"type:uuid;index" json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Example struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	Explanation string `json:"explanation,omitempty"`
}
type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}
type Problem struct {
	ID          uint              `gorm:"primaryKey" json:"problem_id"`
	Title       string            `gorm:"size:240;index" json:"title"`
	Category    string            `gorm:"size:60;index" json:"category"`
	Difficulty  string            `gorm:"size:30;index" json:"difficulty"`
	Description string            `gorm:"type:text" json:"description"`
	Examples    []Example         `gorm:"serializer:json;type:jsonb" json:"examples"`
	Constraints []string          `gorm:"serializer:json;type:jsonb" json:"constraints"`
	Tags        []string          `gorm:"serializer:json;type:jsonb" json:"tags"`
	Templates   map[string]string `gorm:"serializer:json;type:jsonb" json:"templates"`
	TestCases   []TestCase        `gorm:"serializer:json;type:jsonb" json:"-"`
	PassRate    float64           `json:"pass_rate"`
	Submissions int               `json:"submissions"`
	Status      string            `gorm:"size:30" json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type Submission struct {
	ID              string `gorm:"type:uuid;primaryKey" json:"submission_id"`
	UserID          string `gorm:"type:uuid;index"`
	ProblemID       uint   `gorm:"index"`
	Language        string
	Code            string `gorm:"type:text"`
	Status          string
	PassedCases     int
	TotalCases      int
	ExecutionTimeMS int64
	MemoryKB        int64
	CreatedAt       time.Time
}
type LearningProgress struct {
	ID            string `gorm:"type:uuid;primaryKey"`
	UserID        string `gorm:"type:uuid;uniqueIndex:idx_progress"`
	CourseID      uint   `gorm:"uniqueIndex:idx_progress"`
	Progress      int
	LastSectionID string
	UpdatedAt     time.Time
	CreatedAt     time.Time
}
type DailyStudyTime struct {
	ID              string `gorm:"type:uuid;primaryKey"`
	UserID          string `gorm:"type:uuid;index"`
	CourseID        uint
	StudyDate       time.Time `gorm:"type:date;index"`
	DurationMinutes int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type LearningPath struct {
	ID          string `gorm:"size:40;primaryKey" json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Courses     int    `json:"courses"`
	Levels      int    `json:"levels"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	CreatedAt   time.Time
}
type LearningPathStage struct {
	ID           string   `gorm:"size:50;primaryKey" json:"id"`
	PathID       string   `gorm:"size:40;index" json:"path_id"`
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	Hours        int      `json:"hours"`
	Goal         string   `json:"goal"`
	Courses      []string `gorm:"serializer:json;type:jsonb" json:"courses"`
	Prerequisite string   `json:"prerequisite"`
	SortOrder    int      `json:"sort_order"`
}

type ResumeTemplate struct {
	ID          string   `gorm:"size:60;primaryKey" json:"id"`
	Name        string   `json:"name"`
	Category    string   `gorm:"index" json:"category"`
	Description string   `json:"description"`
	Preview     string   `json:"preview"`
	Sections    []string `gorm:"serializer:json;type:jsonb" json:"sections"`
	Style       string   `json:"style"`
	CreatedAt   time.Time
}
type Resume struct {
	ID               string `gorm:"type:uuid;primaryKey"`
	UserID           string `gorm:"type:uuid;index"`
	Filename         string
	ObjectKey        string
	FileSize         int64
	MimeType         string
	AnalysisJSON     string `gorm:"type:jsonb"`
	OptimizedContent string `gorm:"type:text"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Project struct {
	ID          string        `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      string        `gorm:"type:uuid;index" json:"-"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	TechStack   []string      `gorm:"serializer:json;type:jsonb" json:"techStack"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
	Tasks       []ProjectTask `gorm:"foreignKey:ProjectID" json:"tasks"`
}
type ProjectTask struct {
	ID          string   `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID   string   `gorm:"type:uuid;index" json:"-"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Files       []string `gorm:"serializer:json;type:jsonb" json:"files"`
	Analysis    string   `gorm:"type:text" json:"analysis,omitempty"`
	CreatedAt   time.Time
}

type InterviewExam struct {
	ID         string              `gorm:"type:uuid;primaryKey" json:"exam_id"`
	UserID     string              `gorm:"type:uuid;index"`
	Direction  string              `json:"direction"`
	Difficulty string              `json:"difficulty"`
	Questions  []InterviewQuestion `gorm:"serializer:json;type:jsonb" json:"questions"`
	Score      int                 `json:"score"`
	CreatedAt  time.Time           `json:"created_at"`
}
type InterviewQuestion struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Category    string   `json:"category"`
	Difficulty  string   `json:"difficulty"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Constraints []string `json:"constraints"`
	TimeLimit   int      `json:"time_limit"`
	Score       int      `json:"score"`
	Example     *Example `json:"example,omitempty"`
}

type SessionMessage struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       string    `gorm:"type:uuid;index"`
	SessionID    string    `gorm:"type:uuid;index" json:"session_id"`
	Role         string    `json:"role"`
	Agent        string    `json:"agent"`
	Content      string    `gorm:"type:text" json:"content"`
	WorkflowJSON string    `gorm:"type:jsonb" json:"workflow,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
type UserToolSetting struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	UserID    string `gorm:"type:uuid;uniqueIndex:idx_tool"`
	ToolID    string `gorm:"size:60;uniqueIndex:idx_tool"`
	Enabled   bool
	UpdatedAt time.Time
	CreatedAt time.Time
}
type UserProfile struct {
	ID                  string `gorm:"type:uuid;primaryKey"`
	UserID              string `gorm:"type:uuid;uniqueIndex"`
	Level               string
	FocusAreas          []string `gorm:"serializer:json;type:jsonb"`
	WeakAreas           []string `gorm:"serializer:json;type:jsonb"`
	LearningStyle       string
	PreferredDifficulty string
	DailyGoal           int
	TotalStudyTime      int
	Streak              int
	PreferredTimeSlot   string
	SessionCount        int
	ProblemSolvedCount  int
	ProblemAccuracy     float64
	LastActiveAt        time.Time
	UpdatedAt           time.Time
	CreatedAt           time.Time
}
type KnowledgeArea struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
	Color string `json:"color"`
}
type Topic struct {
	Name        string `json:"name"`
	Connections int    `json:"connections"`
	Mastery     int    `json:"mastery"`
}
type UserKnowledgeGraph struct {
	ID           string          `gorm:"type:uuid;primaryKey"`
	UserID       string          `gorm:"type:uuid;uniqueIndex"`
	Areas        []KnowledgeArea `gorm:"serializer:json;type:jsonb"`
	RecentTopics []Topic         `gorm:"serializer:json;type:jsonb"`
	UpdatedAt    time.Time
	CreatedAt    time.Time
}
type KnowledgeState struct {
	ID              string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          string    `gorm:"type:uuid;uniqueIndex:idx_kstate" json:"user_id"`
	SkillName       string    `gorm:"size:120;uniqueIndex:idx_kstate" json:"skill_name"`
	Category        string    `json:"category"`
	Mastery         float64   `json:"mastery"`
	Attempts        int       `json:"attempts"`
	CorrectCount   int       `json:"correct_count"`
	LastPracticedAt time.Time `json:"last_practiced_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
type WorkflowExecution struct {
	ID          string `gorm:"type:uuid;primaryKey" json:"workflow_id"`
	UserID      string `gorm:"type:uuid;index"`
	Status      string `json:"status"`
	CurrentNode string `json:"current_node"`
	InputJSON   string `gorm:"type:jsonb"`
	ResultJSON  string `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type UserActivity struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	UserID    string    `gorm:"type:uuid;index"`
	Type      string    `json:"type"`
	Text      string    `json:"text"`
	Verb      string    `json:"verb,omitempty"`
	Object    string    `json:"object,omitempty"`
	Metadata  string    `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
type Achievement struct {
	ID          string `gorm:"size:60;primaryKey" json:"id"`
	Name        string `json:"name"`
	Description string `json:"desc"`
	Icon        string `json:"icon"`
	CreatedAt   time.Time
}
type UserAchievement struct {
	ID            string `gorm:"type:uuid;primaryKey"`
	UserID        string `gorm:"type:uuid;uniqueIndex:idx_ach"`
	AchievementID string `gorm:"size:60;uniqueIndex:idx_ach"`
	UnlockedAt    *time.Time
}
