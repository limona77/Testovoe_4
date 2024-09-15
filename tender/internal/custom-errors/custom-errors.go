package custom_errors

import "errors"

var (
	ErrTenderAlreadyExists = errors.New("такой тендер уже существует")
	ErrTenderNotFound      = errors.New("тендер не найден")
	ErrBidsNotFound        = errors.New("предложения не найдены")
	ErrUnprocessableEntity = errors.New("неправильные данные")
	ErrBidsAlreadyExists   = errors.New("предложение уже существует")
	ErrAccessDenied        = errors.New("у вас недостаточно прав")
	ErrUserNotFound        = errors.New("пользователь не найден")
)
