package security

import (
	"thumbai/app/access"

	"aahframe.work/aah/config"
	"aahframe.work/aah/security/authc"
	"aahframe.work/aah/security/authz"
)

var _ authz.Authorizer = (*AuthorizationProvider)(nil)

// AuthorizationProvider struct implements `authz.Authorizer` interface.
type AuthorizationProvider struct {
}

// Init method initializes the AuthorizationProvider, this method gets called
// during server start up.
func (a *AuthorizationProvider) Init(appCfg *config.Config) error {
	// NOTE: Init is called on application startup
	return nil
}

// GetAuthorizationInfo method is `authz.Authorizer` interface.
//
// GetAuthorizationInfo method gets called after authentication is successful
// to get Subject's (aka User) access control information such as roles and permissions.
func (a *AuthorizationProvider) GetAuthorizationInfo(authcInfo *authc.AuthenticationInfo) *authz.AuthorizationInfo {
	authzInfo := authz.NewAuthorizationInfo()
	u, found := access.UserStore[authcInfo.PrimaryPrincipal().Value]
	if found {
		authzInfo.AddPermissionString(u.Permissions...)
	}
	return authzInfo
}
