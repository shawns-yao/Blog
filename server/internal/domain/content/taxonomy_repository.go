package content

import "context"

type ColumnRepository interface {
	List(ctx context.Context) ([]*MomentColumn, error)
	GetByID(ctx context.Context, id int64) (*MomentColumn, error)
	Create(ctx context.Context, column *MomentColumn) error
	Update(ctx context.Context, column *MomentColumn) error
	Delete(ctx context.Context, id int64) error
}

type TagRepository interface {
	List(ctx context.Context) ([]*Tag, error)
	ListPublicWithMomentCount(ctx context.Context) ([]*TagPublicItem, error)
	GetByID(ctx context.Context, id int64) (*Tag, error)
	Create(ctx context.Context, tag *Tag) error
	Update(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, id int64) error
}
