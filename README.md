# minissdpd
Client package for communicating with minissdpd over it's Unix socket

# Usage
Examples of package usage can be found in the `minissdpc` client application (https://github.com/forfuncsake/minissdpd/tree/master/cmd/minissdpc).

# Documentation
godoc can be found here: https://godoc.org/github.com/forfuncsake/minissdpd

# Why?
I wanted to run some mock devices (for home automation) on my synology using the Belkin Wemo protocol, but UDP port 1900 was already in use.
Thankfully, Synology implemented the miniupnp/minissdp proxy in DSM and we can use that to advertise additional custom SSDP-based services.

# Contributions
Feedback, Issues and PRs are all welcome!



[![Build Status](https://travis-ci.org/forfuncsake/minissdpd.svg?branch=master)](https://travis-ci.org/forfuncsake/minissdpd)
