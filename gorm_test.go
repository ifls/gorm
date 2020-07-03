package gorm_test

import (
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
	ID           int64  `gorm:"primary_key"`
	Name         string `gorm:"column:namexx"`
	Face         int    `gorm:"type:smallint"`
	Gender       string
	Age          int     `gorm:"default:11"`
	Height       float64 `gorm:"type:double(9,2);precision:4"` //精度没看到效果
	Birthday     *time.Time
	Email        string `gorm:"type:varchar(100);unique_index"`
	Role         string `gorm:"size:255"`        // 设置字段大小为255
	MemberNumber string `gorm:"unique;not null"` // 设置会员号（member number）唯一并且不为空
	Address      string `gorm:"index:addr"`      // 给address字段创建名为addr的索引
	//Pid          int    `gorm:"AUTOINCREMENT"`   // 和default value 冲突
	IgnoreMe int `gorm:"-"` // 忽略本字段

	Company struct {
		Name   string
		Number int
	} `gorm:"EMBEDDED"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	log.Printf("user:%+v\n", u)
	return nil
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
	db = db.Debug()
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
		Height:       10000.0 / 3.0,
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

}

func TestGormUserCreate(t *testing.T) {
	// Migrate the schema 可以更新表结构
	//err := db.AutoMigrate(&User{})
	//if err != nil {
	//	log.Fatal(err)
	//}

	n := 1
	for i := 0; i < n; i++ {
		createUser()
	}
}

func createUser() {
	r := rand.Int()
	rs := strconv.Itoa(r)
	rm := r % 100
	rms := strconv.Itoa(rm)

	t := time.Now()
	// 创建
	db.Create(&User{
		Name:         "name" + rms,
		Gender:       "male",
		MemberNumber: "member" + rs,
		Email:        rms + "@email.com",
		Height:       10000.0 / 3.0,
		Age:          rm,
		Birthday:     &t,
	})
}

func TestGormUserUpdate(t *testing.T) {
	var u User
	u.ID = 1
	checkUser(func(uc *User) {
		db.First(&u)
	})

	u.Name = "222"
	checkUser(func(uc *User) {
		db.Save(&u)
	})

	u.Name = "333"
	checkUser(func(uc *User) {
		db.Save(&u)
	})
}

func TestGormUserDelete(t *testing.T) {
	var u User
	u.ID = 1
	checkUser(func(uc *User) {
		db.First(&u)
	})

	checkUser(func(uc *User) {
		db.Delete(&u)
	})
}

func TestGormUserSelect(t *testing.T) {
	//logger.Default.LogMode(logger.Info)

	// 读取
	checkUser(func(u *User) {
		db.First(&u, 1)
	}) // 查询id为1的product

	checkUser(func(u *User) {
		db.First(&u, "age = ?", "11")
	})

	checkUser(func(u *User) {
		db.Take(&u, "age = ?", "11")
	})

	checkUser(func(u *User) {
		db.Last(&u, "age = ?", "12")
	})

	checkUsers(func(u *[]User) {
		db.Find(&u, "age = ?", "11")
	})

	// ==
	checkUsers(func(u *[]User) {
		db.Where("age = ?", "11").Find(&u)
	})

	// !=
	checkUsers(func(u *[]User) {
		db.Where("age <> ?", "11").Find(&u)
	})

	// IN
	checkUsers(func(u *[]User) {
		db.Where("age IN (?)", []string{"41", "12"}).Find(&u)
	})

	// LIKE
	checkUsers(func(u *[]User) {
		db.Where("namexx LIKE ?", "name2%").Find(&u)
	})

	// AND
	checkUsers(func(u *[]User) {
		db.Where("namexx LIKE ? AND Age = ?", "name2%", "20").Find(&u)
	})

	// TIME
	checkUsers(func(u *[]User) {
		db.Where("birthday < ?", time.Now().Add(-10*time.Minute).Format("2006-01-02 15:04:05")).Find(&u)
	})

	// TIME
	checkUsers(func(u *[]User) {
		db.Where("birthday BETWEEN ? AND ?", time.Now().Add(-10*time.Minute).Format("2006-01-02 15:04:05"), time.Now().Add(-5*time.Minute).Format("2006-01-02 15:04:05")).Find(&u)
	})

	// struct
	checkUsers(func(u *[]User) {
		db.Where(&User{Name: "name74", Age: 74}).Find(&u)
	})

	// map
	checkUsers(func(u *[]User) {
		db.Where(map[string]interface{}{"namexx": "name74", "age": 74}).Find(&u)
	})

	// []int 主键切片  Error 1241: Operand should contain 1 column(s)
	checkUsers(func(u *[]User) {
		db.Where([]int64{20, 21, 22}).Find(&u)
	})

	// OR
	checkUsers(func(u *[]User) {
		db.Where(&User{Name: "name74"}).Or("age = ?", "12").Find(&u)
	})

	//FOR UPDATE 无效
	checkUser(func(u *User) {
		db.Set("gorm:query_option", "FOR UPDATE").First(&u, 10)
	})

	// FirstOrInit
	checkUser(func(u *User) {
		db.FirstOrInit(u, User{Name: "non_existing"})
	})

	// FirstOrCreate 查不到则插入
	checkUser(func(u *User) {
		db.FirstOrCreate(u, User{Name: "non_existing"})
	})

	//
	checkUsers(func(u *[]User) {
		db.Table("users").Select("namexx, age").Scan(&u)
	})
}

func checkUser(fn func(u *User)) {
	var user1 User
	fn(&user1)
	if db.Error != nil {
		log.Fatal(db.Error)
	}
	log.Printf("%+v\n", user1)
}

func checkUsers(fn func(u *[]User)) {
	var user1 []User
	fn(&user1)
	if db.Error != nil {
		log.Fatal(db.Error)
	}
	for _, u := range user1 {
		log.Printf("findMany %+v\n", u)
	}
}

func printStats(db *gorm.DB) {
	sqldb, err := db.DB()
	if err != nil {
		panic("get sql.db error:" + err.Error())
	}

	log.Printf("-->%#v\n", sqldb.Stats())
}
