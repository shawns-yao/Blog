package content

import "errors"

var ErrColumnNotFound = errors.New("手记分区不存在")
var ErrTagNotFound = errors.New("标签不存在")
var ErrMomentNotFound = errors.New("手记不存在")
var ErrMomentShortURLExists = errors.New("手记短链接已存在")
