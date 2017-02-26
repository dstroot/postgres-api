// The MIT License (MIT)
//
// Copyright (c) 2017 Daniel J. Stroot
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

/*
Package main initializes the application and runs it using go1.8
features to stop it gracefully and allow connections to drain.

When it starts it logs it's hostname and what port it is listening on
as well as the go version it was built with and the git commit and
build time (if those were provided using ldflags). For example:

`$ go build -ldflags "-X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.commit=`git rev-parse HEAD` -w -s"`

Main uses a signal channel to listen for OS signals and will gracefully
wait for connections to drain and then stop when it receives a signal
to stop, or it will force a shutdown in 5 seconds.

It also has an error channel where it recieves errors from the
http server and log them.
*/
package main
