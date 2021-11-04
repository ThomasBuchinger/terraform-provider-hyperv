package api

import (
	"encoding/json"
	"text/template"
)

type file struct {
	Path string
	// FileSize                uint64
	Size uint64
	// MinimumSize             uint64
	// Attached                bool
}

type createOrUpdateFileArgs struct {
	Source  string
	VhdJson string
}

var createOrUpdateFileTemplate = template.Must(template.New("CreateOrUpdateFile").Parse(`
$ErrorActionPreference = 'Stop'

`))

func (c *HypervClient) CreateOrUpdateFile(path string, source string) (err error) {
	vhdJson, err := json.Marshal(file{
		Path: path,
	})

	if err != nil {
		return err
	}

	err = c.runFireAndForgetScript(createOrUpdateFileTemplate, createOrUpdateFileArgs{
		Source:  source,
		VhdJson: string(vhdJson),
	})

	return err
}

type getFileArgs struct {
	Path string
}

var getFileTemplate = template.Must(template.New("GetFile").Parse(`
$ErrorActionPreference = 'Stop'
$path='{{.Path}}'
`))

func (c *HypervClient) GetFile(path string) (result file, err error) {
	err = c.runScriptWithResult(getFileTemplate, getFileArgs{
		Path: path,
	}, &result)

	return result, err
}

type deleteFileArgs struct {
	Path string
}

var deleteFileTemplate = template.Must(template.New("DeleteFile").Parse(`
$ErrorActionPreference = 'Stop'

$targetDirectory = (split-path '{{.Path}}' -Parent)
$targetName = (split-path '{{.Path}}' -Leaf)
$targetName = $targetName.Substring(0,$targetName.LastIndexOf('.')).split('\')[-1]

Get-ChildItem -Path $targetDirectory |?{$_.BaseName.StartsWith($targetName)} | %{
	Remove-Item $_.FullName -Force
}
`))

func (c *HypervClient) DeleteFile(path string) (err error) {
	err = c.runFireAndForgetScript(deleteFileTemplate, deleteFileArgs{
		Path: path,
	})

	return err
}
