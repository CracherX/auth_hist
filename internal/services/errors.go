package services

import "errors"

var ErrInvalidToken = errors.New("Access токен не валиден")
var ErrInvalidRefreshToken = errors.New("Отправленный refresh токен не связан с отправленным access токеном, не валиден или вообще не существует в БД")
