package utils

import (
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
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

//RegisterCustomValidations is function for adding validation for fields and struct
func RegisterCustomValidations(v *validator.Validate) error {
	if err := v.RegisterValidation("is_currency", checkIsCurrency); err != nil {
		return errors.Wrap(err, "Register validation for 'is_currency' failed")
	}

	return nil
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
