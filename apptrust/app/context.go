package app

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/service/applications"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/systems"
	"github.com/jfrog/jfrog-cli-application/apptrust/service/versions"
)

type Context interface {
	GetApplicationService() applications.ApplicationService
	GetVersionService() versions.VersionService
	GetSystemService() systems.SystemService
	GetConfig() interface{}
}

type context struct {
	applicationService applications.ApplicationService
	versionService     versions.VersionService
	systemService      systems.SystemService
}

func NewAppContext() Context {
	return &context{
		applicationService: applications.NewApplicationService(),
		versionService:     versions.NewVersionService(),
		systemService:      systems.NewSystemService(),
	}
}

func (c *context) GetApplicationService() applications.ApplicationService {
	return c.applicationService
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
