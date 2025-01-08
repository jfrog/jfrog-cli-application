package app

import (
	"github.com/jfrog/jfrog-cli-application/application/service/systems"
	"github.com/jfrog/jfrog-cli-application/application/service/versions"
)

type Context interface {
	GetVersionService() versions.VersionService
	GetSystemService() systems.SystemService
	GetConfig() interface{}
}

type context struct {
	versionService versions.VersionService
	systemService  systems.SystemService
}

func NewAppContext() Context {
	return &context{
		versionService: versions.NewVersionService(),
		systemService:  systems.NewSystemService(),
	}
}

func (c *context) GetVersionService() versions.VersionService {
	return c.versionService
}

func (c *context) GetSystemService() systems.SystemService {
	return c.systemService
}

func (c *context) GetConfig() interface{} {
	return nil
}
