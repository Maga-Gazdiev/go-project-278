package errors

import "errors"

var (
	ErrLinkNotFound   = errors.New("не найдена ссылка")
	ErrLinkNotCreated = errors.New("не удалось создать ссылку")
	ErrLinkNotDeleted = errors.New("не удалось удалить ссылку")
	ErrLinkNotUpdated = errors.New("не удалось обновить ссылку")
	ErrInvalidLinkID  = errors.New("некорректный id ссылки")
	ErrShortNameTaken = errors.New("короткое имя ссылки уже используется")
	ErrNotValidQuery  = errors.New("введен неправильный параметр запроса, ожидание в формате [1, 10]")
)
