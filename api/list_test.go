package api

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestList(t *testing.T) {
	checkClient(t)

	web := NewSP(spClient).Web()
	listTitle := strings.Replace(uuid.New().String(), "-", "", -1)
	listInfo, err := web.Lists().Add(context.Background(), listTitle, nil)
	if err != nil {
		t.Error(err)
	}
	list := web.Lists().GetByID(listInfo.Data().ID)
	defer func() { _ = list.Delete(context.Background()) }()

	t.Run("GetEntityType", func(t *testing.T) {
		entType, err := list.GetEntityType(context.Background())
		if err != nil {
			t.Error(err)
		}
		if entType == "" {
			t.Error("can't get entity type")
		}
	})

	t.Run("Get", func(t *testing.T) {
		l, err := list.Get(context.Background()) // .Select("*")
		if err != nil {
			t.Error(err)
		}
		if l.Data().Title == "" {
			t.Error("can't unmarshal list info")
		}
		if bytes.Compare(l, l.Normalized()) == -1 {
			t.Error("wrong response normalization")
		}
	})

	t.Run("Items", func(t *testing.T) {
		if _, err := list.Items().Select("Id").Top(1).Get(context.Background()); err != nil {
			t.Error(err)
		}
	})

	t.Run("ParentWeb", func(t *testing.T) {
		if _, err := list.ParentWeb().Get(context.Background()); err != nil {
			t.Error(err)
		}
	})

	t.Run("ReserveListItemID", func(t *testing.T) {
		nextID, err := list.ReserveListItemID(context.Background())
		if err != nil {
			t.Error(err)
		}
		if nextID == 0 {
			t.Error("can't reserve list item ID")
		}
	})

	t.Run("RenderListData", func(t *testing.T) {
		listData, err := list.RenderListData(context.Background(), `<View></View>`)
		if err != nil {
			t.Error(err)
		}
		if listData.Data().FolderPermissions == "" {
			t.Error("incorrect data")
		}
	})

	t.Run("RootFolder", func(t *testing.T) {
		if _, err := list.RootFolder().Get(context.Background()); err != nil {
			t.Error(err)
		}
	})

	t.Run("Recycle", func(t *testing.T) {
		guid := uuid.New().String()
		lr, err := web.Lists().Add(context.Background(), guid, nil)
		if err != nil {
			t.Error(err)
		}
		if err := web.Lists().GetByID(lr.Data().ID).Recycle(context.Background()); err != nil {
			t.Error(err)
		}
	})

}
