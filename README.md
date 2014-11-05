# suck

Suck is a command-line utility that eats c̶h̶i̶l̶d̶r̶e̶n̶ websites. It's basically a
naive web crawler that shoves what it finds into Neo4j.

## Getting Started

Prerequisites:

* [Docker](http://www.docker.com/)
* [Fig](http://orchardup.github.io/fig/index.html)

Start the database:

    $ fig up

Download dependencies:

    $ go get

Build suck:

    $ go build

Start suck:

    $ ./suck [options]

See help:

    $ ./suck -help

## License

The MIT License (MIT)

Copyright (c) 2014 Zach Latta

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
