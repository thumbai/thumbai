package security

import (
	"aahframework.org/config.v0"
	"aahframework.org/security.v0/authc"
	"aahframework.org/security.v0/authz"
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
	//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
	// This code snippet provided as a reference
	//
	// Call your appropriate datasource here (such as DB, API, etc)
	// to get the subject (aka user) authorization details (roles, permissions)
	//__________________________________________________________________________

	// authorities := models.FindUserByEmail(authcInfo.PrimaryPrincipal().Value)

	authzInfo := authz.NewAuthorizationInfo()
	// authzInfo.AddRole(authorities.Roles...)
	// authzInfo.AddPermissionString(authorities.Permissions...)

	return authzInfo
}
