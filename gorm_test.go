package gorm_test

import (
	"database/sql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

type Product struct {
	Code  string
	Price uint
}

type User struct {
	//ID           int64  `gorm:"primary_key"`
	Name         string `gorm:"column:namexx"`
	Face         int    `gorm:"type:smallint"`
	Gender       string
	Age          sql.NullInt64 `gorm:"default:11"`
	Height       float64 `gorm:"type:double(9,2);precision:3"`
	Birthday     *time.Time
	Email        string `gorm:"type:varchar(100);unique_index"`
	Role         string `gorm:"size:255"`        // 设置字段大小为255
	MemberNumber string `gorm:"unique;not null"` // 设置会员号（member number）唯一并且不为空
	Address      string `gorm:"index:addr"`      // 给address字段创建名为addr的索引
	Pid 		 int    `gorm:"auot"`
	IgnoreMe     int    `gorm:"-"`               // 忽略本字段

	Company      struct{
		Name string
		Number int
}						`gorm:"EMBEDDED"`
}

var (
	db *gorm.DB
)

func init() {
	rand.Seed(time.Now().UnixNano())
	dsn := "root:Wfs123456@(localhost)/simple?charset=utf8&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database" + err.Error())
	}
}

func TestGorm(t *testing.T) {
	//db, err := gorm.Open("mysql", "root:@(localhost)/simple?charset=utf8&parseTime=True&loc=Local")
	//if err != nil {
	//	panic("failed to connect database" + err.Error())
	//}
	//defer db.Close()
	//
	//// Migrate the schema
	//db.AutoMigrate(&Product{})
	//
	//// 创建
	//db.Create(&Product{Code: "L1212", Price: 1000})
	//printStats(db)
	//// 读取
	//var product Product
	//db.First(&product, 1) // 查询id为1的product
	//printStats(db)
	//db.First(&product, "code = ?", "L1212") // 查询code为l1212的product
	//printStats(db)
	//// 更新 - 更新product的price为2000
	//db.Model(&product).Update("Price", 2000)
	//printStats(db)

	//db.Exec()
	// 删除 - 删除product
	//db.Delete(&product)
}

func TestGormUser(t *testing.T) {
	rnd := strconv.Itoa(rand.Int())
	// Migrate the schema 可以更新表结构
	err := db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal(err)
	}
	// 创建
	db.Create(&User{
		Name:         "name" + rnd,
		Gender:       "male",
		MemberNumber: "member" + rnd,
		Email:        rnd + "@email.com",
		Height: 10000.0/3.0,
	})
	if db.Error != nil {
		log.Fatal(db.Error)
	}

	// 读取
	var user1 User
	db.First(&user1, 1) // 查询id为1的product
	if db.Error != nil {
		log.Fatal(db.Error)
	}
	log.Printf("%+v\n", user1)
	db.First(&user1, "age = ?", "11") // 查询code为l1212的product
	if db.Error != nil {
		log.Fatal(db.Error)
	}
	log.Printf("%+v\n", user1)
	// 更新 - 更新product的price为2000
	db.Model(&user1).Update("name", "xx")
	if db.Error != nil {
		log.Fatal(db.Error)
	}
	log.Printf("%+v\n", user1)
	// 删除 - 删除product
	//db.Delete(&product)

	//db.NewRecord()
}

func printStats(db *gorm.DB) {
	sqldb, err := db.DB()
	if err != nil {
		panic("get sql.db error:" + err.Error())
	}

	log.Printf("-->%#v\n", sqldb.Stats())
}
