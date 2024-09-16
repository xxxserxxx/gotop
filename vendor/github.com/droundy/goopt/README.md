goopt
-----

A getopt-like processor of command-line flags.  It works much like the
"flag" package, only it processes arguments in a way that is
compatible with the
[getopt_long](http://www.gnu.org/s/libc/manual/html_node/Argument-Syntax.html#Argument-Syntax)
syntax, which is an extension of the syntax recommended by POSIX.

Example program
---------------

The example program named `example/example.go`, is meant to be more
useful for someone trying to see how the package works.  It comes with
a makefile demonstrating how to do some nice tricks, such as enabling
bash completion on your flags, and generating man pages and html
versions of the man page (see
[the man page for the example program](http://github.com/droundy/goopt/blob/master/example/goopt-example.html)).

Test suite
----------

The test suite is the file `.test`, which compiles and verifies the
output of the program in the `test-program` directory, as well as the
`example` program.  You can configure git to run the test suite
automatically by running the `setup-git.sh` script.

Documentation
-------------

Once the package is installed via goinstall, use the following to view
the documentation:

  # godoc --http=:6060

If you installed it from github, you will want to do this from the
source directory:

  # godoc --http=:6060 --path=.

This will run in the foreground, so do it in a terminal without
anything important in it.  Then you can go to http://localhost:6060/
and navigate via the package directory to the documentation or the
left-hand navigation, depending on if it was goinstalled or run from a
git clone.
