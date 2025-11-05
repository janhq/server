package dbschema

import (
	"jan-server/services/llm-api/internal/domain/user"
	"jan-server/services/llm-api/internal/infrastructure/database"
)

func init() {
	database.RegisterSchemaForAutoMigrate(User{})
}

// User represents the persisted user schema tied to an external identity provider.
type User struct {
	BaseModel
	AuthProvider string  `gorm:"type:varchar(50);not null;default:'keycloak'"`
	Issuer       string  `gorm:"type:varchar(255);not null;uniqueIndex:ux_users_issuer_subject"`
	Subject      string  `gorm:"type:varchar(255);not null;uniqueIndex:ux_users_issuer_subject"`
	Username     *string `gorm:"type:varchar(150)"`
	Email        *string `gorm:"type:varchar(320)"`
	Name         *string `gorm:"type:varchar(255)"`
	Picture      *string `gorm:"type:varchar(512)"`
}

// NewSchemaUser converts a domain user into a schema instance.
func NewSchemaUser(u *user.User) *User {
	if u == nil {
		return nil
	}

	return &User{
		BaseModel: BaseModel{
			ID:        u.ID,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		},
		AuthProvider: u.AuthProvider,
		Issuer:       u.Issuer,
		Subject:      u.Subject,
		Username:     u.Username,
		Email:        u.Email,
		Name:         u.Name,
		Picture:      u.Picture,
	}
}

// EtoD converts a schema user back to the domain representation.
func (u *User) EtoD() *user.User {
	if u == nil {
		return nil
	}

	return &user.User{
		ID:           u.ID,
		AuthProvider: u.AuthProvider,
		Issuer:       u.Issuer,
		Subject:      u.Subject,
		Username:     u.Username,
		Email:        u.Email,
		Name:         u.Name,
		Picture:      u.Picture,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
