// Copyright 2021 Danny Hermes
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net"
	"os"

	"github.com/dhermes/tailscale-wsl2/pkg/reverseproxy"
	"github.com/spf13/cobra"
)

func serve(hostIP string, hostPort int, sockAddr string) error {
	err := os.RemoveAll(sockAddr)
	if err != nil {
		return err
	}

	l, err := net.Listen("unix", sockAddr)
	if err != nil {
		return err
	}
	err = os.Chmod(sockAddr, 0666)
	if err != nil {
		return err
	}

	hostAddr := fmt.Sprintf("%s:%d", hostIP, hostPort)
	return reverseproxy.Forward(l, hostAddr)
}

func run() error {
	hostIP := ""
	hostPort := 41113
	sockAddr := "/var/run/tailscale/wsl2-tailscaled.sock"
	short := "Run a TCP to UDS reverse proxy for the Tailscale daemon"

	cmd := &cobra.Command{
		Use:           "tailscale-wsl2-linux",
		Short:         short,
		Long:          short + "\n\nThis converts a Host TCP port (running Tailscale on Windows) into a UDS\nin the WSL2 VM that can be used directly by the Linux 'tailscale' binary",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return serve(hostIP, hostPort, sockAddr)
		},
	}

	cmd.PersistentFlags().StringVar(
		&hostIP,
		"host-ip",
		hostIP,
		"The IP on the Windows Host where the 'tailscaled' proxy is running\nThis is likely the IPv4 address of the 'vEthernet (WSL)' Windows network adapter",
	)
	cmd.PersistentFlags().IntVar(
		&hostPort,
		"host-port",
		hostPort,
		"The port on the Windows Host where the 'tailscaled' proxy is running",
	)
	cmd.PersistentFlags().StringVar(
		&sockAddr,
		"tailscale-socket",
		sockAddr,
		"The path to use for the Tailscale Unix Domain Socket (UDS)",
	)

	err := cobra.MarkFlagRequired(cmd.PersistentFlags(), "host-ip")
	if err != nil {
		return err
	}

	return cmd.Execute()
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
