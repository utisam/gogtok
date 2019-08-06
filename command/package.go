package command

import (
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirkon/goproxy/gomod"
	"github.com/spf13/cobra"
)

func newPackage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Show package information",
	}

	cmd.AddCommand(newPackageName())
	cmd.AddCommand(newPackagePath())

	return cmd
}

func newPackageName() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "name",
		Short: "Show package name",
		RunE: func(_ *cobra.Command, args []string) error {
			dirs := args
			if len(dirs) == 0 {
				dirs = []string{"."}
			}

			for _, dir := range dirs {
				packageName, err := getPackageName(dir)
				if err != nil {
					return err
				}

				fmt.Println(packageName)
			}
			return nil
		},
	}
	return cmd
}

func getPackageName(dir string) (string, error) {
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, fileInfo := range fileInfos {
		fileName := path.Join(dir, fileInfo.Name())
		if fileInfo.IsDir() || !strings.HasSuffix(fileName, ".go") {
			continue
		}

		reader, err := os.Open(fileName)
		if err != nil {
			return "", err
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, fileName, reader, parser.PackageClauseOnly)
		if err != nil {
			return "", err
		}

		packageName := f.Name.Name
		if strings.HasSuffix(packageName, "_test") {
			continue
		}

		return packageName, nil
	}

	return "", fmt.Errorf("no .go files: %s", dir)
}

func newPackagePath() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Show package path",
		RunE: func(_ *cobra.Command, args []string) error {
			dirs := args
			if len(dirs) == 0 {
				dirs = []string{"."}
			}

			for _, dir := range dirs {
				packageName, err := getPackagePath(dir)
				if err != nil {
					return err
				}

				fmt.Println(packageName)
			}
			return nil
		},
	}
	return cmd
}

func getPackagePath(dir string) (string, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	goSRCPath := path.Join(gopath, "src")

	go111Module := os.Getenv("GO111MODULE")
	if go111Module != "on" && go111Module != "off" {
		if strings.HasPrefix(absPath, goSRCPath) {
			go111Module = "off"
		} else {
			go111Module = "on"
		}
	}

	if go111Module == "off" {
		if !strings.HasPrefix(absPath, goSRCPath) {
			return "", fmt.Errorf("%s is not in GOPATH", absPath)
		}

		return absPath[len(goSRCPath)+1:], nil
	}

	filePath, err := findGoModFile(absPath)
	if err != nil {
		return "", err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	mod, err := gomod.Parse(filePath, b)
	if err != nil {
		return "", err
	}

	return path.Join(mod.Name, dir), nil
}

func findGoModFile(absDir string) (string, error) {
	for {
		filePath := path.Join(absDir, "go.mod")

		info, err := os.Stat(filePath)
		if err == nil && !info.IsDir() {
			return filePath, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}

		absDir = path.Dir(absDir)
		if absDir == "/" {
			return "", os.ErrNotExist
		}
	}
}
