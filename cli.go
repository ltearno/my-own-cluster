package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ltearno/my-own-cluster/assetsgen"
	"github.com/ltearno/my-own-cluster/tools"
)

type RegisterBlobWithNameRequest struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Bytes       string `json:"bytes"`
}

type RegisterBlobWithNameResponse struct {
	Status bool   `json:"status"`
	TechID string `json:"tech_id"`
}

type RegisterBlobRequest struct {
	ContentType string `json:"content_type"`
	Bytes       string `json:"bytes"`
}

type RegisterBlobResponse struct {
	Status bool   `json:"status"`
	TechID string `json:"tech_id"`
}

type PlugFunctionRequest struct {
	Method        string            `json:"method"`
	Path          string            `json:"path"`
	Name          string            `json:"name"`
	Tags          map[string]string `json:"tags,omitempty"`
	StartFunction string            `json:"start_function"`
	Data          string            `json:"data"`
}

type PlugFilterRequest struct {
	Name          string `json:"name"`
	StartFunction string `json:"start_function"`
	Data          string `json:"data"`
}

type PlugFileRequest struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Name   string            `json:"name"`
	Tags   map[string]string `json:"tags,omitempty"`
}

type PlugResponse struct {
	Status bool `json:"status"`
}

type PlugFilterResponse struct {
	Status   bool   `json:"status"`
	FilterID string `json:"filter_id"`
}

type UnplugFunctionRequest struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type UnplugResponse struct {
	Status bool `json:"status"`
}

// for reusing underlying tcp connection accross multiple calls,
// must be used by one thread only
var client = &http.Client{Transport: &http.Transport{
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
}}

func registerBlobWithName(baseURL string, name string, contentType string, fileName string) (string, error) {
	contentBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("cannot read file '%s'", fileName)
	}

	encodedBytes := base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(contentBytes)

	reqBody := &RegisterBlobWithNameRequest{
		Name:        name,
		ContentType: contentType,
		Bytes:       encodedBytes,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("cannot marshal json")
	}

	bodyReader := bytes.NewReader(bodyBytes)

	resp, err := client.Post(baseURL+"/api/blob/register", "application/json", bodyReader)
	if err != nil {
		panic(err)
	}

	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	response := &RegisterBlobWithNameResponse{}
	if json.Unmarshal(bytes, response) != nil {
		return "", fmt.Errorf("cannot unmarshall server response : %s", string(bytes))
	}

	if response.Status {
		fmt.Printf("[%s] registered blob '%s' content_type:%s size:%d techID:%s\n", baseURL, name, contentType, len(contentBytes), response.TechID)
	} else {
		return "", fmt.Errorf("[%s] ERROR while registration of '%s' content_type:%s techID:%s", baseURL, name, contentType, response.TechID)
	}

	return response.TechID, nil
}

func registerBlob(baseURL string, contentType string, fileName string) (string, error) {
	contentBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("cannot read file '%s'", fileName)
	}

	encodedBytes := base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(contentBytes)

	reqBody := &RegisterBlobRequest{
		ContentType: contentType,
		Bytes:       encodedBytes,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("cannot marshal json (%v)", err)
	}

	bodyReader := bytes.NewReader(bodyBytes)

	url := baseURL + "/api/blob/register"
	resp, err := client.Post(url, "application/json", bodyReader)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%s status when calling '%s'", resp.Status, url)
	}

	bytes, _ := ioutil.ReadAll(resp.Body)

	response := &RegisterBlobResponse{}
	err = json.Unmarshal(bytes, response)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshall server response (%v), here is the raw response's status: %d and text: %s", err, resp.StatusCode, string(bytes))
	}

	if response.Status {
		fmt.Printf("registered blob content_type:%s size:%d techID:%s\n", contentType, len(contentBytes), response.TechID)
	} else {
		return "", fmt.Errorf("ERROR while registration of content_type:%s techID:%s", contentType, response.TechID)
	}

	return response.TechID, nil
}

func CliPushFunction(verbs []Verb) {
	// skip the verb that triggered us, it is given to us in case it contains options
	baseURL := getAPIBaseURL(verbs[0])
	verbs = verbs[1:]

	functionName := verbs[0].Name
	wasmFileName := verbs[1].Name

	codeType := detectContentTypeFromFileName(wasmFileName)

	registerBlobWithName(baseURL, functionName, codeType, wasmFileName)
}

func CliUploadFile(verbs []Verb) {
	baseURL := getAPIBaseURL(verbs[0])
	method := verbs[0].GetOptionOr("method", "get")
	tagsJSON := verbs[0].GetOptionOr("tags", "{}")
	tags := make(map[string]string)
	err := json.Unmarshal([]byte(tagsJSON), &tags)
	if err != nil {
		fmt.Printf("cannot marshal json tags (%v)\n", err)
		return
	}
	verbs = verbs[1:]

	path := verbs[0].Name
	fileName := verbs[1].Name
	contentType := ""
	if len(verbs) >= 3 {
		contentType = verbs[2].Name
	} else {
		contentType = detectContentTypeFromFileName(fileName)
	}

	techID, err := uploadFile(baseURL, method, path, contentType, fileName, tags)
	if err != nil {
		fmt.Printf("error while uploading file, %v\n", err)
		return
	}

	fmt.Printf("registered file '%s' '%s' content_type:%s techID:%s\n", fileName, path, contentType, techID)
}

func CliUploadDir(verbs []Verb) {
	baseURL := getAPIBaseURL(verbs[0])
	method := verbs[0].GetOptionOr("method", "get")
	tagsJSON := verbs[0].GetOptionOr("tags", "{}")
	tags := make(map[string]string)
	err := json.Unmarshal([]byte(tagsJSON), &tags)
	if err != nil {
		fmt.Printf("cannot marshal json tags (%v)\n", err)
		return
	}
	verbs = verbs[1:]

	pathPrefix := verbs[0].Name
	directoryName := verbs[1].Name

	directoryName, err = filepath.Abs(directoryName)
	if err != nil {
		fmt.Printf("error getting absolute path (%v)\n", err)
		return
	}

	count := 0
	countError := 0

	filepath.Walk(directoryName, func(path string, info os.FileInfo, err error) error {
		urlPath := filepath.Join(pathPrefix, path[len(directoryName):])
		if !info.IsDir() {
			fmt.Printf("%s  =>  %s\n", path, urlPath)
			_, err := uploadFile(baseURL, method, urlPath, detectContentTypeFromFileName(path), path, tags)
			if err != nil {
				fmt.Printf(" error while uploading %s to %s (%v)\n", path, urlPath, err)
				countError++
			}
			count++
		}

		return nil
	})

	fmt.Printf("uploaded %d files (%d errors).\n", count, countError)
}

var bindingsExtensions map[string]string = map[string]string{
	"c":      ".h",
	"c-syms": ".syms",
	"ts":     ".d.ts",
	"rust":   ".rs",
}

func PrintBindings(module string, language string) {
	extension, ok := bindingsExtensions[language]
	if !ok {
		panic("unknown language for bindings, you can contribute to language bindings at https://github.com/ltearno/my-own-cluster")
	}

	resourceName := fmt.Sprintf("assets/%s-api-guest%s", module, extension)
	if language == "rust" {
		resourceName = strings.ReplaceAll(resourceName, "-", "_")
	}
	cGuestBindingCode, err := assetsgen.Asset(resourceName)
	if err != nil {
		panic("library bindings not found, you can contribute to language bindings at https://github.com/ltearno/my-own-cluster")
	}

	fmt.Print(string(cGuestBindingCode))
}

func CliGuestApi(verbs []Verb) {
	verbs = verbs[1:]

	module := verbs[0].Name
	language := verbs[1].Name

	// will panic if not handled
	PrintBindings(module, language)
}

func CliVersion(verbs []Verb) {
	serverBaseUrl := getAPIBaseURL(verbs[0])

	fmt.Println("my-own-cluster")
	fmt.Println("Version: " + GetVersion())
	fmt.Println("Remote: " + serverBaseUrl)
}

func CliExportDatabase(verbs []Verb) {
	baseURL := getAPIBaseURL(verbs[0])

	resp, err := client.Get(baseURL + "/api/admin/export-database")
	if err != nil {
		panic(err)
	}

	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Println(string(bytes))
}

func CliRemote(verbs []Verb) {
	fmt.Println(getAPIBaseURL(verbs[0]))
}

func CliPlugFunction(verbs []Verb) {
	serverBaseURL := getAPIBaseURL(verbs[0])
	method := verbs[0].GetOptionOr("method", "get")
	tagsJSON := verbs[0].GetOptionOr("tags", "{}")
	tags := make(map[string]string)
	err := json.Unmarshal([]byte(tagsJSON), &tags)
	if err != nil {
		fmt.Printf("cannot marshal json tags (%v)\n", err)
		return
	}
	verbs = verbs[1:]

	path := verbs[0].Name
	name := verbs[1].Name
	startFunction := ""
	if len(verbs) >= 3 {
		startFunction = verbs[2].Name
	}
	data := ""
	if len(verbs) >= 4 {
		data = verbs[3].Name
	}

	reqBody := &PlugFunctionRequest{
		Method:        method,
		Path:          path,
		Name:          name,
		StartFunction: startFunction,
		Data:          data,
		Tags:          tags,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("cannot marshal json (%v)\n", err)
		return
	}

	bodyReader := bytes.NewReader(bodyBytes)

	resp, err := client.Post(serverBaseURL+"/api/function/plug", "application/json", bodyReader)
	if err != nil {
		fmt.Printf("error during http request (%v)\n", err)
		return
	}

	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	response := &PlugResponse{}
	if json.Unmarshal(bytes, response) != nil {
		fmt.Printf("cannot unmarshall server response : %s\n", string(bytes))
		return
	}

	if !response.Status {
		fmt.Printf("error (response:%v)!\n", response)
	}
}

func CliUnplug(verbs []Verb) {
	serverBaseUrl := getAPIBaseURL(verbs[0])
	method := verbs[0].GetOptionOr("method", "get")
	verbs = verbs[1:]

	path := verbs[0].Name

	reqBody := &UnplugFunctionRequest{
		Method: method,
		Path:   path,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("cannot marshal json (%v)\n", err)
		return
	}

	bodyReader := bytes.NewReader(bodyBytes)

	resp, err := client.Post(serverBaseUrl+"/api/function/unplug", "application/json", bodyReader)
	if err != nil {
		fmt.Printf("error during http request (%v)\n", err)
		return
	}

	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	response := &UnplugResponse{}
	if json.Unmarshal(bytes, response) != nil {
		fmt.Printf("cannot unmarshall server response : %s\n", string(bytes))
		return
	}

	if !response.Status {
		fmt.Printf("error (response:%v)!\n", response)
	}
}

func CliPlugFilter(verbs []Verb) {
	serverBaseUrl := getAPIBaseURL(verbs[0])
	verbs = verbs[1:]

	name := verbs[0].Name
	startFunction := ""
	if len(verbs) > 1 {
		startFunction = verbs[1].Name
	}
	data := ""
	if len(verbs) > 2 {
		data = verbs[2].Name
	}

	reqBody := &PlugFilterRequest{
		Name:          name,
		StartFunction: startFunction,
		Data:          data,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("cannot marshal json (%v)\n", err)
		return
	}

	bodyReader := bytes.NewReader(bodyBytes)

	resp, err := client.Post(serverBaseUrl+"/api/filter/plug", "application/json", bodyReader)
	if err != nil {
		fmt.Printf("error during http request (%v)\n", err)
		return
	}

	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	response := &PlugFilterResponse{}
	if json.Unmarshal(bytes, response) != nil {
		fmt.Printf("cannot unmarshall server response : %s\n", string(bytes))
		return
	}

	if !response.Status {
		fmt.Printf("error (response:%v)!\n", response)
	} else {
		fmt.Printf("%s\n", response.FilterID)
	}
}

func CliUnplugFilter(verbs []Verb) {
	serverBaseUrl := getAPIBaseURL(verbs[0])
	verbs = verbs[1:]

	filterID := verbs[0].Name

	req, err := http.NewRequest("delete", serverBaseUrl+"/api/filter/plug/"+filterID, nil)
	req.Header.Set("content-type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error during http request (%v)\n", err)
		return
	}

	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	response := &PlugResponse{}
	if json.Unmarshal(bytes, response) != nil {
		fmt.Printf("cannot unmarshall server response : %s\n", string(bytes))
		return
	}

	if !response.Status {
		fmt.Printf("error (response:%v)!\n", response)
	} else {
		fmt.Printf("ok, deleted\n")
	}
}

func detectContentTypeFromFileName(name string) string {
	i := strings.LastIndex(name, ".")
	if i < 0 {
		return "application/octet-stream"
	}

	extension := name[i+1:]

	mimeType, ok := tools.MimeTypes[extension]
	if !ok {
		return "application/octet-stream"
	}

	return mimeType
}

func getBaseUrl(verb Verb) string {
	defaultUrl, ok := os.LookupEnv("MYOWNCLUSTER_SERVER_BASE_URL")
	if !ok {
		defaultUrl = "https://localhost:8443"
	}

	return verb.GetOptionOr("baseUrl", defaultUrl)
}

func getAPIBaseURL(verb Verb) string {
	baseURL := getBaseUrl(verb)

	return baseURL + "/my-own-cluster"
}

func uploadFile(serverBaseUrl string, method string, path string, contentType string, fileName string, tags map[string]string) (string, error) {
	techID, err := registerBlob(serverBaseUrl, contentType, fileName)
	if err != nil {
		return "", fmt.Errorf("error while registering blob (%v)", err)
	}

	reqBody := &PlugFileRequest{
		Method: method,
		Path:   path,
		Name:   fmt.Sprintf("techID://%s", techID),
		Tags:   tags,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("cannot marshal json (%v), raw string: %v", err, reqBody)
	}

	bodyReader := bytes.NewReader(bodyBytes)

	resp, err := client.Post(serverBaseUrl+"/api/file/plug", "application/json", bodyReader)
	if err != nil {
		return "", err
	}

	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	response := &PlugResponse{}
	err = json.Unmarshal(bytes, response)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshall server response (%v) : %s", err, string(bytes))
	}

	if !response.Status {
		return "", fmt.Errorf("bad response status")
	}

	return techID, nil
}

type CallFunctionRequest struct {
	Name           string    `json:"name"`
	StartFunction  *string   `json:"start_function,omitempty"`  // defaults to "_start"
	Arguments      *[]int    `json:"arguments,omitempty"`       // defaults to []int{}
	Mode           *string   `json:"mode,omitempty"`            // "direct" or "posix" (only for webassembly code, not applicable for javascript)
	Input          *string   `json:"input,omitempty"`           // defaults to empty buffer
	POSIXFilename  *string   `json:"posix_file_name,omitempty"` // defaults to "a.out"
	POSIXArguments *[]string `json:"posix_arguments,omitempty"` // defaults to []string{}
}

type CallFunctionResponse struct {
	Result int    `json:"result"`
	Output string `json:"output"`
	Error  bool   `json:"error"`
}

func CliCallFunction(verbs []Verb) {
	// skip the verb that triggered us, it is given to us in case it contains options
	baseUrl := getAPIBaseURL(verbs[0])
	verbs = verbs[1:]

	functionName := verbs[0].Name
	mode := strings.ToLower(verbs[0].GetOptionOr("mode", "direct"))
	input := base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString([]byte(verbs[0].GetOptionOr("input", "")))

	bodyReq := &CallFunctionRequest{}
	startFunction := verbs[0].GetOptionOr("start_function", "_start")
	bodyReq.StartFunction = &startFunction

	switch mode {
	case "posix":
		arguments := []string{}
		for a := 1; a < len(verbs); a++ {
			arguments = append(arguments, verbs[a].Name)
		}

		bodyReq.Name = functionName
		bodyReq.Mode = &mode
		bodyReq.Input = &input
		posixFileName := verbs[0].GetOptionOr("wasi_file_name", functionName)
		bodyReq.POSIXFilename = &posixFileName
		bodyReq.POSIXArguments = &arguments
		break

	case "direct":
		arguments := []int{}
		for a := 1; a < len(verbs); a++ {
			an, err := strconv.Atoi(verbs[a].Name)
			if err != nil {
				fmt.Println("Bad argument !", verbs[a].Name)
				return
			}
			arguments = append(arguments, an)
		}

		bodyReq.Name = functionName
		bodyReq.Mode = &mode
		bodyReq.Input = &input
		bodyReq.Arguments = &arguments
		break

	default:
		fmt.Printf("unkwown run mode '%s'\n", mode)
		return
	}

	bodyBytes, err := json.Marshal(bodyReq)
	if err != nil {
		fmt.Printf("cannot marshal json\n")
		return
	}

	bodyReader := bytes.NewReader(bodyBytes)

	resp, err := client.Post(baseUrl+"/api/function/call", "application/json", bodyReader)
	if err != nil {
		panic(err)
	}

	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Print(string(bytes))
}
