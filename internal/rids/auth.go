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
	return s.newMethod("", "ValidationBearerToken.$token").internal()
}

func (s *auth) Login() *Pattern {
	return s.newMethod("", "login").noAuth().post()
}

func (s *auth) HavePermission() *Pattern {
	return s.newMethod("", "havePermission").internal()
}
