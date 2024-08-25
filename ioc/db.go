package ioc

import (
	"context"
	"database/sql"
	evaluationv1 "github.com/MuxiKeStack/be-api/gen/proto/evaluation/v1"
	"github.com/MuxiKeStack/be-evaluation/pkg/limiter"
	"github.com/MuxiKeStack/be-evaluation/pkg/logger"
	"github.com/MuxiKeStack/be-evaluation/repository/dao"
	sql2 "github.com/seata/seata-go/pkg/datasource/sql"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"time"
)

func InitDB(l logger.Logger, lm limiter.Limiter) *gorm.DB {
	return InitMysqlDB(l, lm)
}

func InitMysqlDB(l logger.Logger, lm limiter.Limiter) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("mysql", &cfg); err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			SlowThreshold: 0,
			LogLevel:      glogger.Info, // 以Debug模式打印所有Info级别能产生的gorm日志
		}),
	})
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	err = db.Callback().Query().Before("*").Register("RateLimitGormMiddleware", func(d *gorm.DB) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		ok, er := lm.Limit(ctx, "kstack:evaluation:sql-gorm:query")
		if er == nil && ok {
			// 触发限流
			_ = d.AddError(evaluationv1.ErrorGormTooManyRequest("GORM-SQL请求限流"))
			return
		}
		if er != nil {
			l.Error("限流失败", logger.Error(er))
		}
	})
	if err != nil {
		panic(err)
	}
	return db
}

func InitATMysqlDB(l logger.Logger, lm limiter.Limiter) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("mysql", &cfg); err != nil {
		panic(err)
	}
	sqlDB, err := sql.Open(sql2.SeataATMySQLDriver, cfg.DSN)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			SlowThreshold: 0,
			LogLevel:      glogger.Info, // 以Debug模式打印所有Info级别能产生的gorm日志
		}),
	})
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	err = db.Callback().Query().Before("*").Register("RateLimitGormMiddleware", func(d *gorm.DB) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		ok, er := lm.Limit(ctx, "kstack:evaluation:sql-gorm:query")
		if er == nil && ok {
			// 触发限流
			_ = d.AddError(evaluationv1.ErrorGormTooManyRequest("GORM-SQL请求限流"))
			return
		}
		if er != nil {
			l.Error("限流失败", logger.Error(er))
		}
	})
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(s string, i ...interface{}) {
	g(s, logger.Field{Key: "args", Val: i})
}
