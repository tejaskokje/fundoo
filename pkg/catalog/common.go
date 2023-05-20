package catalog

import (
	"errors"

	"github.com/twitchtv/twirp"
)

var (
	ErrSkuNameCategoryRequired = twirp.InvalidArgument.Error("sku, category and name fields are required")
	ErrSearchQueryRequired     = twirp.InvalidArgument.Error("search query required")
	ErrNoResultFound           = twirp.NotFoundError("no result found")
	ErrInvalidProduct          = errors.New("invalid product")
	ErrProductAlreadyExists    = twirp.AlreadyExists.Error("product already exists")
)
