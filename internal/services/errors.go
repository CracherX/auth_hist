package services

import "errors"

var ErrInvalidToken = errors.New("access токен не валиден")
var ErrInvalidRefreshToken = errors.New("отправленный refresh токен не связан с отправленным access токеном или вообще не существует в БД")
