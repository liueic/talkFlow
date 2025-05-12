package config

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var DB *sql.DB // Global variable to hold the database connection

func InitSQLite() {
	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = "./db/talkflow.db" // 默认路径
	}

	// 自动创建数据库目录
	dir := ""
	if idx := strings.LastIndex(dbPath, "/"); idx != -1 {
		dir = dbPath[:idx]
	}
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("无法创建数据库目录: %v", err)
		}
	}

	// 建立数据库连接
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("无法连接到SQLite数据库：%v", err)
	}

	// 对连接池进行配置
	db.SetMaxOpenConns(1)                  // SQLite 不支持并发写，多连接可能导致锁冲突，所以一般设置为1
	db.SetMaxIdleConns(1)                  // 同上
	db.SetConnMaxLifetime(0 * time.Second) // 永不过期

	// Ping 一下数据库，确保连接可用
	if err = db.Ping(); err != nil {
		log.Fatalf("无法 ping SQLite 数据库：%v", err)
	}

	log.Println("成功连接到 SQLite 数据库.")

	// Assign the database connection to the global variable
	DB = db

	// Handle graceful shutdown to close the database connection
	handleShutdown(db)
}

func InitTables() {
	createRegisterTable := `
    CREATE TABLE IF NOT EXISTS register (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE,
        password TEXT,
        email TEXT,
        avatar TEXT,
        created_at DATETIME,
        register_ip TEXT,
        is_register BOOLEAN,
        last_login_ip TEXT,
        last_login_time DATETIME
    );`
	createVisitorTable := `
	CREATE TABLE IF NOT EXISTS visitor (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT,
    created_at DATETIME,
    visitor_ip TEXT,
    is_register BOOLEAN
);`
	createLogTable := `
    CREATE TABLE IF NOT EXISTS log (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT,
        error TEXT,
        timestamp TEXT,
        ip TEXT
    );`
	_, err := DB.Exec(createRegisterTable)
	if err != nil {
		log.Fatalf("创建 register 表失败: %v", err)
	}

	_, err = DB.Exec(createVisitorTable)
	if err != nil {
		log.Fatalf("创建 visitor 表失败: %v", err)
	}

	_, err = DB.Exec(createLogTable)
	if err != nil {
		log.Fatalf("创建 log 表失败: %v", err)
	}
}

func handleShutdown(db *sql.DB) {
	// Create a channel to listen for OS signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("正在关闭 SQLite 数据库连接...")
		if err := db.Close(); err != nil {
			log.Printf("关闭 SQLite 数据库连接时出错：%v", err)
		} else {
			log.Println("SQLite 数据库连接已关闭.")
		}
		os.Exit(0)
	}()
}
