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

	"github.com/spf13/cobra"
	"tailscale.com/safesocket"

	"github.com/dhermes/tailscale-wsl2/pkg/reverseproxy"
)

func serve(vethernetWSLIP string, port int, fromAddr string) error {
	addr := fmt.Sprintf("%s:%d", vethernetWSLIP, port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return reverseproxy.Forward(l, fromAddr)
}

func run() error {
	vethernetWSLIP := ""
	port := 41113
	fromAddr := fmt.Sprintf("localhost:%d", safesocket.WindowsLocalPort)
	short := "Run a Host-to-WSL2 TCP reverse proxy for the Tailscale daemon"

	cmd := &cobra.Command{
		Use:           "tailscale-wsl2-windows",
		Short:         short,
		Long:          short + "\n\nSince the Tailscale daemon binds only to localhost, it is not available from WSL2.\nThis runs a proxy that binds to the 'vEthernet (WSL)' network interface.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return serve(vethernetWSLIP, port, fromAddr)
		},
	}

	cmd.PersistentFlags().StringVar(
		&vethernetWSLIP,
		"vethernet-wsl-ip",
		vethernetWSLIP,
		"The IPv4 address of the 'vEthernet (WSL)' Windows network adapter\nThis should be from the 20-bit block ('172.16.0.0/12') used by Docker, e.g. '172.27.64.1'",
	)
	cmd.PersistentFlags().IntVar(
		&port,
		"port",
		port,
		"The port to use for the TCP reverse proxy",
	)
	cmd.PersistentFlags().StringVar(
		&fromAddr,
		"from-addr",
		fromAddr,
		"The local-only address where the Windows Tailscale daemon is running",
	)

	err := cobra.MarkFlagRequired(cmd.PersistentFlags(), "vethernet-wsl-ip")
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
