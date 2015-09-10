Package proxydial provides a safer equivalent to Go's builtin
[net.Dial](https://godoc.org/net#Dial), that you can use if you're acting as a
proxy (or in any way making http requests to URLs you don't trust).

It prevents connections to internal IP addresses, which can cause [severe vulnerabilities](https://www.gnu.gl/blog/Posts/multiple-vulnerabilities-in-pocket/).

Installation
===========

    go get github.com/superhuman/proxydial

Usage
=====

See [GoDoc](https://godoc.org/github.com/superhuman/proxydial) for usage
documentation.

License
=======

proxydial is made available under the MIT license. See LICENSE.mit for details.

Meta-fu
=======

Bug reports and feature requests are welcome. If you think you've found a security vulnerability,
please [email me](mailto:conrad@superhuman.com) directly.
