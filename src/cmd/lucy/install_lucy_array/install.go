package install_lucy_array

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type InstallLucyArray struct {
}

func (installLucyArray *InstallLucyArray) Help(command string) {

}

type InstallType struct {
	className    string
	typename     string
	defaultValue string
	imports      string
}

var (
	installs = []*InstallType{}
)

func init() {
	installs = append(installs, &InstallType{
		className: "ArrayBool",
		typename:  "boolean",
	})
	installs = append(installs, &InstallType{
		className: "ArrayByte",
		typename:  "byte",
	})
	installs = append(installs, &InstallType{
		className: "ArrayShort",
		typename:  "short",
	})
	installs = append(installs, &InstallType{
		className: "ArrayInt",
		typename:  "int",
	})
	installs = append(installs, &InstallType{
		className: "ArrayLong",
		typename:  "long",
	})
	installs = append(installs, &InstallType{
		className: "ArrayFloat",
		typename:  "float",
	})
	installs = append(installs, &InstallType{
		className: "ArrayDouble",
		typename:  "double",
	})
	installs = append(installs, &InstallType{
		className: "ArrayObject",
		typename:  "Object",
		imports: `
		import java.lang.Object;
`,
	})

	installs = append(installs, &InstallType{
		className: "ArrayString",
		typename:  "String",
	})
}

func (installLucyArray *InstallLucyArray) RunCommand(command string, args []string) {
	path := os.Getenv("LUCYROOT")
	if path == "" {
		fmt.Println("env variable LUCYPATH is not set")
		os.Exit(1)
	}
	dest := filepath.Join(path, "lib/lucy/deps")
	os.MkdirAll(dest, 0755) // ignore errors
	err := os.Chdir(dest)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, v := range installs {
		javaFile := v.className + ".java"
		t := strings.Replace(array_template, "ArrayTTT", v.className, -1)
		t = strings.Replace(t, "TTT", v.typename, -1)
		t = strings.Replace(t, "DEFAULT_INIT", v.defaultValue, -1)
		t = strings.Replace(t, "IMPORTS", v.imports, -1)
		err := ioutil.WriteFile(javaFile, []byte(t), 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		cmd := exec.Command("javac", javaFile)
		out, err := cmd.Output()
		if err != nil {
			fmt.Println(err)
			if out != nil && len(out) > 0 {
				os.Stdout.Write(out)
			}
			os.Exit(3)
		}
	}
}
