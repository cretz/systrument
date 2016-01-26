package context

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cretz/systrument/data"
	"github.com/cretz/systrument/resource"
	"github.com/cretz/systrument/util"
	"io/ioutil"
	"log"
	"os"
)

type Context struct {
	util.Logger
	resource.Resources
	Data         *data.Data
	IsRemote     bool
	BaseLocalDir string
	TempDir      string
	RemotePipe   *LocalToRemotePipe
}

var unmarshalStripped = func(byts []byte, v interface{}) error {
	// Strip comments (this happens after template application)
	properByts, err := ioutil.ReadAll(util.NewCommentStrippedJSONReader(bytes.NewBuffer(byts)))
	if err != nil {
		return fmt.Errorf("Unable to strip JSON comments: %v", err)
	}
	return json.Unmarshal(properByts, v)
}

func FromConfigFiles(files []string, verbose bool, overrideLocalDir string) (*Context, error) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "syst-temp")
	if err != nil {
		return nil, fmt.Errorf("Unable to create temporary dir: %v", err)
	}
	ctx := &Context{
		Logger:       util.GoLoggerWrapper(log.New(os.Stdout, "", log.LstdFlags), verbose),
		Resources:    resource.LocalResources(),
		Data:         data.NewData(),
		TempDir:      tempDir,
		BaseLocalDir: overrideLocalDir,
	}
	if overrideLocalDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("Unable to obtain working dir: %v", err)
		}
		ctx.BaseLocalDir = wd
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

func FromRemoteStdPipe(verbose bool, overrideLocalDir string) (*Context, error) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "syst-temp")
	if err != nil {
		return nil, fmt.Errorf("Unable to create temporary dir: %v", err)
	}
	// Since this is remote, we need it usable by everyone since SFTP from the other side writes here...
	if err := os.Chmod(tempDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("Unable to change temp dir privs: %v", err)
	}
	pipe := NewLocalToRemotePipe(os.Stdin, os.Stdout)
	// We need to grab the data from the remote
	// TODO: I would like to support << and >> for remote template parsing, but it would require
	// 	us sending back the entire set of data each time because we need "prev" data
	conf, err := pipe.Request("get-context-data")
	if err != nil {
		return nil, fmt.Errorf("Unable to get context data: %v", err)
	}
	ctx := &Context{}
	ctx.Logger = util.GoLoggerWrapper(log.New(os.Stdout, "", log.LstdFlags), verbose)
	ctx.Resources = newRemoteResources(ctx)
	ctx.Data = data.NewData()
	if err = json.Unmarshal([]byte(conf), &ctx.Data.Values); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal context data from JSON: %v", err)
	}
	ctx.IsRemote = true
	ctx.BaseLocalDir = overrideLocalDir
	ctx.TempDir = tempDir
	ctx.RemotePipe = pipe
	return ctx, nil
}
