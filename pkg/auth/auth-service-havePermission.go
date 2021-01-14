package auth

import (
	"github.com/zaros-tecnology/api-proxy-nats/pkg/models"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service/request"
)

func (s *srv) userHavePermission(r *request.CallRequest) {

	var req models.HavePermissionRequest
	r.ParseData(&req)

	user := r.Usuario()

	var perm models.Permission
	err := s.DB().
		Raw(req.QueryPermission(), user.ID, req.Service, req.Endpoint, req.Method).
		First(&perm).Error
	if err != nil {
		r.Error(err)
		return
	}

	r.OK()
}
