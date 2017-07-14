package services

import (
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/tokens/password_resets"
)

func PasswordResetter(store data.AccountStore, cfg *config.Config, token string, password string) error {
	claims, err := password_resets.Parse(token, cfg)
	if err != nil {
		return FieldErrors{{"token", ErrInvalidOrExpired}}
	}

	id, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return err
	}

	account, err := store.Find(id)
	if err != nil {
		return err
	}
	if account == nil {
		return FieldErrors{{"account", ErrNotFound}}
	} else if account.Locked {
		return FieldErrors{{"account", ErrLocked}}
	} else if account.Archived() {
		return FieldErrors{{"account", ErrLocked}}
	}

	if claims.LockExpired(account.PasswordChangedAt) {
		return FieldErrors{{"token", ErrInvalidOrExpired}}
	}

	fieldError := passwordValidator(cfg, password)
	if fieldError != nil {
		return FieldErrors{*fieldError}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
	if err != nil {
		return err
	}

	err = store.SetPassword(id, hash)
	if err != nil {
		return err
	}
	return nil
}
