package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

// Root element
type ApiConfig struct {
	XMLName   xml.Name         `xml:"apilist"`
	Def_apis  DefaultEndpoint  `xml:"default-entrypoints"`
	EndPoints []CustomEndpoint `xml:"endpoint"`
}

// Map for default endpoints (may be useless)
type DefaultEndpoint struct {
	XMLName xml.Name `xml:"default-entrypoints"`
	GET     string   `xml:"default-get"`
	POST    string   `xml:"default-post"`
	PUT     string   `xml:"default-put"`
	DELETE  string   `sml:"default-delete"`
}

// Mapping API object
type CustomEndpoint struct {
	XMLName xml.Name  `xml:"endpoint"`
	Uri     CustomUri `xml:"uri"`
	Action  []Action  `xml:"action"`
	// Process Process   `xml:"function-extra"`
	// Pre     PredefinedProcess `xml:"predefined-process"`
	// Usr     UserPRocess       `xml:"user-process"`
}

// Mapping URI object
type CustomUri struct {
	XMLName   xml.Name `xml:"uri"`
	Value     string   `xml:",innerxml"`
	UseParams bool     `xml:"use_params,attr"`
	Params    string   `xml:"path_params,attr"`
}

type Action struct {
	XMLName   xml.Name `xml:"action"`
	Method    string   `xml:"method,attr"`
	UseBody   bool     `xml:"use_body,attr"`
	BodyItems string   `xml:"body_items,attr"`
	Func      []string `xml:"function-byname"`
}

type MethodOptions struct {
	//	List of params in specified in uri Path
	PathParams []string
	//	List of files and headers (may be obselete, will investigate during dev)
	Files       []multipart.File
	FileHeaders []multipart.FileHeader
	//	Identifiers to getFormFile
	FileIdents []string
}

func parseConfig() ApiConfig {
	path := "/usr/src/app/config/config.xml"
	//	 Fallback to default config
	if _, err := os.Stat(path); err != nil {
		path = "/usr/src/app/config/default-config.xml"
	}
	//	LOCAL DEV TEST
	if operating := runtime.GOOS; operating == "windows" {
		path = "C:/Users/bean/project-y4/msgb/go-rest/config/default-config.xml"
	}
	xmlFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)
	var apis ApiConfig

	xml.Unmarshal(byteValue, &apis)

	return apis
}

func (api CustomEndpoint) parseMethodOptions(c *gin.Context) (MethodOptions, error) {
	var options = MethodOptions{}
	//	Collect path parameters present
	if api.Uri.UseParams {
		parameters := strings.Split(api.Uri.Params, " ")
		for _, param := range parameters {
			options.PathParams = append(options.PathParams, c.Param(param))
		}
	}
	if api.Action[0].UseBody {
		identifiers := strings.Split(api.Action[0].BodyItems, " ")
		for _, identifier := range identifiers {
			file, fileHeaders := getFormFile(c, identifier)
			//	Break in case unable to deal with file (caused by end user)
			if file == nil {
				return MethodOptions{}, errors.New("Failed to retrieve file.")
			}
			options.Files = append(options.Files, file)
			options.FileHeaders = append(options.FileHeaders, fileHeaders)
			options.FileIdents = append(options.FileIdents, identifier)
		}
	}
	return options, nil
}

// used by main.go
func MainAction(c *gin.Context) {
	api := getApiConfig(c)
	//	No api found
	if api == nil {
		return
	}
	options, err := api.parseMethodOptions(c)
	//	Throw error if an error occured when parsing
	if err != nil {
		return
	}
	for _, act := range api.Action {
		//	Actual execution of string based method
		err1, err2 := Call(act.Func[0], c, options)
		if err2 != nil {
			fmt.Println("Error calling the method:" + act.Func[0])
			fmt.Println(err.Error())
			c.String(http.StatusInternalServerError, "Error calling the method.")
		}
		if err1 != nil {
			fmt.Println("Error within the method:" + act.Func[0])
			fmt.Println(err.Error())
			c.String(http.StatusInternalServerError, "Error within the method.")
		}
	}
}

// For mapping functions
type stubMapping map[string]interface{}

var StubStorage = stubMapping{
	"listItems":            listItems,
	"getItem":              getItem,
	"putItem":              putItem,
	"pushItem":             pushItem,
	"predefinedBackground": processPredefinedBackground,
	"predefinedBGray":      processPredefinedGray,
	"putItemNamed":         putItemNamed,
}

// Dont ask (https://medium.com/@vicky.kurniawan/go-call-a-function-from-string-name-30b41dcb9e12)
func Call(funcName string, params ...interface{}) (result interface{}, err error) {
	f := reflect.ValueOf(StubStorage[funcName])
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is out of index.")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	var res []reflect.Value
	res = f.Call(in)
	result = res[0].Interface()
	return
}

// This loop is roundabout for finding right api config (it would be best replaced if possible)
func getApiConfig(c *gin.Context) *CustomEndpoint {
	endpoint_path := c.FullPath()
	endpoint_method := c.Request.Method
	for _, api := range config.EndPoints {
		if endpoint_path == api.Uri.Value {
			for _, action := range api.Action {
				if endpoint_method == action.Method {
					//	Reduce action list to only 1
					api.Action = []Action{action}
					return &api
				}
			}
		}
	}
	return nil
}

func placeholderFunction() {
	fmt.Println("placeholderFunctionCall")
}
