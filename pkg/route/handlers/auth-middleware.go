package handlers

import (
	"net/http"
	"strings"

	"github.com/zaros-tecnology/api-proxy-nats/pkg/service"
)

// AuthMiddleware middleware de autenticação oauth.v3
func AuthMiddleware(oauth ...*service.AuthRid) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, auth := range oauth {
				if strings.Contains(r.RequestURI, "/ws") {
					break
				}
				request := strings.Split(r.RequestURI, "/")
				if len(request) > 2 && request[2] == auth.Rid.Name() {

					var validBearer bool = true
					for _, p := range auth.Patterns {
						if !p.Auth() {
							endpoint := "/api/" + strings.ReplaceAll(p.EndpointNoMethod(), ".", "/")
							if r.RequestURI == endpoint && r.Method == p.Method {
								validBearer = false
								break
							}

							if strings.Contains(p.EndpointNoMethod(), "$") {
								partsEndpoint := strings.Split(endpoint, "/")
								partsRequest := strings.Split(r.RequestURI, "/")
								if len(partsEndpoint) == len(partsRequest) {
									match := true
									for i := range partsRequest {
										if strings.Contains(partsEndpoint[i], "$") {
											continue
										}
										if partsEndpoint[i] != partsRequest[i] {
											validBearer = true
											match = false
											break
										}
									}
									if match {
										validBearer = false
									}
								}
							}
						}

					}

					if validBearer {
						ok := auth.Auth.ValidationBearerToken(r, w)
						if !ok {
							//http.Error(w, err.Error(), http.StatusBadRequest)
							return
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
