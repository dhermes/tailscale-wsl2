# `tailscale-wsl2`

## TL;DR

Running two reverse proxies (one in Windows, one in the WSL2 Linux VM), the
Windows Tailscale daemon can be accessed via WSL2:

```
$ tailscale ip -4
Warning: client version "1.18.1-t0a4334048-gb05dc05d4" != tailscaled server version "1.18.0-t22d969975-g7022e5a4c"
100.101.102.102
$ tailscale ip -4 2> /dev/null
100.101.102.102
$
$ tailscale status 2> /dev/null
100.101.102.102 suspicious-hawking   dhermes@     windows -
100.101.102.103 pedantic-yonath      dhermes@     linux   -
```

## Why?

On standard Linux, `tailscaled` runs via `systemd` and makes deep modifications
to network interfaces and routing to make sure Tailscale packets get handled
by Tailscale. However, on WSL2 there is no `systemd` and the Linux VM doesn't
have "real" control of Windows host networking, so running `tailscale` directly
in WSL2 isn't **really** an option.

However, Tailscale for Windows is a perfectly good choice and all of the
**host** networking changes made by it are still visible to the WSL2 VM.
This means things like Tailscale DNS and Tailscale routing "just work"

```
$ nslookup pedantic-yonath
Server:         100.100.100.100
Address:        100.100.100.100#53

Name:   pedantic-yonath.dhermes.github.beta.tailscale.net
Address: 100.101.102.103

$
$ ssh dhermes@pedantic-yonath
dhermes@pedantic-yonath:~$
dhermes@pedantic-yonath:~$ tailscale ip -4
100.101.102.103
```

However, from the Linux side of WSL2, there is still no **direct** way to
get information about the current node and other parts of the Tailnet:

```
$ tailscale ip -4
Failed to connect to local Tailscale daemon for /localapi/v0/status; not running?
Error: dial unix /var/run/tailscale/tailscaled.sock: connect: no such file or directory
```

This repository provides two reverse proxy binaries (one for Windows, one for
Linux) that make it possible to directly reference the Windows-managed
Tailscale daemon from inside the WSL2 VM.

## How?

On Linux, the `tailscale` binary expects to communicate with `tailscaled` via
a Unix Domain Socket (UDS) at `/var/run/tailscale/tailscaled.sock`. However,
Windows (i.e. not Unix-y) doesn't have a UDS equivalent, so there isn't even
a host socket that the WSL2 VM **could** try to interact with. Due to this
Windows feature difference, the Tailscale daemon runs as a cleartext HTTP
service on `localhost:41112` (as of Tailscale 1.18.1).

Unfortunately the local network interface on Windows is (intentionally) not
exposed to the WSL2 VM. However, there **is** a virtual network interface
that can be used to bridge the Windows host to the WSL2 VM:

```
PS C:\Users\dhermes> ipconfig

Windows IP Configuration

...
Ethernet adapter vEthernet (WSL):

   Connection-specific DNS Suffix  . :
   Link-local IPv6 Address . . . . . : fe80::ed6d:71c5:2607:3ce6%22
   IPv4 Address. . . . . . . . . . . : 172.27.64.1
   Subnet Mask . . . . . . . . . . . : 255.255.240.0
   Default Gateway . . . . . . . . . :
...
```

To expose the Tailscale daemon to the WSL2 VM, we bind a server to the
virtual network IP (e.g. here `172.27.64.1`) and run a reverse proxy
**in Windows** that will be available to the WSL2 VM:

```
PS C:\Users\dhermes\tailscale-wsl2> go install .\cmd\tailscale-wsl2-windows\
PS C:\Users\dhermes\tailscale-wsl2> tailscale-wsl2-windows.exe --vethernet-wsl-ip 172.27.64.1
```

Running a TCP reverse proxy is super easy with the super awesome
`inet.af/tcpproxy` [package][2]. This package was of course created by the
lovely folks at Tailscale. They **also** made the `inet.af/wf` [package][4] for
Windows Firewall operations, which turns out to be necessary to bind to
the virtual network IP. See [Windows Firewall][3] for more information about
which firewall rules are necessary.

But, this only solves half of the problem. The `tailscale` CLI still assumes
it will have a UDS to talk to. To solve this problem, we run a **second**
reverse proxy, but on the Linux side of the house:

```
$ go install ./cmd/tailscale-wsl2-linux/
$ sudo tailscale-wsl2-linux --host-ip 172.27.64.1 --tailscale-socket /var/run/tailscale/tailscaled.sock
```

The **default** value of `--tailscale-socket` is actually
`/var/run/tailscale/wsl2-tailscaled.sock` to avoid colliding with the socket
that `tailscaled` expects to be the owner of. If that default is used instead
then it needs to be explicitly provided to `tailscale`

```
$ tailscale --socket /var/run/tailscale/wsl2-tailscaled.sock ip -4
Warning: client version "1.18.1-t0a4334048-gb05dc05d4" != tailscaled server version "1.18.0-t22d969975-g7022e5a4c"
100.101.102.102
$ tailscale --socket /var/run/tailscale/wsl2-tailscaled.sock ip -4 2> /dev/null
100.101.102.102
```

Having the client and server version mismatch is not great and actually
causes the server to **reject** some requests:

```
$ tailscale ping pedantic-yonath
Warning: client version "1.18.1-t0a4334048-gb05dc05d4" != tailscaled server version "1.18.0-t22d969975-g7022e5a4c"
2021/12/05 19:47:58 GotNotify: Version mismatch! frontend="1.18.1-t0a4334048-gb05dc05d4" backend="1.18.0-t22d969975-g7022e5a4c"
Notify.ErrMessage: GotNotify: Version mismatch! frontend="1.18.1-t0a4334048-gb05dc05d4" backend="1.18.0-t22d969975-g7022e5a4c"
```

Installing `tailscale` from source is probably the best plan to ensure this
version mismatch doesn't occur, but at least for now the custom APT
[package][1] matches the Windows version:

```
$ [sudo] apt-get install tailscale=1.18.0
$ tailscale version
1.18.0
  tailscale commit: 22d9699759fa34247153a542e9c4af5696c01fdf
  other commit: 7022e5a4ccce1d12fbe4f679d641d816d81491a1
  go version: go1.17.2-ts7037d3ea51
$
$ tailscale ip -4
100.101.102.102
$ tailscale ping pedantic-yonath
pong from pedantic-yonath (100.101.102.103) via 192.168.7.131:41641 in 5ms
```

[1]: https://tailscale.com/kb/1039/install-ubuntu-2004/
[2]: https://pkg.go.dev/inet.af/tcpproxy
[3]: WINDOWS_FIREWALL.md
[4]: https://pkg.go.dev/inet.af/wf
