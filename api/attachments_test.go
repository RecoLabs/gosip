package api

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestAttachments(t *testing.T) {
	checkClient(t)

	web := NewSP(spClient).Web()
	listTitle := uuid.New().String()

	if _, err := web.Lists().Add(context.Background(), listTitle, nil); err != nil {
		t.Error(err)
	}
	list := web.Lists().GetByTitle(listTitle)
	item, err := list.Items().Add(context.Background(), []byte(`{"Title":"Attachment test"}`))
	if err != nil {
		t.Error(err)
	}

	t.Run("Add", func(t *testing.T) {
		attachments := map[string][]byte{
			"att_01.txt": []byte("attachment 01"),
			"att_02.txt": []byte("attachment 02"),
			"att_03.txt": []byte("attachment 03"),
			"att_04.txt": []byte("attachment 04"),
		}
		for fileName, content := range attachments {
			if _, err := list.Items().GetByID(item.Data().ID).Attachments().Add(context.Background(), fileName, bytes.NewBuffer(content)); err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		data, err := list.Items().GetByID(item.Data().ID).Attachments().Get(context.Background())
		if err != nil {
			t.Error(err)
		}
		if len(data.Data()) != 4 {
			t.Error("wrong number of attachments")
		}
		if bytes.Compare(data, data.Normalized()) == -1 {
			t.Error("response normalization error")
		}
	})

	t.Run("GetByName", func(t *testing.T) {
		data, err := list.Items().GetByID(item.Data().ID).Attachments().GetByName("att_01.txt").Get(context.Background())
		if err != nil {
			t.Error(err)
		}
		if data.Data().FileName != "att_01.txt" {
			t.Error("wrong attachment name")
		}
		if bytes.Compare(data, data.Normalized()) == -1 {
			t.Error("response normalization error")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := list.Items().GetByID(item.Data().ID).Attachments().GetByName("att_02.txt").Delete(context.Background()); err != nil {
			t.Error(err)
		}
	})

	t.Run("Recycle", func(t *testing.T) {
		if envCode == "2013" {
			t.Skip("is not supported with SP 2013")
		}

		if err := list.Items().GetByID(item.Data().ID).Attachments().GetByName("att_03.txt").Recycle(context.Background()); err != nil {
			t.Error(err)
		}
	})

	t.Run("Download", func(t *testing.T) {
		expectedContent := []byte("attachment 04")
		content, err := list.Items().GetByID(item.Data().ID).Attachments().GetByName("att_04.txt").Download(context.Background())
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(content, expectedContent) {
			t.Error("wrong attachment content")
		}
	})

	if err := list.Delete(context.Background()); err != nil {
		t.Error(err)
	}

}
