package route

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/zaros-tecnology/api-proxy-nats/pkg/models"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/rids"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service/request"

	"github.com/go-chi/chi"
)

func (b *routeService) handler(endpoint rids.EndpointRest, w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	params := make(map[string]string)
	for _, p := range endpoint.Params {
		value := r.FormValue(p)
		if value != "" {
			params[p] = value
		}
	}
	for _, p := range endpoint.Params {
		value := chi.URLParam(r, p)
		if value != "" {
			params[p] = value
		}
	}

	call := request.CallRequest{
		Data:     data,
		Params:   params,
		Form:     r.Form,
		Header:   r.Header,
		Endpoint: endpoint.Endpoint,
	}

	values := strings.Split(endpoint.Endpoint, ".")

	permission := models.HavePermissionRequest{
		Service:  values[0],
		Endpoint: strings.ReplaceAll(strings.Join(values, "."), "."+values[len(values)-1], ""),
		Method:   values[len(values)-1],
	}

	if endpoint.Authenticated {
		if len(b.auths) > 0 {
		}
		callAuth := request.NewRequest(permission)
		callAuth.Form = r.Form
		callAuth.Header = r.Header

		if !b.auths[0].Auth.UserHavePermission(callAuth) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	result, err := b.Nats().Conn.Request(endpoint.Endpoint, call.ToJSON(), b.Nats().Timeout())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	var er request.ErrorRequest
	if er.Parse(result.Data); er.Message != "" {
		w.WriteHeader(er.Code)
		w.Write(er.ToJSON())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(result.Data)
}
