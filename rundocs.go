// Copyright 2013 bee authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.
package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var cmdRundocs = &Command{
	UsageLine: "rundocs [-isDownload=true] [-docport=8888]",
	Short:     "rundocs will run the docs server,default is 8089",
	Long: `
-d meaning will download the docs file from github
-p meaning server the Server on which port, default is 8089

`,
}

var (
	swaggerVersion = "2"
	swaggerlink    = "https://github.com/beego/swagger/archive/v" + swaggerVersion + ".zip"
)

type docValue string

func (d *docValue) String() string {
	return fmt.Sprint(*d)
}

func (d *docValue) Set(value string) error {
	*d = docValue(value)
	return nil
}

var isDownload docValue
var docport docValue

func init() {
	cmdRundocs.Run = runDocs
	cmdRundocs.Flag.Var(&isDownload, "isDownload", "weather download the Swagger Docs")
	cmdRundocs.Flag.Var(&docport, "docport", "doc server port")
}

func runDocs(cmd *Command, args []string) int {
	if isDownload == "true" {
		downloadFromURL(swaggerlink, "swagger.zip")
		err := unzipAndDelete("swagger.zip")
		if err != nil {
			fmt.Println("has err exet unzipAndDelete", err)
		}
	}
	if docport == "" {
		docport = "8089"
	}
	if _, err := os.Stat("swagger"); err != nil && os.IsNotExist(err) {
		fmt.Println("there's no swagger, please use bee rundocs -isDownload=true downlaod first")
		os.Exit(2)
	}
	fmt.Println("start the docs server on: http://127.0.0.1:" + docport)
	log.Fatal(http.ListenAndServe(":"+string(docport), http.FileServer(http.Dir("swagger"))))
	return 0
}

func downloadFromURL(url, fileName string) {
	var down bool
	if fd, err := os.Stat(fileName); err != nil && os.IsNotExist(err) {
		down = true
	} else if fd.Size() == int64(0) {
		down = true
	} else {
		ColorLog("[%s] Filename %s already exist\n", INFO, fileName)
		return
	}
	if down {
		ColorLog("[%s]Downloading %s to %s\n", SUCC, url, fileName)
		output, err := os.Create(fileName)
		if err != nil {
			ColorLog("[%s]Error while creating %s: %s\n", ERRO, fileName, err)
			return
		}
		defer output.Close()

		response, err := http.Get(url)
		if err != nil {
			ColorLog("[%s]Error while downloading %s:%s\n", ERRO, url, err)
			return
		}
		defer response.Body.Close()

		n, err := io.Copy(output, response.Body)
		if err != nil {
			ColorLog("[%s]Error while downloading %s:%s\n", ERRO, url, err)
			return
		}
		ColorLog("[%s] %d bytes downloaded.\n", SUCC, n)
	}
}

func unzipAndDelete(src string) error {
	ColorLog("[%s]start to unzip file from %s\n", INFO, src)
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		if f.FileInfo().IsDir() {
			os.MkdirAll(f.Name, f.Mode())
		} else {
			f, err := os.OpenFile(
				f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	os.RemoveAll("swagger")
	err = os.Rename("swagger-"+swaggerVersion, "swagger")
	if err != nil {
		ColorLog("[%s]Rename swagger-%s to swagger:%s\n", ERRO, swaggerVersion, err)
	}
	ColorLog("[%s]Start delete src file %s\n", INFO, src)
	return os.RemoveAll(src)
}
