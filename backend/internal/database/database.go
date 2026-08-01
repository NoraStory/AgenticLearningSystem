package database

import (
	"context"
	"fmt"
	"time"

	"codeforge/backend/internal/config"
	"codeforge/backend/internal/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Services struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func Connect(cfg config.Config) (*Services, error) {
	var db *gorm.DB
	var err error
	for i := 0; i < 20; i++ {
		db, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
		if err == nil {
			break
		}
		time.Sleep(1500 * time.Millisecond)
	}
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := dropDeadTables(db); err != nil {
		return nil, fmt.Errorf("drop dead tables: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: 0, PoolSize: 10})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb = nil
	}
	return &Services{DB: db, Redis: rdb}, nil
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{}, &model.RefreshToken{}, &model.Course{}, &model.Comment{}, &model.CourseLike{}, &model.Favorite{}, &model.Note{},
		&model.Problem{}, &model.Submission{}, &model.LearningProgress{}, &model.DailyStudyTime{}, &model.LearningPath{}, &model.LearningPathStage{},
		&model.ResumeTemplate{}, &model.Resume{}, &model.Project{}, &model.ProjectTask{}, &model.InterviewExam{}, &model.SessionMessage{},
		&model.UserToolSetting{}, &model.UserProfile{}, &model.WorkflowExecution{}, &model.UserActivity{}, &model.Achievement{}, &model.UserAchievement{}, &model.KnowledgeState{}, &model.BktParam{},
	)
}

// dropDeadTables 显式删除已从模型移除的旧表（AutoMigrate 只建不删）。
// user_knowledge_graphs 已删除（死代码：仅 Demo 用户种子数据，运行时零写入）。
func dropDeadTables(db *gorm.DB) error {
	if db.Migrator().HasTable("user_knowledge_graphs") {
		return db.Migrator().DropTable("user_knowledge_graphs")
	}
	return nil
}
