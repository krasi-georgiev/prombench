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
	"os"
	"strings"

	"github.com/pkg/errors"
	shell "github.com/progrium/go-shell"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup a Kubernetes cluster on AWS",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSetup(cmd, args); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}

func initKops() {
	bkt, err := terraform("output kops_state_bucket")
	if err != nil {
		panic(err)
	}
	stateBucket = bkt

	kops = shell.Cmd("kops", "--state", bkt).OutputFn()
}

func runSetup(cmd *cobra.Command, args []string) error {
	// if len(args) < 1 {
	// 	return errors.New("scenario path")
	// }
	if err := checkDeps("aws", "terraform", "kops"); err != nil {
		return err
	}

	// path := args[0]

	// spec, err := readSpec(args[0])
	// if err != nil {
	// 	return errors.Wrap(err, "read spec file")
	// }

	tf := shell.Run(
		fmt.Sprintf("AWS_REGION=%q terraform apply", awsRegion),
		fmt.Sprintf("-var \"dns_domain=%s\"", domain),
		fmt.Sprintf("-var \"cluster_name=%s\"", clusterName),
		"./templates",
	)
	if tf.ExitStatus != 0 {
		return errors.Errorf("terraform failed %d", tf.ExitStatus)
	}

	initKops()

	// Target of expanded template files.
	// if err := os.MkdirAll(filepath.Join(path, ".build"), 0755); err != nil {
	// 	return errors.Wrap(err, "create build dir")
	// }

	out, err := kops("get cluster")
	if err != nil {
		return errors.Wrap(err, "kops get cluster")
	}
	if !strings.Contains(out, clusterName+"."+domain) {
		_, err := kops(
			"create cluster",
			"--name", clusterName+"."+domain,
			"--cloud aws ",
			"--zones", awsRegion+"a",
			"--kubernetes-version 1.5.2",
			"--master-size t2.large",
			"--yes",
		)
		if err != nil {
			return err
		}
	}

	if err := shell.Run(
		"EDITOR='./ed.sh manifests/kops/regular-ig.yaml'",
		"kops --state", stateBucket,
		"edit instancegroup nodes",
	).Error(); err != nil {
		return err
	}

	if err := shell.Run(
		"EDITOR='./ed.sh manifests/kops/prometheus-ig.yaml'",
		"kops --state", stateBucket,
		"create instancegroup prometheus",
	).Error(); err != nil {
		return err
	}

	if _, err := kops("update cluster --yes"); err != nil {
		return err
	}

	return nil
}
