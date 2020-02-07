package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gopkg.in/gormigrate.v1"
)

type SqliteDatabase struct {
	database *gorm.DB
}

func newSqliteDatabase() *SqliteDatabase {
	instance = &SqliteDatabase{}
	return instance
}

var instance *SqliteDatabase
var once sync.Once

func GetSqliteDatabase() *SqliteDatabase {
	once.Do(func() {
		instance = newSqliteDatabase()
	})
	return instance
}

func (s *SqliteDatabase) Init(path string) *SqliteDatabase {
	var err error

	s.database, err = gorm.Open("sqlite3", path)
	if err != nil {
		fmt.Println("SqliteDatabase: can't open database", err)
		s.database = nil
		return s
	}

	s.migrate()
	return s
}

func (s *SqliteDatabase) migrate() *SqliteDatabase {
	migrator := gormigrate.New(s.database, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201905301400",
			Migrate: func(database *gorm.DB) error {
				return database.AutoMigrate(&StatusModel{}).Error
			},
			Rollback: func(database *gorm.DB) error {
				return database.DropTableIfExists("status").Error
			},
		},
	})

	err := migrator.Migrate()
	if err != nil {
		fmt.Println("SqliteDatabase: can't migrate database", err)
		s.database = nil
		return s
	}
	return s
}

func (s *SqliteDatabase) GetDatabase() *gorm.DB {
	return s.database
}

func NewStatusModel() *StatusModel {
	return &StatusModel{}
}

type StatusModel struct {
	gorm.Model
	Heartbeat time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (StatusModel) TableName() string {
	return "status"
}

func (s *StatusModel) BeatHeart() *StatusModel {
	tx := GetSqliteDatabase().GetDatabase().Begin()
	defer tx.Commit()

	tx.FirstOrCreate(s)
	tx.Model(s).Update("heartbeat", time.Now())
	return s
}

func (s *StatusModel) Load() *StatusModel {
	GetSqliteDatabase().GetDatabase().FirstOrCreate(s)
	return s
}

func main() {
	fmt.Println("hello")
	GetSqliteDatabase().Init("hello-gorm.db")
	statusModel := NewStatusModel()

	// finished := make(chan bool)
	// go func() {
	for {
		statusModel.BeatHeart()
		fmt.Println("heartbeat")
		time.Sleep(10 * time.Second)
	}
	// finished <- true
	// }()

	// <-finished
	// fmt.Println("gorm")
}
