package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
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
	XMLName      xml.Name `xml:"default-entrypoints"`
	PutItem      string   `xml:"put-item"`
	PutItemNamed string   `xml:"put-item-named"`
	GetItem      string   `xml:"get-item"`
	GetItemNamed string   `xml:"get-item-named"`
	PostStatus   string   `xml:"post-status"`
}

// Mapping API object
type CustomEndpoint struct {
	XMLName        xml.Name         `xml:"endpoint"`
	Uri            CustomUri        `xml:"uri"`
	SimpleFunction []SimpleFunction `xml:"simple-function"`
	DefinedRouting *DefinedRouting  `xml:"pre-defined-routing,omitempty"`
	CustomRouting  *CustomRouting   `xml:"user-defined-routing,omitempty"`
}

// Mapping URI object
type CustomUri struct {
	XMLName   xml.Name `xml:"uri"`
	Value     string   `xml:",innerxml"`
	UseParams bool     `xml:"use_params,attr"`
	Params    string   `xml:"path_params,attr"`
}

type SimpleFunction struct {
	XMLName      xml.Name `xml:"simple-function"`
	Method       string   `xml:"method,attr"`
	UseBody      bool     `xml:"use_body,attr"`
	BodyItems    string   `xml:"body_items,attr"`
	FunctionName string   `xml:",innerxml"`
}

type DefinedRouting struct {
	XMLName     xml.Name `xml:"pre-defined-routing"`
	ObjectIdent string   `xml:"object_identifier,attr"`
	Ques        []Queue  `xml:"queue"`
}

type CustomRouting struct {
	XMLName      xml.Name `xml:"user-defined-routing"`
	ObjectIdent  string   `xml:"object_identifier,attr"`
	RoutingIdent string   `xml:"routing_identifier,attr"`
}

type Queue struct {
	// XMLName xml.Name
	Que    string           `xml:",chardata"`
	Name   *string          `xml:"name,attr"`
	Input  *string          `xml:"input,attr"`
	Output *string          `xml:"output,attr"`
	Params *json.RawMessage `xml:"params"`
}

type MethodOptions struct {
	//	List of params in specified in uri Path
	PathParams  *[]string
	Files       *[]multipart.File
	FileHeaders *[]multipart.FileHeader
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

// func (api CustomEndpoint) parseMethodOptions(c *gin.Context) (MethodOptions, error) {
// 	var options = MethodOptions{}
// 	//	Collect path parameters present
// 	if api.Uri.UseParams {
// 		parameters := strings.Split(api.Uri.Params, " ")
// 		for _, param := range parameters {
// 			options.PathParams = append(options.PathParams, c.Param(param))
// 		}
// 	}
// 	if api.SimpleFunction[0].UseBody {
// 		identifiers := strings.Split(api.SimpleFunction[0].BodyItems, " ")
// 		for _, identifier := range identifiers {
// 			file, fileHeaders := getFormFile(c, identifier)
// 			//	Break in case unable to deal with file (caused by end user)
// 			if file == nil {
// 				return MethodOptions{}, errors.New("Failed to retrieve file.")
// 			}
// 			options.Files = append(options.Files, file)
// 			options.FileHeaders = append(options.FileHeaders, fileHeaders)
// 			options.FileIdents = append(options.FileIdents, identifier)
// 		}
// 	}
// 	return options, nil
// }

func getSFuncOptions(c *gin.Context, uri CustomUri, sf SimpleFunction) MethodOptions {
	options := MethodOptions{}
	if uri.UseParams {
		idents := strings.Split(uri.Params, " ")
		params := []string{}
		for _, ident := range idents {
			params = append(params, c.Param(ident))
		}
		options.PathParams = &params
	}
	if sf.UseBody {
		idents := strings.Split(sf.BodyItems, " ")
		files := []multipart.File{}
		file_headers := []multipart.FileHeader{}
		for _, ident := range idents {
			f, fh := getFormFile(c, ident)
			files = append(files, f)
			file_headers = append(file_headers, fh)
		}
		options.Files = &files
		options.FileHeaders = &file_headers
	}
	return options
}

// For mapping functions
type stubMapping map[string]interface{}

var StubStorage = stubMapping{
	"postItem":      simpPostItem,
	"putItem":       simpPutItem,
	"putItemNamed":  simpPutItemNamed,
	"listBuckets":   simpListBuckets,
	"bucketDetails": simpListBucketDetais,
	"getItem":       simpGetItem,
	"getItemNamed":  simpGetItemNamed,
	"deleteBucket":  simpDeleteBucket,
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
	if res != nil && len(res) > 0 {
		result = res[0].Interface()
	} else {
		result = nil
	}
	return
}

// This loop is roundabout for finding right api config (it would be best replaced if possible)
func getApiConfig(c *gin.Context) *CustomEndpoint {
	endpoint_path := c.FullPath()
	endpoint_method := c.Request.Method
	for _, api := range config.EndPoints {
		if endpoint_path == api.Uri.Value {
			for _, action := range api.SimpleFunction {
				if endpoint_method == action.Method {
					//	Reduce action list to only 1
					api.SimpleFunction = []SimpleFunction{action}
					return &api
				}
			}
		}
	}
	return nil
}
