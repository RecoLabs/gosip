package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// AddValidateResp - add validate using path response type with helper processor methods
type AddValidateResp []byte

// ValidateAddOptions AddValidateUpdateItemUsingPath method options
type ValidateAddOptions struct {
	DecodedPath       string
	NewDocumentUpdate bool
	CheckInComment    string
}

// AddValidateFieldResult field result struct
type AddValidateFieldResult struct {
	ErrorCode    int
	ErrorMessage string
	FieldName    string
	FieldValue   string
	HasException bool
	ItemID       int `json:"ItemId"`
}

// AddValidate adds new item in this list using AddValidateUpdateItemUsingPath method.
// formValues fingerprints https://github.com/koltyakov/sp-sig-20180705-demo/blob/master/src/03-pnp/FieldTypes.md#field-data-types-fingerprints-sample
func (items *Items) AddValidate(ctx context.Context, formValues map[string]string, options *ValidateAddOptions) (AddValidateResp, error) {
	endpoint := fmt.Sprintf("%s/AddValidateUpdateItemUsingPath()", getPriorEndpoint(items.endpoint, "/items"))
	client := NewHTTPClient(items.client)
	type formValue struct {
		FieldName  string `json:"FieldName"`
		FieldValue string `json:"FieldValue"`
	}
	var formValuesArray []*formValue
	for n, v := range formValues {
		formValuesArray = append(formValuesArray, &formValue{
			FieldName:  n,
			FieldValue: v,
		})
	}
	payload := map[string]interface{}{"formValues": formValuesArray}
	if options != nil {
		payload["bNewDocumentUpdate"] = options.NewDocumentUpdate
		payload["checkInComment"] = options.CheckInComment
		if options.DecodedPath != "" {
			payload["listItemCreateInfo"] = map[string]interface{}{
				"__metadata": map[string]string{"type": "SP.ListItemCreationInformationUsingPath"},
				"FolderPath": map[string]interface{}{
					"__metadata": map[string]string{"type": "SP.ResourcePath"},
					"DecodedUrl": checkGetRelativeURL(options.DecodedPath, items.endpoint),
				},
			}
		}
	}
	body, _ := json.Marshal(payload)

	var res AddValidateResp
	var err error

	res, err = client.Post(ctx, endpoint, bytes.NewBuffer(body), items.config)
	if err != nil {
		return res, err
	}

	var errs []error
	for _, f := range res.Data() {
		if f.HasException {
			errs = append(errs, fmt.Errorf("%s: %s", f.FieldName, f.ErrorMessage))
		}
	}
	if len(errs) > 0 {
		return res, fmt.Errorf("%v", errs)
	}

	return res, nil
}

/* AddValidate response helpers */

// Data unmarshals AddValidate response
func (avResp *AddValidateResp) Data() []AddValidateFieldResult {
	var d []AddValidateFieldResult
	r := &struct {
		D struct {
			AddValidateUpdateItemUsingPath struct {
				Results []AddValidateFieldResult `json:"results"`
			} `json:"AddValidateUpdateItemUsingPath"`
		} `json:"d"`
		Value []AddValidateFieldResult `json:"value"`
	}{}
	_ = json.Unmarshal(*avResp, &r)
	if r.Value != nil {
		return r.Value
	}
	if r.D.AddValidateUpdateItemUsingPath.Results != nil {
		return r.D.AddValidateUpdateItemUsingPath.Results
	}
	return d
}

// Value gets created item's value from the response
func (avResp *AddValidateResp) Value(fieldName string) string {
	dd := avResp.Data()
	for _, d := range dd {
		if d.FieldName == fieldName {
			return d.FieldValue
		}
	}
	return ""
}

// ID gets created item's ID from the response
func (avResp *AddValidateResp) ID() int {
	v := avResp.Value("Id")
	d, _ := strconv.Atoi(v)
	return d
}
