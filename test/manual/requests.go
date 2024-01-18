package manual

import (
	"context"
	"fmt"
	"strings"

	"github.com/recolabs/gosip"
	"github.com/recolabs/gosip/api"
)

// CheckBasicPost : try creating an item
// noinspection GoUnusedExportedFunction
func CheckBasicPost(ctx context.Context, client *gosip.SPClient) (string, error) {
	sp := api.NewHTTPClient(client)
	endpoint := client.AuthCnfg.GetSiteURL() + "/_api/web/lists/getByTitle('Custom')/items"
	body := `{"__metadata":{"type":"SP.Data.CustomListItem"},"Title":"Test"}`

	data, err := sp.Post(ctx, endpoint, strings.NewReader(body), nil)
	if err != nil {
		return "", fmt.Errorf("unable to read a response: %w", err)
	}

	return string(data), nil
}
