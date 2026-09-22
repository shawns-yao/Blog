package content

import "errors"

var ErrColumnNotFound = errors.New("手记分区不存在")
var ErrColumnHierarchy = errors.New("分类最多两级，上级必须是一级分类，且不能将有子分类的分类移为二级")
var ErrColumnInUse = errors.New("分类仍有关联文章或二级分类，请先调整归属")
var ErrTagNotFound = errors.New("标签不存在")
var ErrMomentNotFound = errors.New("手记不存在")
var ErrMomentShortURLExists = errors.New("手记短链接已存在")
