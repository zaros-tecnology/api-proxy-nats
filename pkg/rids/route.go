package rids

type route struct{ Base }

var routeImp route

// Route func
func Route() *route {
	routeImp.name = "route"
	routeImp.label = "route"
	return &routeImp
}

func (r *route) SocketNewMessage() *Pattern {
	return r.NewMethod("", "newMessageSocket").Internal()
}

func (r *route) NewService() *Pattern {
	return r.NewMethod("", "newService").Internal()
}

func (r *route) Restart() *Pattern {
	return r.NewMethod("", "restart").Internal()
}
