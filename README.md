logmunch
========

Install with `go install github.com/msiebuhr/logmunch/bin/logmunch`.

Somewhat ad-hoc tool for parsing logs in logfmt and json-styles from
logentries.

Create a file, `~/.logmunch` with a list of providers and their default
parameters described as URLs. In particular, logentries' key for
`pull.logentries.com`:

    logentries://:xxxxxxxx-xxxx-xxxx-xxxxxxxxxxxxxxxxx@pull.logentries.com/

After that, running `logmunch -source logentries:Test/heroku -filter "H12"
-limit 10` will fetch ten entries with the text `H12` from `hosts/Test/heroku`
in logentries.

`logenries:`
 * https://logentries.com/doc/api-download/

`file:`
 * Defaults to stdin.
 * `file:///./local.txt` to read local files.

Basic use
---------

(I'm not terribly good at keeping README's up to date, so please run `logmunch
-h` to get a current overview.)

    // Parse a local file
    logmunch -source=file:/./local.log

    // Round timestamps and generate compound key X from A and B
    logmunch -source=- \
        -round-time=1h \
        -compound=X,A,B

Developer docs
--------------

The internal API used for fetching/parsing/filtering/outputting logs is split
out from the main binary, so it should be possible to re-use.

Documentation is at [godoc.org](http://godoc.org/github.com/msiebuhr/logmunch).

License
-------

ISC
