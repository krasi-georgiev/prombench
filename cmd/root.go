// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	yaml "gopkg.in/yaml.v1"

	"github.com/pkg/errors"
	shell "github.com/progrium/go-shell"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "prombench",
	Short: "A tool to run reproducible Prometheus benchmarks",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

var (
	domain          = "dev.coreos.systems"
	awsRegion       = "eu-west-1"
	clusterName     = ""
	stateBucket     = ""
	fullClusterName = ""

	whoami    = shell.Cmd("whoami").OutputFn()
	kops      func(...interface{}) (string, error)
	terraform = shell.Cmd("terraform").OutputFn()
)

const (
	clusterNamePrefix = "prombench-"
)

func init() {
	shell.Trace = true
	shell.Panic = false
	shell.Tee = os.Stdout

	if d := os.Getenv("DOMAIN"); d != "" {
		domain = d
	}
	if r := os.Getenv("AWS_REGION"); r != "" {
		awsRegion = r
	}
	if c := os.Getenv("CLUSTER_NAME"); c != "" {
		clusterName = c
	} else {
		name, err := whoami()
		if err != nil {
			panic(err)
		}
		clusterName = clusterNamePrefix + name
	}

	fullClusterName = clusterName + "." + domain

	initKops()
}

const specFile = "spec.yaml"

type Spec struct {
}

func readSpec(dir string) (*Spec, error) {
	b, err := ioutil.ReadFile(filepath.Join(dir, specFile))
	if err != nil {
		return nil, err
	}
	var spec Spec
	if err := yaml.Unmarshal(b, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

func checkDeps(files ...string) error {
	for _, f := range files {
		if _, err := exec.LookPath(f); err != nil {
			if e, ok := err.(*exec.Error); ok && e.Err == exec.ErrNotFound {
				return errors.Errorf("%s executable missing", f)
			}
			return err
		}
	}
	return nil
}
