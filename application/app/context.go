package app

import (
	"github.com/jfrog/jfrog-cli-application/application/service"
)

type Context interface {
	GetVersionService() service.VersionService
	GetSystemService() service.SystemService
	GetConfig() interface{}
}

type context struct {
	versionService service.VersionService
	systemService  service.SystemService
}

func NewAppContext() Context {
	return &context{
		versionService: service.NewVersionService(),
		systemService:  service.NewSystemService(),
	}
}

func (c *context) GetVersionService() service.VersionService {
	return c.versionService
}

func (c *context) GetSystemService() service.SystemService {
	return c.systemService
}

func (c *context) GetConfig() interface{} {
	return nil
}
