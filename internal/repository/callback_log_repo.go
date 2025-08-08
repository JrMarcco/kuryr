package repository

import "github.com/JrMarcco/kuryr/internal/repository/dao"

type CallbackLogRepo interface{}

var _ CallbackLogRepo = (*DefaultCallbackLogRepo)(nil)

type DefaultCallbackLogRepo struct {
	dao dao.CallbackLogDao
}

func NewCallbackLogRepo(dao dao.CallbackLogDao) CallbackLogRepo {
	return &DefaultCallbackLogRepo{
		dao: dao,
	}
}
