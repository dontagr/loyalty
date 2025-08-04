package checkup

import (
	"net/http"
	"reflect"
	"strconv"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func NewCustomValidator() (*CustomValidator, error) {
	validate := validator.New()
	err := validate.RegisterValidation("algLuna", algLunaValidator)
	if err != nil {
		return nil, err
	}
	err = validate.RegisterValidation("floatGtZero", floatGtZeroValidator)
	if err != nil {
		return nil, err
	}
	return &CustomValidator{validator: validate}, nil
}

func algLunaValidator(fl validator.FieldLevel) bool {
	switch v := fl.Field(); v.Kind() {
	case reflect.String:
		number := v.String()

		return algLuna(number)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		number := strconv.FormatInt(v.Int(), 10)

		return algLuna(number)
	default:
		return false
	}
}

func algLuna(number string) bool {
	var sum int
	nDigits := len(number)
	isSecond := false

	for i := nDigits - 1; i >= 0; i-- {
		dchar := number[i]
		if !unicode.IsDigit(rune(dchar)) {
			return false
		}

		digit, _ := strconv.Atoi(string(dchar))
		if isSecond {
			digit = digit * 2
		}
		if digit > 9 {
			digit = digit - 9
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}

func floatGtZeroValidator(fl validator.FieldLevel) bool {
	if fl.Field().Kind() == reflect.Float64 {
		return fl.Field().Float() > 0
	}
	return false
}
