logmunch
========

Install with `go install github.com/msiebuhr/logmunch/bin/logmunch`.

Somewhat ad-hoc tool for parsing logs in logfmt and json-styles from logentries.

Create a file, `.logmunch` with a list of providers and their default
parameters described as URLs. In particular, logentries' key for
`pull.logentries.com`:

    logentries://:xxxxxxxx-xxxx-xxxx-xxxxxxxxxxxxxxxxx@pull.logentries.com/

After that, running `logmunch -source logentries:Test/heroku -filter "H12"
-limit 10` will fetch ten entries with the text `H12` from `hosts/Test/heroku`
in logentries.

Developer docs
--------------

See http://godoc.org/github.com/msiebuhr/logmunch

License
-------

ISC
