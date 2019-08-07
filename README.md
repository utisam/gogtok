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
* `list fields [file] [name]`: Parse source files and show fields
* `new [name]`: Generate new script from boilerplate
* `package name [dir]`: Show package name of the directory
* `package path [dir]`: Show package path of the directory
