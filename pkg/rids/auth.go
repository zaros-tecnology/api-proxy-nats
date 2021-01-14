package rids

type auth struct{ Base }

var authImp auth

// auth func
func Auth() *auth {
	authImp.name = "auth"
	authImp.label = "Autenticador"
	return &authImp
}

func (s *auth) ValidationBearerToken() *Pattern {
	return s.NewMethod("", "ValidationBearerToken.$token").Internal()
}

func (s *auth) Login() *Pattern {
	return s.NewMethod("", "login").NoAuth().Post()
}

func (s *auth) HavePermission() *Pattern {
	return s.NewMethod("", "havePermission").Internal()
}
