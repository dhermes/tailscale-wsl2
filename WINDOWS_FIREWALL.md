# Windows Firewall

> **NOTE**: These instructions assume all actions are taken by clicking
> through the Windows Defender UI. A "production-grade" application would not
> put end users through such an approach, but would instead create the
> necessary firewall rules at install time. The Tailscale folks have already
> made this (i.e. creating firewall rules programmatically) an easy task with
> the `inet.af/wf` [package][8] for Windows Firewall operations.

When a server binds to an IP, it can limit to `localhost` only or can choose
other IPs associated with the machine. For example on a Windows machine on WiFi
running a WSL2 VM, the server could bind to the machine's IP on the WiFi
network (e.g. `192.168.7.107`) or to the IP for the virtual network interface
exposed to WSL2 (e.g. `172.27.64.1`).

For a Windows program that binds to an IP other than `localhost`, all
TCP or UDP traffic will be blocked by Windows Defender Firewall unless a
firewall rule is added.

## New Application

When a new application (i.e. an application that has not been seen by
Windows Defender Firewall) binds to a non-`localhost` IP, a prompt shows up
asking if the user would like to allow network traffic into the application
via a new firewall:

![Windows Defender sees a new binary][1]

Checking the "Public networks" checkbox and clicking "Allow Access" will
allow computers on any network to connect. This includes the virtual network
interface exposed to WSL2 but also includes network interfaces we don't need
(or want exposed). So we'll need to modify the firewall so that it only allows
traffic in from the WSL2 VM.

## Modify Firewall Rule

Opening the Windows Defender Firewall application only gives the basic
options:

![Windows Defender basic options][2]

By clicking "Advanced settings" you can actually dive into the firewall
rules and find the "Inbound" rule

![Windows Defender inbound rules][3]

## Desired Rule

By default, the "Windows Security Alert" prompt will create **two** firewall
rules: one for TCP and one for UDP. However, in this case we only need TCP
so having the second one is more permissive than we have need for, it can
be deleted:

![Windows Defender UDP rule][4]

Select the remaining TCP rule, click the "Properties" option and then open
the "Scope" properties section. Here, in addition to the local IPs, we can
restrict "remote" traffic to the `/24` subnet for the virtual WSL2
network interface (e.g. `172.27.64.0/24`):

![Windows Defender remote IPs][5]

Trying to save this as-is will **fail** due to the edge traversal setting:

![Windows Defender invalid][6]

To actually save with the new IP range, go to the "Advanced" properties
section and change the "Edge traversal" setting from the default value of
"Defer to user" to "Block edge traversal":

![Edge Traversal][7]

We don't have a need for NAT edge traversal because this traffic is all
expected to be local (even if it **appears** remote over the virtual WSL2
interface).

[1]: _images/01-defender-new.png
[2]: _images/02-defender-basic.png
[3]: _images/03-defender-inbound.png
[4]: _images/04-defender-udp.png
[5]: _images/05-defender-source-ips.png
[6]: _images/06-defender-invalid.png
[7]: _images/07-defender-edge-traversal.png
[8]: https://pkg.go.dev/inet.af/wf
