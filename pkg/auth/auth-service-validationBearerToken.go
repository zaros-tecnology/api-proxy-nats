package auth

import (
	"net/http"
	"strconv"
	"time"

	"github.com/zaros-tecnology/api-proxy-nats/internal/models"
	"github.com/zaros-tecnology/api-proxy-nats/internal/utils"
)

// RemoveAccessToken remove token
func (s *srv) RemoveAccessToken(token string) (err error) {
	delete(s.userToken, token)
	delete(s.tokenExpires, token)
	return s.DB().Where(&models.Oauth2Tokens{Access: token}).Delete(&models.Oauth2Tokens{}).Error
}

func (s *srv) HandleTokenRequest(w http.ResponseWriter, r *http.Request) error {
	return s.server.HandleTokenRequest(w, r)
}

func (s *srv) ValidationBearerToken(r *http.Request, w http.ResponseWriter) bool {
	token, ok := utils.GetBearer(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}

	if user := s.userToken[token]; user != nil {

		if s.tokenExpires[token].Before(time.Now().UTC()) {
			s.RemoveAccessToken(token)
			return false
		}

		r.Header.Add("User", string(models.ToJSON(user)))
		return true
	}

	var userToken models.Oauth2Tokens

	err := s.DB().Raw(userToken.QueryByToken(), token).First(&userToken).Error
	if err != nil || userToken.Expired {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return false
	}

	var client models.Oauth2Clients

	err = s.DB().Raw(client.QueryByID(), userToken.DataToken.ClientID).First(&client).Error
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return false
	}

	var user models.User

	code, _ := strconv.Atoi(client.Oauth2ClientsData.UserID)
	err = s.DB().First(&user, code).Error
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return false
	}

	r.Header.Add("User", string(models.ToJSON(user)))

	s.userToken[token] = &user

	s.setAccessExpiresIn(token, user)

	return true
}
func (s *srv) setAccessExpiresIn(accessToken string, oauth models.User) {
	go func() {
		s.m.Lock()
		defer s.m.Unlock()
		t := time.Now().UTC().Add(time.Hour * 3)

		s.userToken[accessToken] = &oauth
		s.tokenExpires[accessToken] = &t

		s.DB().Model(&oauth).Where(&models.Oauth2Tokens{Access: accessToken}).Update("expires_at", t)
	}()
}
