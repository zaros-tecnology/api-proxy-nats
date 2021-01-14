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
	return r.newMethod("", "newMessageSocket").internal()
}

func (r *route) NewService() *Pattern {
	return r.newMethod("", "newService").internal()
}

func (r *route) Restart() *Pattern {
	return r.newMethod("", "restart").internal()
}
