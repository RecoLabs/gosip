package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/koltyakov/gosip"
)

//go:generate ggen -ent Folder -conf -mods Select,Expand -helpers Data,Normalized

// Folder represents SharePoint Lists & Document Libraries Folder API queryable object struct
// Always use NewFolder constructor instead of &Folder{}
type Folder struct {
	client    *gosip.SPClient
	config    *RequestConfig
	endpoint  string
	modifiers *ODataMods
}

// FolderInfo - folder API response payload structure
type FolderInfo struct {
	Exists            bool      `json:"Exists"`
	IsWOPIEnabled     bool      `json:"IsWOPIEnabled"`
	ItemCount         int       `json:"ItemCount"`
	Name              string    `json:"Name"`
	ProgID            string    `json:"ProgID"`
	ServerRelativeURL string    `json:"ServerRelativeUrl"`
	TimeCreated       time.Time `json:"TimeCreated"`
	TimeLastModified  time.Time `json:"TimeLastModified"`
	UniqueID          string    `json:"UniqueId"`
	WelcomePage       string    `json:"WelcomePage"`
}

// FolderResp - folder response type with helper processor methods
type FolderResp []byte

// NewFolder ...
func NewFolder(client *gosip.SPClient, endpoint string, config *RequestConfig) *Folder {
	return &Folder{
		client:    client,
		endpoint:  endpoint,
		config:    config,
		modifiers: NewODataMods(),
	}
}

// ToURL gets endpoint with modificators raw URL
func (folder *Folder) ToURL() string {
	return toURL(folder.endpoint, folder.modifiers)
}

// Get gets this folder data object
func (folder *Folder) Get(ctx context.Context) (FolderResp, error) {
	client := NewHTTPClient(folder.client)
	return client.Get(ctx, folder.ToURL(), folder.config)
}

// Update updates Folder's metadata with properties provided in `body` parameter
// where `body` is byte array representation of JSON string payload relevant to SP.Folder object
func (folder *Folder) Update(ctx context.Context, body []byte) (FolderResp, error) {
	body = patchMetadataType(body, "SP.Folder")
	client := NewHTTPClient(folder.client)
	return client.Update(ctx, folder.endpoint, bytes.NewBuffer(body), folder.config)
}

// Delete deletes this folder (can't be restored from a recycle bin)
func (folder *Folder) Delete(ctx context.Context) error {
	client := NewHTTPClient(folder.client)
	_, err := client.Delete(ctx, folder.endpoint, folder.config)
	return err
}

// Recycle moves this folder to the recycle bin
func (folder *Folder) Recycle(ctx context.Context) error {
	client := NewHTTPClient(folder.client)
	endpoint := fmt.Sprintf("%s/Recycle", folder.endpoint)
	_, err := client.Post(ctx, endpoint, nil, folder.config)
	return err
}

// Folders gets sub folders queryable collection
func (folder *Folder) Folders() *Folders {
	return NewFolders(
		folder.client,
		fmt.Sprintf("%s/Folders", folder.endpoint),
		folder.config,
	)
}

// ParentFolder gets parent folder of this folder
func (folder *Folder) ParentFolder() *Folder {
	return NewFolder(
		folder.client,
		fmt.Sprintf("%s/ParentFolder", folder.endpoint),
		folder.config,
	)
}

// Props gets Properties API instance queryable collection for this Folder
func (folder *Folder) Props() *Properties {
	return NewProperties(
		folder.client,
		fmt.Sprintf("%s/Properties", folder.endpoint),
		folder.config,
		"folder",
	)
}

// Files gets files queryable collection in this folder
func (folder *Folder) Files() *Files {
	return NewFiles(
		folder.client,
		fmt.Sprintf("%s/Files", folder.endpoint),
		folder.config,
	)
}

// ListItemAllFields gets this folder Item data object metadata
func (folder *Folder) ListItemAllFields(ctx context.Context) (ListItemAllFieldsResp, error) {
	endpoint := fmt.Sprintf("%s/ListItemAllFields", folder.endpoint)
	apiURL, _ := url.Parse(endpoint)

	query := apiURL.Query()
	for k, v := range folder.modifiers.Get() {
		query.Set(k, TrimMultiline(v))
	}

	apiURL.RawQuery = query.Encode()
	client := NewHTTPClient(folder.client)

	data, err := client.Get(ctx, apiURL.String(), folder.config)
	if err != nil {
		return nil, err
	}
	data = NormalizeODataItem(data)
	return data, nil
}

// GetItem gets this folder Item API object metadata
func (folder *Folder) GetItem(ctx context.Context) (*Item, error) {
	scoped := NewFolder(folder.client, folder.endpoint, folder.config)
	data, err := scoped.Conf(HeadersPresets.Verbose).Select("Id").ListItemAllFields(ctx)
	if err != nil {
		return nil, err
	}

	res := &struct {
		Metadata struct {
			URI string `json:"uri"`
		} `json:"__metadata"`
	}{}

	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	item := NewItem(
		folder.client,
		res.Metadata.URI,
		folder.config,
	)
	return item, nil
}

// ContextInfo gets current Context Info object data
func (folder *Folder) ContextInfo(ctx context.Context) (*ContextInfo, error) {
	return NewContext(folder.client, folder.ToURL(), folder.config).Get(ctx)
}

// ToDo:
// StorageMetrics

// ensureFolder ensures folder existence by its server relative URL
// mode: modern (SP 2016 and newer), legacy (SP 2013)
func ensureFolder(ctx context.Context, web *Web, serverRelativeURL string, currentRelativeURL string, mode string) ([]byte, error) {
	headers := map[string]string{}
	for key, val := range getConfHeaders(web.config) {
		headers[key] = val
	}
	headers["X-Gosip-NoRetry"] = "true"
	headers["X-Gosip-NoHooks"] = "true"
	conf := &RequestConfig{
		Headers: headers,
	}

	getFolder := func(serverRelativeURL string) *Folder {
		serverRelativeURL = url.PathEscape(serverRelativeURL)
		return web.GetFolderByPath(serverRelativeURL)
	}
	if mode == "legacy" {
		// Get folder is used in
		getFolder = web.GetFolder
	}

	data, err := getFolder(currentRelativeURL).Conf(conf).Get(ctx)
	if err != nil {
		splitted := strings.Split(currentRelativeURL, "/")
		if len(splitted) == 1 {
			return nil, err
		}
		splitted = splitted[0 : len(splitted)-1]
		currentRelativeURL = strings.Join(splitted, "/")
		return ensureFolder(ctx, web, serverRelativeURL, currentRelativeURL, mode)
	}

	curFolders := strings.Split(currentRelativeURL, "/")
	expFolders := strings.Split(serverRelativeURL, "/")

	if len(curFolders) == len(expFolders) {
		return data, nil
	}

	createFolders := expFolders[len(curFolders):]
	for _, folder := range createFolders {
		data, err = getFolder(currentRelativeURL).Folders().Add(ctx, url.PathEscape(folder))
		if err != nil {
			return nil, err
		}
		currentRelativeURL += "/" + folder
	}

	return data, nil
}
