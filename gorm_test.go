package gorm_test

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

type Product struct {
	gorm.Model
	Code string
	Price uint
}

func TestGorm(t *testing.T) {
	//gorm.Open(mysql.Open(dbDSN), &gorm.Config{})
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
	//
	////db.Exec()
	//// 删除 - 删除product
	//db.Delete(&product)
}

func TestGorm2(t *testing.T) {
	dbDSN := "root:@(localhost)/simple?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dbDSN), &gorm.Config{})
	//db, err := gorm.Open("mysql", )
	if err != nil {
		panic("failed to connect database" + err.Error())
	}
	//defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Product{})

	// 创建
	db.Create(&Product{Code: "L1212", Price: 1000})
	printStats(db)

	// 读取
	var product Product
	db.First(&product, 1) // 查询id为1的product
	db.Save(&product)
	db.Exec("select * from products")
	db.Delete(&product)
	printStats(db)
	db.First(&product, "code = ?", "L1212") // 查询code为l1212的product
	printStats(db)
	// 更新 - 更新product的price为2000
	db.Model(&product).Update("Price", 2000)
	printStats(db)

	//db.Exec()
	// 删除 - 删除product
	//db.Delete(&product)
}

func printStats(db *gorm.DB)  {
	//idb := db.CommonDB()
	//
	//if d, ok := idb.(*sql.DB); ok {
	//	st := d.Stats()
	//	log.Printf("%#v\n", st)
	//}

	//log.Printf("-->%#v\n", db.DB().Stats())
}