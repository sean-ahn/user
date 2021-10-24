package mysql

import (
	"context"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/sean-ahn/user/backend/model"
)

func GetUser(ctx context.Context, exec boil.ContextExecutor, id int) (*model.User, error) {
	u, err := model.Users(model.UserWhere.UserID.EQ(id)).One(ctx, exec)
	if err != nil {
		return nil, err
	}
	return u, nil
}
