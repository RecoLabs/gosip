package api

import (
	"bytes"
	"context"
	"testing"
)

func TestUsers(t *testing.T) {
	checkClient(t)

	sp := NewSP(spClient)
	users := sp.Web().SiteUsers()
	endpoint := spClient.AuthCnfg.GetSiteURL() + "/_api/Web/SiteUsers"
	user := &UserInfo{}

	t.Run("Constructor", func(t *testing.T) {
		u := NewUsers(spClient, endpoint, nil)
		if _, err := u.Select("Id").Get(context.Background()); err != nil {
			t.Error(err)
		}
	})

	t.Run("ToURL", func(t *testing.T) {
		if users.ToURL() != endpoint {
			t.Errorf(
				"incorrect endpoint URL, expected \"%s\", received \"%s\"",
				endpoint,
				users.ToURL(),
			)
		}
	})

	t.Run("GetUsers", func(t *testing.T) {
		data, err := users.Select("Id").Top(5).Get(context.Background())
		if err != nil {
			t.Error(err)
		}

		if len(data.Data()) == 0 {
			t.Error("can't get users")
		}

		if bytes.Compare(data, data.Normalized()) == -1 {
			t.Error("wrong response normalization")
		}
	})

	t.Run("GetUser", func(t *testing.T) {
		data, err := NewSP(spClient).Web().CurrentUser().Get(context.Background())
		if err != nil {
			t.Error(err)
		}
		user = data.Data()
	})

	t.Run("GetByID", func(t *testing.T) {
		if user.ID == 0 {
			t.Skip("no user ID to use in the test")
		}

		data, err := users.GetByID(user.ID).Select("Id").Get(context.Background())
		if err != nil {
			t.Error(err)
		}

		if data.Data().ID != user.ID {
			t.Errorf(
				"incorrect user ID, expected \"%d\", received \"%d\"",
				user.ID,
				data.Data().ID,
			)
		}
	})

	t.Run("GetByLoginName", func(t *testing.T) {
		if envCode == "2013" {
			t.Skip("is not supported with SP 2013")
		}
		if user.LoginName == "" {
			t.Skip("no user LoginName to use in the test")
		}

		data, err := users.GetByLoginName(user.LoginName).Select("LoginName").Get(context.Background())
		if err != nil {
			t.Error(err)
		}

		if data.Data().LoginName != user.LoginName {
			t.Errorf(
				"incorrect user LoginName, expected \"%s\", received \"%s\"",
				user.LoginName,
				data.Data().LoginName,
			)
		}
	})

	t.Run("GetByEmail", func(t *testing.T) {
		if user.Email == "" {
			t.Skip("no user Email to use in the test")
		}

		data, err := users.GetByEmail(user.Email).Select("Email").Get(context.Background())
		if err != nil {
			t.Error(err)
		}

		if data.Data().Email != user.Email {
			t.Errorf(
				"incorrect user Email, expected \"%s\", received \"%s\"",
				user.Email,
				data.Data().Email,
			)
		}
	})

}
