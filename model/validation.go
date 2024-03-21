package model

import (
	"github.com/biter777/countries"
	"github.com/go-playground/validator"
)

// Validation functions
func isCountryCode(fl validator.FieldLevel) bool {
	country := fl.Field().String()
	if country == "" {
		return true
	}
	countryCode := countries.ByName(country)
	return countryCode != countries.Unknown
}

func validateGender(fl validator.FieldLevel) bool {
	gender := fl.Field().String()
	switch Gender(gender) {
	case Male, Female, "":
		return true
	default:
		return false
	}
}

func validatePlatform(fl validator.FieldLevel) bool {
	platform := fl.Field().String()
	switch Platform(platform) {
	case Android, IOS, Web, "":
		return true
	default:
		return false
	}
}

func validateAgeRange(fl validator.FieldLevel) bool {
	ageStart := fl.Parent().FieldByName("AgeStart").Int()
	ageEnd := fl.Parent().FieldByName("AgeEnd").Int()
	return ageStart == 0 || ageEnd == 0 || ageStart <= ageEnd
}

func Validator() *validator.Validate {
	validate := validator.New()
	validate.RegisterValidation("countrycode", isCountryCode)
	validate.RegisterValidation("gender", validateGender)
	validate.RegisterValidation("platform", validatePlatform)
	validate.RegisterValidation("ageRange", validateAgeRange)

	return validate
}
