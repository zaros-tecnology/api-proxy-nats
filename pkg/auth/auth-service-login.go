package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"

	"github.com/zaros-tecnology/api-proxy-nats/pkg/models"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service/request"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/utils"

	"github.com/gofrs/uuid"

	oauthModels "gopkg.in/oauth2.v3/models"
)

func (s *srv) login(r *request.CallRequest) {

	var u models.User
	r.ParseData(&u)

	if u.Email == "" || u.Password == "" {
		r.ErrorRequest(&request.ErrorNotFound)
		return
	}

	var user models.User
	err := s.DB().Preload("Profiles").Where(&models.User{Email: u.Email}).First(&user).Error
	if err != nil {
		r.Error(err)
		return
	}

	if user.ID == uuid.Nil {
		r.Error(fmt.Errorf("Invalid email or password"))
		return
	}

	h := utils.Hash{}
	err = h.Compare(user.Password, u.Password)
	if err != nil {
		r.Error(fmt.Errorf("Invalid email or password"))
		return
	}

	clientID, _ := uuid.NewV4()
	clientSecret, _ := uuid.NewV4()

	err = s.Create(&oauthModels.Client{
		ID:     clientID.String()[:8],
		Secret: clientSecret.String()[:8],
		Domain: "",
		UserID: user.ID.String(),
	})
	if err != nil {
		r.Error(err)
		return
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET",
		fmt.Sprintf("https://generatetoken.com.br/token?grant_type=client_credentials&client_id=%v&client_secret=%v&scope=all",
			clientID.String()[:8],
			clientSecret.String()[:8]), nil)

	err = s.HandleTokenRequest(w, req)
	if err != nil {
		r.Error(err)
		return
	}

	defer w.Result().Body.Close()

	data, err := ioutil.ReadAll(w.Result().Body)
	if err != nil {
		r.Error(err)
		return
	}

	var token models.Token
	json.Unmarshal(data, &token)

	token.Name = user.Nome

	r.OK(token)
}
