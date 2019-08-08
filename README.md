# gogtok

Toolkit for shell script used in `go:generate`.

## Getting started

```bash
go get -u github.com/utisam/gogtok/cmd/gogtok
```

1. `gogtok new FILENAME` to generate a new script file
2. Add `//go:generate ./filename.sh` in your code.
3. Write script to render source

## Commands

* `glue`: Generate glue code
* `import [packages...]`: Generate import statement
* `list funcs/values/types [files...]`: Parse source files and show declarations
* `list fields [file] [name]`: Parse source files and show fields
* `new [name]`: Generate new script from boilerplate
* `package name [dir]`: Show package name of the directory
* `package path [dir]`: Show package path of the directory

## Examples

### Specify `text/template`

`goimports` cannot resolve package which has the same name with other packages.
For example, `template` is matched with `text/template` and `html/template`,
then you must specify package.

```bash
render_import() {
    echo
    gogtok import \
        "text/template" \
        "crypto/rand" \
        "$(gogtok package path ..)/config"
}
```

`gogtok import` generates a import block only if there are multiple packages.
`gogtok package` rsespects `GO111MODULE` and parse `go.mod` in the parent directory.

### Generate glue code

```bash
render_convert() {
    echo
    cat <<-EOS
    func convert(a *OtherStruct) *SomeStruct {
        return &SomeStruct{
    EOS
    gogtok list fields some.go SomeStruct | xargs gogtok glue a
    cat <<-EOS
        }
    }
    EOS
}
```

This function generates following code when `SomeStruct` has 4 fields `W`, `X`, `Y`, `Z`:

```go
func convert(a *OtherStruct) *SomeStruct {
    return *SomeStruct{
        W: a.W,
        X: a.X,
        Y: a.Y,
        Z: a.Z,
    }
}
```

`gogtok glue` has options to customize glue code. See `gogtok help glue`.

### List all `reflect.Kind`

```bash
source <(go env)
gogtok list values \
    --filter-decl-type Kind \
    "${GOROOT}/src/reflect/type.go"
```

### List name, type and tags

`name`, `type`, `tags` and `tag[key]` can be used to `--columns` option.

```bash
gogtok list fields --print0 --columns 'name,type,tag[json]' file.go SomeStruct | \
    while IFS= read -r -d '' name && read -r -d '' type && read -r -d '' tag_json
do
    echo "name: $name, type: $type, tag[json]: $tag_json"
done
```

`--print0` is usefull when you want to handle types and/or tags which contain space.
Like `find -print0`, NULL chars are used to output.
`xargs -0` can be used to pass the regular commands.
