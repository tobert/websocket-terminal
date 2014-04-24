# websocket-terminal

Combines [term.js](https://github.com/chjj/term.js) with a
Go websocket backend to get something much less complicated
to deploy than [tty.js](https://github.com/chjj/tty.js).

# Security

There is none. Use at your own risk.

### Usage

It's a bit raw at the moment.

    go get github.com/gorilla/websocket
    go get github.com/kr/pty
    git clone https://github.com/tobert/websocket-terminal.git
    cd websocket-terminal
    go build
    ./websocket-terminal -cmd /bin/bash -addr :9000

### Limitations

* does not support non-ASCII/Latin1 characters at all
* horribly insecure
* probably slow

### Encoding Issues

The Go server and the javascript in index.html are currently
using base64 encoding for data moving across the wire. The
websocket is using text messages as well as a result.

This is quite fragile since any utf characters outside of
Latin1 will cause errors. I've tested it and Japanese kana
certainly break things, but for the most part shell tools seem
to work.

### Future

term.js needs to be refactored to use ArrayBuffers with a Uint8Array
view and some kind of character decoding from utf8 -> utf16 since
almost all POSIX systems use utf8 and javascript strings are utf16.
This will almost certainly be faster too.

### License

MIT
