package l10n_test

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/l10n"

	_ "github.com/go-sql-driver/mysql"
)

type Product struct {
	ID int `gorm:"primary_key"`
	l10n.Locale
	Code      string `l10n:"sync"`
	Name      string
	DeletedAt time.Time
}

func (Product) LocaleCreateable() {}

var dbGlobal, dbCN, dbEN *gorm.DB

func init() {
	// CREATE USER 'qor'@'localhost' IDENTIFIED BY 'qor';
	// CREATE DATABASE qor_l10n;
	// GRANT ALL ON qor_l10n.* TO 'gorm'@'localhost';
	db, _ := gorm.Open("mysql", "qor:qor@/qor_l10n?charset=utf8&parseTime=True")
	l10n.RegisterCallbacks(&db)

	db.DropTable(&Product{})
	db.AutoMigrate(&Product{})

	dbGlobal = &db
	dbCN = dbGlobal.Set("l10n:locale", "zh")
	dbEN = dbGlobal.Set("l10n:locale", "en")
}

func checkHasErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func checkHasProductInLocale(db *gorm.DB, locale string, t *testing.T) {
	var count int
	if db.Where("language_code = ?", locale).Count(&count); count != 1 {
		t.Errorf("should has only one product for locale %v, but found %v", locale, count)
	}
}

func checkHasProductInAllLocales(db *gorm.DB, t *testing.T) {
	checkHasProductInLocale(db, "", t)
	checkHasProductInLocale(db, "zh", t)
	checkHasProductInLocale(db, "en", t)
}

func TestCreateWithCreate(t *testing.T) {
	product := Product{Code: "CreateWithCreate"}
	checkHasErr(t, dbGlobal.Create(&product).Error)
	checkHasErr(t, dbCN.Create(&product).Error)
	checkHasErr(t, dbEN.Create(&product).Error)

	checkHasProductInAllLocales(dbGlobal.Model(&Product{}).Where("id = ? AND code = ?", product.ID, "CreateWithCreate"), t)
}

func TestCreateWithSave(t *testing.T) {
	product := Product{Code: "CreateWithSave"}
	checkHasErr(t, dbGlobal.Create(&product).Error)
	checkHasErr(t, dbCN.Create(&product).Error)
	checkHasErr(t, dbEN.Create(&product).Error)

	checkHasProductInAllLocales(dbGlobal.Model(&Product{}).Where("id = ? AND code = ?", product.ID, "CreateWithSave"), t)
}

func TestUpdate(t *testing.T) {
	product := Product{Code: "Update", Name: "global"}
	checkHasErr(t, dbGlobal.Create(&product).Error)
	sharedDB := dbGlobal.Model(&Product{}).Where("id = ? AND code = ?", product.ID, "Update")

	product.Name = "中文名"
	checkHasErr(t, dbCN.Create(&product).Error)
	checkHasProductInLocale(sharedDB.Where("name = ?", "中文名"), "zh", t)

	product.Name = "English Name"
	checkHasErr(t, dbEN.Create(&product).Error)
	checkHasProductInLocale(sharedDB.Where("name = ?", "English Name"), "en", t)

	product.Name = "新的中文名"
	product.Code = "NewCode // should be ignored when update"
	dbCN.Save(&product)
	checkHasProductInLocale(sharedDB.Where("name = ?", "新的中文名"), "zh", t)

	product.Name = "New English Name"
	product.Code = "NewCode // should be ignored when update"
	dbEN.Save(&product)
	checkHasProductInLocale(sharedDB.Where("name = ?", "New English Name"), "en", t)
}

func TestQuery(t *testing.T) {
	product := Product{Code: "Query", Name: "global"}
	dbGlobal.Create(&product)
	dbCN.Create(&product)

	var productCN Product
	dbCN.First(&productCN, product.ID)
	if productCN.LanguageCode != "zh" {
		t.Error("Should find localized zh product with mixed mode")
	}

	if dbCN.Set("l10n:mode", "locale").First(&productCN, product.ID).RecordNotFound() {
		t.Error("Should find localized zh product with locale mode")
	}

	if dbCN.Set("l10n:mode", "global").First(&productCN); productCN.LanguageCode != "" {
		t.Error("Should find global product with global mode")
	}

	var productEN Product
	dbEN.First(&productEN, product.ID)
	if productEN.LanguageCode != "" {
		t.Error("Should find global product for en with mixed mode")
	}

	if !dbEN.Set("l10n:mode", "locale").First(&productEN, product.ID).RecordNotFound() {
		t.Error("Should find no record with locale mode")
	}

	if dbEN.Set("l10n:mode", "global").First(&productEN); productEN.LanguageCode != "" {
		t.Error("Should find global product with global mode")
	}
}

func TestDelete(t *testing.T) {
	product := Product{Code: "Delete", Name: "global"}
	dbGlobal.Create(&product)
	dbCN.Create(&product)

	if dbCN.Delete(&product).RowsAffected != 1 {
		t.Errorf("Should delete localized record")
	}

	if dbEN.Delete(&product).RowsAffected != 0 {
		t.Errorf("Should delete none record in unlocalized locale")
	}
}