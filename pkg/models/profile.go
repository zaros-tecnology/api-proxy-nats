package models

// Profile de acesso
type Profile struct {
	Base
	Name string

	Users       []*User       `gorm:"many2many:user_profile;"`
	Permissions []*Permission `gorm:"many2many:profile_permission;"`
}

// Permission perfil
type Permission struct {
	Base
	Service      string
	ServiceLabel string
	Nome         string
	Rid          string
	Method       string

	Profiles []*Profile `gorm:"many2many:profile_permission;"`
}

type AddRemoveUserProfileRequest struct {
	UserID uint `json:"userId"`
}

type ProfileCopyRequest struct {
	Name string `json:"nome"`
}

type ProfileCreateRequest struct {
	Name        string `json:"name" validate:"required"`
	Permissions []uint `json:"permissions"`
	Users       []uint `json:"users"`
}

type ProfileUpdateRequest struct {
	Name string `json:"nome" validate:"required"`
}
