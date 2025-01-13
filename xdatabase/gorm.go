package xdatabase

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Path            string `yaml:"path"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Config          string `yaml:"config"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	IsConsole       bool   `yaml:"is_console"`
}

type PostgresConfig struct {
	URI       string `yaml:"uri"`
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	User      string `yaml:"user"`
	DBName    string `yaml:"dbname"`
	Password  string `yaml:"password"`
	SSLMode   string `yaml:"sslmode"`
	TimeZone  string `yaml:"timeZone"`
	IsConsole bool   `yaml:"is_console"`
}

// NewMySQLGormDb 创建MySQL客户端
func NewMySQLGormDb(config *MySQLConfig) (e *gorm.DB, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		config.Username,
		config.Password,
		config.Path,
		config.Database,
		config.Config,
	)

	cfg := mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         255,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}

	defer func() {
		if err != nil {
			return
		}
		db, _ := e.DB()
		db.SetMaxIdleConns(config.MaxIdleConns)
		db.SetMaxOpenConns(config.MaxOpenConns)
		db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)
	}()

	loggerLevel := logger.Silent
	if config.IsConsole {
		loggerLevel = logger.Info
	}
	// 打开数据库连接
	return gorm.Open(mysql.New(cfg), &gorm.Config{
		Logger: logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				// IgnoreRecordNotFoundError: true,
				LogLevel:      loggerLevel, // Log level
				Colorful:      true,        // 使用彩色打印
				SlowThreshold: time.Second, // 慢 SQL 阈值
			},
		),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
}

func NewPostgresGormDb(config *PostgresConfig) (e *gorm.DB, err error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s TimeZone=%s",
		config.Host,
		config.Port,
		config.User,
		config.DBName,
		config.Password,
		config.SSLMode,
		config.TimeZone,
	)

	loggerLevel := logger.Silent
	if config.IsConsole {
		loggerLevel = logger.Info
	}
	return gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel:      loggerLevel,
				Colorful:      true,
				SlowThreshold: time.Second,
			},
		),
	})
}

func NewPostgresGormDbWithDSN(dsn string) (e *gorm.DB, err error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logger.Info),
	},
	)
}
