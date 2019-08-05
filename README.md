# gogtok

Toolkit for shell script used in `go:generate`.

## Getting started

```sh
go get -u github.com/utisam/gogtok/cmd/gogtok
gogtok new filename > filename.sh
chmod 755 filename.sh
```

and add `//go:generate ./filename.sh` in your code.

## Commands

* `glue`: Generate glue code
* `import [packages...]`: Generate import statement
* `list funcs/values/types [files...]`: Parse source files and show declarations
* `list fields [file] [name]`
* `new [name]`: Generate new script from boilerplate

TODO:

* `package name [dir]`
* `package path [dir]`
