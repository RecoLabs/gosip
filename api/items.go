package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/koltyakov/gosip"
)

// Items represent SharePoint Lists & Document Libraries Items API queryable collection struct
// Always use NewItems constructor instead of &Items{}
type Items struct {
	client    *gosip.SPClient
	config    *RequestConfig
	endpoint  string
	modifiers *ODataMods
}

// ItemsResp - items response type with helper processor methods
type ItemsResp []byte

// ItemsPage - paged items
type ItemsPage struct {
	Items       ItemsResp
	HasNextPage func() bool
	GetNextPage func() (*ItemsPage, error)
}

// NewItems - Items struct constructor function
func NewItems(client *gosip.SPClient, endpoint string, config *RequestConfig) *Items {
	return &Items{
		client:    client,
		endpoint:  endpoint,
		config:    config,
		modifiers: NewODataMods(),
	}
}

// ToURL gets endpoint with modificators raw URL
func (items *Items) ToURL() string {
	return toURL(items.endpoint, items.modifiers)
}

// Conf receives custom request config definition, e.g. custom headers, custom OData mod
func (items *Items) Conf(config *RequestConfig) *Items {
	items.config = config
	return items
}

// Select adds $select OData modifier
func (items *Items) Select(oDataSelect string) *Items {
	items.modifiers.AddSelect(oDataSelect)
	return items
}

// Expand adds $expand OData modifier
func (items *Items) Expand(oDataExpand string) *Items {
	items.modifiers.AddExpand(oDataExpand)
	return items
}

// Filter adds $filter OData modifier
func (items *Items) Filter(oDataFilter string) *Items {
	items.modifiers.AddFilter(oDataFilter)
	return items
}

// Top adds $top OData modifier
func (items *Items) Top(oDataTop int) *Items {
	items.modifiers.AddTop(oDataTop)
	return items
}

// Skip adds $skiptoken OData modifier
func (items *Items) Skip(skipToken string) *Items {
	items.modifiers.AddSkip(skipToken)
	return items
}

// OrderBy adds $orderby OData modifier
func (items *Items) OrderBy(oDataOrderBy string, ascending bool) *Items {
	items.modifiers.AddOrderBy(oDataOrderBy, ascending)
	return items
}

// Get gets Items API queryable collection
func (items *Items) Get() (ItemsResp, error) {
	sp := NewHTTPClient(items.client)
	data, err := sp.Get(items.ToURL(), getConfHeaders(items.config))
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GetPaged gets Paged Items collection
func (items *Items) GetPaged() (*ItemsPage, error) {
	data, err := items.Get()
	if err != nil {
		return nil, err
	}
	res := &ItemsPage{
		Items: data,
		HasNextPage: func() bool {
			return data.HasNextPage()
		},
		GetNextPage: func() (*ItemsPage, error) {
			nextURL := data.NextPageURL()
			if nextURL == "" {
				return nil, fmt.Errorf("unable to get next page")
			}
			return NewItems(items.client, nextURL, items.config).GetPaged()
		},
	}
	return res, nil
}

// GetAll gets all items in a list using internal page helper. The use case of the method is getting all the content from large lists.
// Method ignores custom sorting and filtering as not supported for the large lists due to throttling limitations.
func (items *Items) GetAll() ([]ItemResp, error) {
	return getAll(nil, nil, items)
}

func getAll(res []ItemResp, cur ItemsResp, items *Items) ([]ItemResp, error) {
	if res == nil && cur == nil {
		itemsCopy := NewItems(items.client, items.endpoint, items.config)
		for key, val := range items.modifiers.Get() {
			switch key {
			case "$select":
				itemsCopy.modifiers.AddSelect(val)
			case "$expand":
				itemsCopy.modifiers.AddExpand(val)
			case "$top":
				top, _ := strconv.Atoi(val)
				itemsCopy.modifiers.AddTop(top)
			}
		}
		itemsResp, err := itemsCopy.Get()
		if err != nil {
			return nil, err
		}
		res = itemsResp.Data()
		cur = itemsResp
	}
	nextURL := getODataCollectionNextPageURL(cur)
	if nextURL == "" {
		return res, nil
	}
	nextItemsResp, err := NewItems(items.client, nextURL, items.config).Get()
	if err != nil {
		return res, err
	}
	res = append(res, nextItemsResp.Data()...)
	return getAll(res, nextItemsResp, items)
}

// Add adds new item in this list. `body` parameter is byte array representation of JSON string payload relevalt to item metadata object.
func (items *Items) Add(body []byte) (ItemResp, error) {
	body = patchMetadataTypeCB(body, func() string {
		endpoint := getPriorEndpoint(items.endpoint, "/Items")
		list := NewList(items.client, endpoint, nil)
		oDataType, _ := list.GetEntityType() // ToDo: add caching for Entity Types
		return oDataType
	})
	sp := NewHTTPClient(items.client)
	return sp.Post(items.endpoint, body, getConfHeaders(items.config))
}

// GetByID gets item data object by its ID
func (items *Items) GetByID(itemID int) *Item {
	return NewItem(
		items.client,
		fmt.Sprintf("%s(%d)", items.endpoint, itemID),
		items.config,
	)
}

// GetByCAML gets items data using CAML query
func (items *Items) GetByCAML(caml string) (ItemsResp, error) {
	endpoint := fmt.Sprintf("%s/GetItems", strings.TrimRight(items.endpoint, "/Items"))
	apiURL, _ := url.Parse(endpoint)
	query := url.Values{}
	for k, v := range items.modifiers.Get() {
		query.Add(k, TrimMultiline(v))
	}
	apiURL.RawQuery = query.Encode()

	request := &struct {
		Query struct {
			Metadata struct {
				Type string `json:"type"`
			} `json:"__metadata"`
			ViewXML string `json:"ViewXml"`
		} `json:"query"`
	}{}

	request.Query.Metadata.Type = "SP.CamlQuery"
	request.Query.ViewXML = TrimMultiline(caml)

	body, _ := json.Marshal(request)

	sp := NewHTTPClient(items.client)
	return sp.Post(apiURL.String(), []byte(body), getConfHeaders(items.config))
}

// ToDo:
// Batch

/* Response helpers */

// Data : to get typed data
func (itemsResp *ItemsResp) Data() []ItemResp {
	collection, _ := parseODataCollection(*itemsResp)
	items := []ItemResp{}
	for _, item := range collection {
		items = append(items, ItemResp(item))
	}
	return items
}

// NextPageURL : gets next page OData collection
func (itemsResp *ItemsResp) NextPageURL() string {
	return getODataCollectionNextPageURL(*itemsResp)
}

// HasNextPage : returns is true if next page exists
func (itemsResp *ItemsResp) HasNextPage() bool {
	return itemsResp.NextPageURL() != ""
}

// // GetNext : gets next page OData collection
// func (itemsResp *ItemsResp) GetNext(items *Items) (ItemsResp, error) {
// 	nextURL := getODataCollectionNextPageURL(*itemsResp)
// 	if nextURL == "" {
// 		return nil, fmt.Errorf("unable to get next page")
// 	}
// 	return NewItems(items.client, nextURL, items.config).Get()
// }

// Unmarshal : to unmarshal to custom object
func (itemsResp *ItemsResp) Unmarshal(obj interface{}) error {
	// collection := parseODataCollection(*itemsResp)
	// data, _ := json.Marshal(collection)
	data, _ := NormalizeODataCollection(*itemsResp)
	return json.Unmarshal(data, obj)
}
