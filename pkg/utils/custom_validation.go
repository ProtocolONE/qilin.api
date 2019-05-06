package utils

import (
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
	"qilin-api/pkg/model"
)

type Currency string

const (
	Ruble         Currency = "RUB"
	Euro          Currency = "EUR"
	UsDollar      Currency = "USD"
	PoundSterling Currency = "GBP"
)

var currencyList = []Currency{
	Ruble,
	Euro,
	UsDollar,
	PoundSterling,
}

var acceptableRoles = []string {
	model.Accountant,
	model.Support,
	model.Developer,
	model.Publisher,
	model.Manager,
}

//RegisterCustomValidations is function for adding validation for fields and struct
func RegisterCustomValidations(v *validator.Validate) error {
	if err := v.RegisterValidation("is_currency", checkIsCurrency); err != nil {
		return errors.Wrap(err, "Register validation for 'is_currency' failed")
	}

	if err := v.RegisterValidation("non_admin_role", checkNonAdminRole); err != nil {
		return errors.Wrap(err, "Register validation for 'non_admin_role' failed")
	}

	return nil
}

func checkNonAdminRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	for _, r := range acceptableRoles {
		if r == role {
			return true
		}
	}

	return false
}

func checkIsCurrency(fl validator.FieldLevel) bool {
	currency := fl.Field().String()
	return IsCurrency(currency)
}

//IsCurrency is function that checks string for available currencies
func IsCurrency(cur string) bool {
	for _, c := range currencyList {
		if string(c) == cur {
			return true
		}
	}
	return false
}