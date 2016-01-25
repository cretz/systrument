package context

import (
	"github.com/cretz/systrument/data"
	"io/ioutil"
	"fmt"
	"github.com/cretz/systrument/util"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"github.com/cretz/systrument/resource"
)

type Context struct {
	util.Logger
	resource.Resources
	Data *data.Data
	IsRemote bool
}

var unmarshalStripped = func(byts []byte, v interface{}) error {
	// Strip comments (this happens after template application)
	properByts, err := ioutil.ReadAll(util.NewCommentStrippedJSONReader(bytes.NewBuffer(byts)))
	if err != nil {
		return fmt.Errorf("Unable to strip JSON comments: %v", err)
	}
	return json.Unmarshal(properByts, v)
}

func FromConfigFiles(files []string, verbose bool) (*Context, error) {
	ctx := &Context{
		Logger: util.GoLoggerWrapper(log.New(os.Stdout, "", log.LstdFlags), verbose),
		Resources: resource.LocalResources(),
		Data: data.NewData(),
	}
	// Load each file, strip JSON comments, load into data
	for _, file := range files {
		byts, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("Unable to read file %v: %v", file, err)
		}
		if err = ctx.Data.ApplyTemplateAndMerge(byts, unmarshalStripped); err != nil {
			return nil, fmt.Errorf("Error handling config file %v: %v", file, err)
		}
	}
	return ctx, nil
}

func FromRemoteStdPipe() (*Context, error) {
	panic("TODO")
}