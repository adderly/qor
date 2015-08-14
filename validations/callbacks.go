package validations

import "github.com/jinzhu/gorm"

var skipValidations = "validations:skip_validations"

func validate(scope *gorm.Scope) {
	if result, ok := scope.DB().Get(skipValidations); !(ok && result.(bool)) {
		scope.CallMethodWithErrorCheck("Validate")
	}
}

func RegisterCallbacks(db *gorm.DB) {
	callback := db.Callback()
	callback.Create().After("gorm:before_create").Register("validations:validate", validate)
	callback.Update().After("gorm:before_update").Register("validations:validate", validate)
}