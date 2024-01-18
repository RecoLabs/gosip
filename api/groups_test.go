package api

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestGroups(t *testing.T) {
	checkClient(t)

	sp := NewSP(spClient)
	groups := sp.Web().SiteGroups()
	endpoint := spClient.AuthCnfg.GetSiteURL() + "/_api/Web/SiteGroups"
	newGroupName := uuid.New().String()
	newGroupNameRemove := uuid.New().String()

	t.Run("Constructor", func(t *testing.T) {
		g := NewGroups(spClient, endpoint, nil)
		if _, err := g.Select("Id").Get(context.Background()); err != nil {
			t.Error(err)
		}
	})

	t.Run("ToURL", func(t *testing.T) {
		if groups.ToURL() != endpoint {
			t.Errorf(
				"incorrect endpoint URL, expected \"%s\", received \"%s\"",
				endpoint,
				groups.ToURL(),
			)
		}
	})

	t.Run("GetGroups", func(t *testing.T) {
		data, err := groups.Select("Id").Top(5).Get(context.Background())
		if err != nil {
			t.Error(err)
		}
		if len(data.Data()) == 0 {
			t.Error("can't get groups")
		}
	})

	t.Run("GetGroup", func(t *testing.T) {
		data, err := groups.Select("Id,Title").Top(1).Get(context.Background())
		if err != nil {
			t.Error(err)
		}
		if len(data.Data()) == 0 {
			t.Error("can't get groups")
		}
		if data.Data()[0].Data().Title == "" {
			t.Error("can't get groups")
		}
		if bytes.Compare(data, data.Normalized()) == -1 {
			t.Error("response normalization error")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		data0, err := groups.Select("Id,Title").Top(1).Get(context.Background())
		if err != nil {
			t.Error(err)
		}

		groupID := data0.Data()[0].Data().ID
		data, err := groups.GetByID(groupID).Select("Id").Get(context.Background())
		if err != nil {
			t.Error(err)
		}

		if groupID != data.Data().ID {
			t.Errorf(
				"incorrect group ID, expected \"%d\", received \"%d\"",
				groupID,
				data.Data().ID,
			)
		}
	})

	t.Run("GetByName", func(t *testing.T) {
		data0, err := groups.Select("Id,Title").Top(1).Get(context.Background())
		if err != nil {
			t.Error(err)
		}

		groupTitle := data0.Data()[0].Data().Title
		data, err := groups.GetByName(groupTitle).Select("Title").Get(context.Background())
		if err != nil {
			t.Error(err)
		}

		if groupTitle != data.Data().Title {
			t.Errorf(
				"incorrect group Title, expected \"%s\", received \"%s\"",
				groupTitle,
				data.Data().Title,
			)
		}
	})

	t.Run("Add", func(t *testing.T) {
		if _, err := groups.Conf(headers.verbose).Add(context.Background(), newGroupName, nil); err != nil {
			t.Error(err)
		}

		u, err := sp.Web().CurrentUser().Select("Id").Get(context.Background())
		if err != nil {
			t.Error(err)
		}

		if envCode != "spo" {
			t.Skip("is not supported with legacy SP")
		}
		if err := groups.GetByName(newGroupName).SetOwner(context.Background(), u.Data().ID); err != nil {
			t.Error(err)
		}
	})

	t.Run("RemoveByLoginName", func(t *testing.T) {
		if _, err := groups.Conf(headers.verbose).Add(context.Background(), newGroupNameRemove, nil); err != nil {
			t.Error(err)
		}
		if err := groups.RemoveByLoginName(context.Background(), newGroupNameRemove); err != nil {
			t.Error(err)
		}
	})

	if err := groups.RemoveByLoginName(context.Background(), newGroupName); err != nil {
		t.Error(err)
	}

}
