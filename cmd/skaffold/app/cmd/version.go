/*
Copyright 2018 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/GoogleContainerTools/skaffold/cmd/skaffold/app/flags"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/version"
	"github.com/GoogleContainerTools/skaffold/pkg/webhook/constants"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tcnksm/go-latest"
)

var versionFlag = flags.NewTemplateFlag("{{.Version}}\n", version.Info{})

func NewCmdVersion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunVersion(out, cmd)
		},
	}

	cmd.Flags().VarP(versionFlag, "output", "o", versionFlag.Usage())
	return cmd
}

func RunVersion(out io.Writer, cmd *cobra.Command) error {
	if err := versionFlag.Template().Execute(out, version.Get()); err != nil {
		return errors.Wrap(err, "executing template")
	}

	// Get the latest release
	verCheckCh := make(chan *latest.CheckResponse)
	go func() {
		githubTag := &latest.GithubTag{
			Owner:      constants.GithubOwner,
			Repository: constants.GithubRepo,
		}

		res, err := latest.Check(githubTag, version.Get().Version)

		if err != nil {
			return
		}

		verCheckCh <- res
	}()

	select {
	case <-time.After(2 * time.Second):
	case res := <-verCheckCh:
		if res.Outdated {
			fmt.Fprintf(out, "The latest version is v%s. ", res.Current)
			switch runtime.GOOS {
			case "darwin":
				fmt.Fprintf(out, "Please run the following command to update.\n\n")
				fmt.Fprintf(out, "curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/%s/skaffold-linux-amd64 && chmod +x skaffold && sudo mv skaffold /usr/local/bin\n", res.Current)
			case "windows":
				fmt.Fprintf(out, "Please download the latest binary from the following url and palce it in $PATH directory.\n\n")
				fmt.Fprintf(out, "https://storage.googleapis.com/skaffold/releases/%s/skaffold-windows-amd64.exe\n", res.Current)
			default:
				fmt.Fprintf(out, "Please run the following command to update.\n\n")
				fmt.Fprintf(out, "curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/%s/skaffold-linux-amd64 && chmod +x skaffold && sudo mv skaffold /usr/local/bin\n", res.Current)
			}
		}
	}

	return nil
}
