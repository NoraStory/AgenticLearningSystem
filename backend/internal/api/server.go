package api

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"codeforge/backend/internal/config"
	"codeforge/backend/internal/database"
	"codeforge/backend/internal/llm"
	"codeforge/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Server struct {
	cfg      config.Config
	services *database.Services
	store    *storage.Store
	llm      *llm.Client
	router   *gin.Engine
}
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func New(cfg config.Config, services *database.Services, store *storage.Store, llmClient *llm.Client) *Server {
	s := &Server{cfg: cfg, services: services, store: store, llm: llmClient}
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(gin.Recovery(), s.logger(), s.cors())
	r.GET("/health", s.health)
	api := r.Group("/api/v1")
	api.Use(s.optionalAuth())
	api.POST("/auth/register", s.register)
	api.POST("/auth/login", s.login)
	api.POST("/auth/refresh", s.refresh)
	api.GET("/users/me", s.me)
	api.PUT("/users/me", s.updateMe)
	api.POST("/users/me/avatar", s.uploadAvatar)
	api.GET("/users/me/streak", s.streak)
	api.GET("/users/me/progress", s.progress)
	api.GET("/users/me/activities", s.activities)
	api.GET("/users/me/achievements", s.achievements)
	api.GET("/users/me/favorites", s.favorites)
	api.GET("/users/me/notes", s.notes)
	api.GET("/courses", s.listCourses)
	api.GET("/courses/recommended", s.recommendedCourses)
	api.GET("/courses/resources", s.courseResources)
	api.GET("/courses/tags", s.courseTags)
	api.GET("/tags/hot", s.hotTags)
	api.GET("/courses/:id", s.courseDetail)
	api.GET("/courses/:id/comments", s.comments)
	api.POST("/courses/:id/comments", s.createComment)
	api.POST("/courses/:id/like", s.toggleLike)
	api.POST("/courses/:id/bookmark", s.toggleBookmark)
	api.GET("/problems", s.listProblems)
	api.GET("/problems/daily", s.dailyProblem)
	api.GET("/problems/:id", s.problemDetail)
	api.GET("/problems/:id/templates", s.problemTemplates)
	api.POST("/code/run", s.runCode)
	api.POST("/code/submit", s.submitCode)
	api.GET("/progress", s.progress)
	api.PUT("/progress/courses/:id", s.updateCourseProgress)
	api.POST("/progress/time", s.recordTime)
	api.GET("/learning-paths", s.learningPaths)
	api.GET("/learning-paths/:id/stages", s.learningPathStages)
	api.POST("/favorites", s.addFavorite)
	api.DELETE("/favorites/:course_id", s.deleteFavorite)
	api.GET("/favorites", s.favorites)
	api.POST("/notes", s.createNote)
	api.GET("/notes", s.notes)
	api.POST("/agent/chat", s.agentChat)
	api.GET("/agent/history", s.agentHistory)
	api.GET("/agent/sessions", s.agentSessions)
	api.DELETE("/agent/sessions/:id", s.deleteAgentSession)
	api.POST("/agent/chat/upload", s.agentUpload)
	api.GET("/agent/tools", s.agentTools)
	api.PATCH("/agent/tools/:id", s.patchAgentTool)
	api.GET("/agent/profile", s.agentProfile)
	api.GET("/agent/knowledge", s.agentKnowledge)
	api.POST("/agent/workflow/confirm", s.confirmWorkflow)
	api.GET("/agent/workflow/:id", s.workflowStatus)
	api.GET("/resume/templates", s.resumeTemplates)
	api.POST("/resume/upload", s.resumeUpload)
	api.POST("/resume/analyze", s.resumeAnalyze)
	api.POST("/resume/optimize", s.resumeOptimize)
	api.POST("/resume/export", s.resumeExport)
	api.GET("/projects", s.listProjects)
	api.POST("/projects", s.createProject)
	api.POST("/projects/upload", s.projectUpload)
	api.POST("/projects/analyze", s.projectAnalyze)
	api.GET("/interview/exams", s.listExams)
	api.POST("/interview/exams/generate", s.generateExam)
	api.GET("/interview/exams/:id", s.examDetail)
	api.POST("/interview/exams/:id/questions/:question_id/run", s.runExamQuestion)
	api.POST("/interview/exams/:id/submit", s.submitExam)
	api.GET("/search", s.search)
	s.router = r
	return s
}
func (s *Server) Router() *gin.Engine { return s.router }
func success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: 200, Message: "success", Data: data})
}
func fail(c *gin.Context, status, code int, message string) {
	c.JSON(status, Response{Code: code, Message: message, Data: nil})
}
func userID(c *gin.Context) string {
	if v, ok := c.Get("user_id"); ok {
		return v.(string)
	}
	return database.DemoUserID
}

func (s *Server) cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := ""
		for _, v := range s.cfg.CORSOrigins {
			if strings.TrimSpace(v) == origin || strings.TrimSpace(v) == "*" {
				allowed = origin
				break
			}
		}
		if allowed != "" {
			c.Header("Access-Control-Allow-Origin", allowed)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Disposition")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
func (s *Server) logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(p gin.LogFormatterParams) string {
		return p.TimeStamp.Format(time.RFC3339) + " level=info module=HTTP trace_id=" + p.Request.Header.Get("X-Request-ID") + " method=" + p.Method + " path=" + p.Path + " status=" + fmtInt(p.StatusCode) + " latency=" + p.Latency.String() + "\n"
	})
}
func fmtInt(v int) string {
	const digits = "0123456789"
	if v == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for v > 0 {
		buf = append([]byte{digits[v%10]}, buf...)
		v /= 10
	}
	return string(buf)
}
func (s *Server) optionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		if token != "" {
			claims := jwt.MapClaims{}
			parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) { return []byte(s.cfg.JWTSecret), nil })
			if err == nil && parsed.Valid {
				if sub, ok := claims["sub"].(string); ok {
					c.Set("user_id", sub)
				}
			}
		}
		if _, ok := c.Get("user_id"); !ok {
			c.Set("user_id", database.DemoUserID)
		}
		c.Next()
	}
}
func (s *Server) signToken(uid string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{"sub": uid, "exp": time.Now().Add(ttl).Unix(), "iat": time.Now().Unix()}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
}
func tokenHash(v string) string { sum := sha256.Sum256([]byte(v)); return hex.EncodeToString(sum[:]) }
func (s *Server) health(c *gin.Context) {
	sqlDB, err := s.services.DB.DB()
	dbOK := err == nil && sqlDB.Ping() == nil
	redisOK := s.services.Redis != nil
	status := http.StatusOK
	if !dbOK {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, gin.H{"status": map[bool]string{true: "ok", false: "degraded"}[dbOK], "database": dbOK, "redis": redisOK, "minio": "configured", "time": time.Now().Format(time.RFC3339)})
}
