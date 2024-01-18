package api

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestView(t *testing.T) {
	checkClient(t)

	web := NewSP(spClient).Web()
	listURI := getRelativeURL(spClient.AuthCnfg.GetSiteURL()) + "/Shared%20Documents"

	t.Run("Get", func(t *testing.T) {
		data, err := web.GetList(listURI).Views().DefaultView().Get(context.Background())
		if err != nil {
			t.Error(err)
		}
		if data.Data().ID == "" {
			t.Error("can't unmarshal data")
		}
		if bytes.Compare(data, data.Normalized()) == -1 {
			t.Error("wrong response normalization")
		}
	})

	t.Run("SetViewXML", func(t *testing.T) {
		guid := uuid.New().String()
		meta := map[string]interface{}{
			"Title":        guid,
			"PersonalView": true,
		}
		data, _ := json.Marshal(meta)
		vr, err := web.GetList(listURI).Views().Add(context.Background(), data)
		if err != nil {
			t.Error(err)
		}
		if _, err := web.GetList(listURI).Views().GetByID(vr.Data().ID).
			SetViewXML(context.Background(), vr.Data().ListViewXML); err != nil {
			t.Error(err)
		}
		if err := web.GetList(listURI).Views().GetByID(vr.Data().ID).Delete(context.Background()); err != nil {
			t.Error(err)
		}
	})

	t.Run("UpdateDelete", func(t *testing.T) {
		guid := uuid.New().String()
		meta := map[string]interface{}{
			"Title":        guid,
			"PersonalView": true,
		}
		data, _ := json.Marshal(meta)
		vr, err := web.GetList(listURI).Views().Add(context.Background(), data)
		if err != nil {
			t.Error(err)
		}
		if _, err := web.GetList(listURI).Views().GetByID(vr.Data().ID).
			Update(context.Background(), []byte(`{"PersonalView":false}`)); err != nil {
			t.Error(err)
		}
		if err := web.GetList(listURI).Views().GetByID(vr.Data().ID).Delete(context.Background()); err != nil {
			t.Error(err)
		}
	})

}
