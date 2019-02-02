package scenarios

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"runtime"
	"time"

	"github.com/koltyakov/gosip/auth/onlineaddinonly"
	"github.com/koltyakov/gosip/cnfg"
)

// GetAuthTest : get auth test scenario
func GetAuthTest() {
	startAtProc := time.Now()
	startAt := time.Now()

	_, filename, _, _ := runtime.Caller(1)
	configPath := path.Join(path.Dir(filename), "../../config/private.addinonly.json")
	config, err := cnfg.InitAuthConfigAddinOnly(configPath, "")
	if err != nil {
		fmt.Printf("unable to get config: %v", err)
		return
	}
	fmt.Printf("config: %v\n", config)
	fmt.Printf("config read in, sec: %f\n", time.Since(startAt).Seconds())
	startAt = time.Now()

	authToken, err := onlineaddinonly.GetAuth(config)
	if err != nil {
		fmt.Printf("unable to get auth: %v", err)
		return
	}
	// fmt.Printf("auth token: %s\n", authToken)
	fmt.Printf("authenticated in, sec: %f\n", time.Since(startAt).Seconds())
	startAt = time.Now()

	///
	authToken, err = onlineaddinonly.GetAuth(config)
	if err != nil {
		fmt.Printf("unable to get auth: %v", err)
		return
	}
	// fmt.Printf("auth token: %s\n", authToken)
	fmt.Printf("second auth in, sec: %f\n", time.Since(startAt).Seconds())
	startAt = time.Now()
	///

	apiEndpoint := config.SiteURL + "/_api/web?$select=Title"
	req, err := http.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		fmt.Printf("unable to create a request: %v", err)
		return
	}

	req.Header.Set("Accept", "application/json;odata=minimalmetadata")
	// req.Header.Set("Content-Type", "application/json;odata=minimalmetadata")
	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unable to request api: %v", err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("unable to read a response: %v", err)
		return
	}

	type apiResponse struct {
		Title string `json:"Title"`
	}

	results := &apiResponse{}

	err = json.Unmarshal(data, &results)
	if err != nil {
		fmt.Printf("unable to parse a response: %v", err)
		return
	}

	fmt.Println("=== Response from API ===")
	fmt.Printf("Web title: %v\n", results.Title)
	fmt.Printf("api requested in, sec: %f\n", time.Since(startAt).Seconds())
	fmt.Printf("summary time, sec: %f\n", time.Since(startAtProc).Seconds())

}
