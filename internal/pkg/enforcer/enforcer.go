package enforcer

import "github.com/casbin/casbin/v3"

type Enforcer interface {
	Enforce(subject string, object string, action string) (bool, error)
}

type Casbin struct {
	casbin *casbin.Enforcer
}

func NewCasbin(casbin *casbin.Enforcer) *Casbin {
	return &Casbin{casbin: casbin}
}

func (c *Casbin) Enforce(subject string, object string, action string) (bool, error) {
	return c.casbin.Enforce(subject, object, action)
}
