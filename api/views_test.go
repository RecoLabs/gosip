package api

import (
	"bytes"
	"context"
	"testing"
)

func TestViews(t *testing.T) {
	checkClient(t)

	web := NewSP(spClient).Web()
	listURI := getRelativeURL(spClient.AuthCnfg.GetSiteURL()) + "/Shared%20Documents"
	view, err := getAnyView()
	if err != nil {
		t.Error(err)
	}

	t.Run("Get", func(t *testing.T) {
		data, err := web.GetList(listURI).Views().Get(context.Background())
		if err != nil {
			t.Error(err)
		}
		if data.Data()[0].Data().ID == "" {
			t.Error("can't unmarshal data")
		}
		if bytes.Compare(data, data.Normalized()) == -1 {
			t.Error("wrong response normalization")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		if _, err := web.GetList(listURI).Views().GetByID(view.ID).Get(context.Background()); err != nil {
			t.Error(err)
		}
	})

	t.Run("DefaultView", func(t *testing.T) {
		if _, err := web.GetList(listURI).Views().DefaultView().Get(context.Background()); err != nil {
			t.Error(err)
		}
	})

	t.Run("GetByTitle", func(t *testing.T) {
		if _, err := web.GetList(listURI).Views().GetByTitle(view.Title).Get(context.Background()); err != nil {
			t.Error(err)
		}
	})

	// ToDo:
	// Add

}

func getAnyView() (*ViewInfo, error) {
	web := NewSP(spClient).Web()
	listURI := getRelativeURL(spClient.AuthCnfg.GetSiteURL()) + "/Shared%20Documents"
	data, err := web.GetList(listURI).Views().Top(1).Get(context.Background())
	if err != nil {
		return nil, err
	}
	return data.Data()[0].Data(), nil
}
