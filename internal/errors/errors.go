package errors

import "errors"

var (
	ErrLinkNotFound   = errors.New("Не найдена ссылка")
	ErrLinkNotCreated = errors.New("Не удалось создать ссылку")
	ErrLinkNotDeleted = errors.New("Не удалось удалить ссылку")
	ErrLinkNotUpdated = errors.New("Не удалось обновить ссылку")
	ErrInvalidLinkID  = errors.New("Некорректный id ссылки")
	ErrShortNameTaken = errors.New("Короткое имя ссылки уже используется")
	ErrNotValidQuery  = errors.New("Введен неправильный параметр запроса, ожидание в формате [1, 10]")
)
