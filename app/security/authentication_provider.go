package security

import (
	"thumbai/app/access"

	"aahframe.work/config"
	"aahframe.work/security/authc"
)

var _ authc.Authenticator = (*AuthenticationProvider)(nil)

// AuthenticationProvider struct implements `authc.Authenticator` interface.
type AuthenticationProvider struct {
}

// Init method initializes the AuthenticationProvider, this method gets called
// during server start up.
func (a *AuthenticationProvider) Init(appCfg *config.Config) error {
	// NOTE: Init is called on application startup
	return nil
}

// GetAuthenticationInfo method is `authc.Authenticator` interface
func (a *AuthenticationProvider) GetAuthenticationInfo(authcToken *authc.AuthenticationToken) (*authc.AuthenticationInfo, error) {
	u, found := access.UserStore[authcToken.Identity]
	if !found {
		return nil, authc.ErrSubjectNotExists
	}

	authcInfo := authc.NewAuthenticationInfo()
	authcInfo.Principals = append(authcInfo.Principals,
		&authc.Principal{
			Value:     u.Username,
			IsPrimary: true,
			Realm:     "inmemory",
		})
	authcInfo.Credential = u.Password
	authcInfo.IsLocked = u.Locked
	authcInfo.IsExpired = u.Expired

	return authcInfo, nil
}
