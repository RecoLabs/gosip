package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRecords(t *testing.T) {
	checkClient(t)

	if envCode == "2013" {
		t.Skip("is not supported with SP 2013")
	}

	sp := NewSP(spClient)

	// Activate in place record management feature
	if err := sp.Site().Features().Add(context.Background(), "da2e115b-07e4-49d9-bb2c-35e93bb9fca9", true); err != nil {
		t.Error(err)
	}

	folder := sp.Web().GetFolder("Shared Documents")

	folderName := uuid.New().String()
	docs := []string{
		fmt.Sprintf("%s.txt", uuid.New().String()),
		fmt.Sprintf("%s.txt", uuid.New().String()),
	}

	if _, err := folder.Folders().Add(context.Background(), folderName); err != nil {
		t.Error(err)
	}

	for _, doc := range docs {
		if _, err := folder.Folders().GetByName(folderName).Files().Add(context.Background(), doc, []byte(doc), true); err != nil {
			t.Error(err)
		}
	}

	t.Run("Records/Declare", func(t *testing.T) {
		item, err := folder.Folders().GetByName(folderName).Files().GetByName(docs[0]).GetItem(context.Background())
		if err != nil {
			t.Error(err)
		}
		if err := item.Records().Declare(context.Background()); err != nil {
			t.Error(err)
		}
	})

	t.Run("Records/IsRecord", func(t *testing.T) {
		item, err := folder.Folders().GetByName(folderName).Files().GetByName(docs[0]).GetItem(context.Background())
		if err != nil {
			t.Error(err)
		}
		if _, err := item.Records().IsRecord(context.Background()); err != nil {
			t.Error(err)
		}
	})

	t.Run("Records/Undeclare", func(t *testing.T) {
		item, err := folder.Folders().GetByName(folderName).Files().GetByName(docs[0]).GetItem(context.Background())
		if err != nil {
			t.Error(err)
		}
		if err := item.Records().Undeclare(context.Background()); err != nil {
			t.Error(err)
		}
		recDate, err := item.Records().RecordDate(context.Background())
		if err != nil {
			t.Error(err)
		}
		if recDate.String() != "0001-01-01 00:00:00 +0000 UTC" {
			t.Error("something wrong")
		}
	})

	t.Run("Records/DeclareWithDate", func(t *testing.T) {
		if envCode != "spo" {
			t.Skip("is not supported with old SharePoint versions")
		}

		item, err := folder.Folders().GetByName(folderName).Files().GetByName(docs[1]).GetItem(context.Background())
		if err != nil {
			t.Error(err)
		}
		strTime := "2019-01-01T08:00:00.000Z"
		date, _ := time.Parse(time.RFC3339, strTime)
		if err := item.Records().DeclareWithDate(context.Background(), date); err != nil {
			t.Error(err)
		}
		recDate, err := item.Records().RecordDate(context.Background())
		if err != nil {
			t.Error(err)
		}
		if date != recDate {
			t.Error("wrong record date")
		}
		// Undeclare to delete after tests
		if err := item.Records().Undeclare(context.Background()); err != nil {
			t.Error(err)
		}
	})

	if err := folder.Folders().GetByName(folderName).Delete(context.Background()); err != nil {
		t.Error(err)
	}

}
