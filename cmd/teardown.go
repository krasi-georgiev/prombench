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

	"github.com/pkg/errors"
	shell "github.com/progrium/go-shell"
	"github.com/spf13/cobra"
)

// teardownCmd represents the teardown command
var teardownCmd = &cobra.Command{
	Use:   "teardown",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runTeardown(cmd, args); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(teardownCmd)
}

func runTeardown(cmd *cobra.Command, args []string) error {
	if err := checkDeps("aws", "terraform", "kops"); err != nil {
		return err
	}

	if _, err := kops(
		"delete cluster --yes",
		"--name", clusterName+"."+domain,
	); err != nil {
		return errors.Wrap(err, "kops delete cluster")
	}

	tf := shell.Run(
		fmt.Sprintf("AWS_REGION=%q terraform destroy -force", awsRegion),
		fmt.Sprintf("-var \"dns_domain=%s\"", domain),
		fmt.Sprintf("-var \"cluster_name=%s\"", clusterName),
		"./templates",
	)
	if tf.ExitStatus != 0 {
		return errors.Errorf("terraform failed %d", 1)
	}

	shell.Run("rm -f terraform.tfstate*")

	return nil
}
