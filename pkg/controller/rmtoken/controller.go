package rmtoken

import "github.com/sirupsen/logrus"

type Controller struct {
	param        *Param
	tokenManager TokenManager
}

func New(param *Param, tokenManager TokenManager) *Controller {
	return &Controller{
		param:        param,
		tokenManager: tokenManager,
	}
}

type Param struct{}

type TokenManager interface {
	Remove(logE *logrus.Entry) error
}
