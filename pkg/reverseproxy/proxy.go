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

package reverseproxy

import (
	"net"

	"inet.af/tcpproxy"
)

// Forward runs a proxy for an already bound net listener.
//
// This enables the listener to be both a TCP socket or a Unix domain
// socket (UDS).
func Forward(l net.Listener, addr string) error {
	dp := tcpproxy.To(addr)
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go dp.HandleConn(conn)
	}
}
