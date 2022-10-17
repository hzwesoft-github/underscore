package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/hzwesoft-github/underscore/openwrt"
)

func printSections(sections ...openwrt.UciSection) {
	for _, o := range sections {
		fmt.Println(o.Name, o.Type, o.Anonymous)
	}
}

func printOptions(options ...openwrt.UciOption) {
	for _, o := range options {
		switch o.Type {
		case openwrt.UCI_TYPE_STRING:
			fmt.Println(o.Name, o.Type, o.Value)
		case openwrt.UCI_TYPE_LIST:
			fmt.Println(o.Name, o.Type, o.Values)
		}
	}
}

func checkResult() {
	cmd := exec.Command("lsof", "-p", strconv.Itoa(os.Getpid()))
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	fmt.Println("")
	fmt.Println("opened fds: ")
	fmt.Println(string(output))
}

// 层级API
func addByHierarchyApi(ctx *openwrt.UciContext) {
	ctx.DelPackage("hierarchy")

	pkg, err := ctx.AddPackage("hierarchy")
	if err != nil {
		panic(err)
	}
	defer pkg.Unload()

	if err := pkg.AddSection("lan00", "lan"); err != nil {
		panic(err)
	}

	wanSection := pkg.LoadSection("lan00")
	wanSection.SetStringOption("name", "LAN1")
	wanSection.AddListOption("devices", "eth0", "eth1")
	wanSection.AddListOption("devices", "eth2")

	deviceSection, err := pkg.AddUnnamedSection("device")
	if err != nil {
		panic(err)
	}

	deviceSection.SetStringOption("name", "br-lan")
	deviceSection.AddListOption("ports", "eth0")
	deviceSection.AddListOption("ports", "eth1")

	pkg.Commit(false)
}

func loadByHierarchyApi(ctx *openwrt.UciContext) {
	pkg, err := ctx.LoadPackage("hierarchy")
	if err != nil {
		panic(err)
	}
	defer pkg.Unload()

	fmt.Println("")
	fmt.Println("load section lan00")
	wanSection := pkg.LoadSection("lan00")

	option1 := wanSection.LoadOption("name")
	option2 := wanSection.LoadOption("devices")

	printOptions(*option1, *option2)

	fmt.Println("")
	fmt.Println("list section lan00")
	options := wanSection.ListOptions()
	printOptions(options...)

	fmt.Println("")
	fmt.Println("list all sections")
	sections := pkg.ListSections()
	printSections(sections...)

	fmt.Println("")
	fmt.Println("list unnamed section options")
	fmt.Println("can't list")
	options = sections[1].ListOptions()
	printOptions(options...)
}

func editSectionByHierarchyApi(ctx *openwrt.UciContext) {
	pkg, err := ctx.LoadPackage("hierarchy")
	if err != nil {
		panic(err)
	}
	defer pkg.Unload()

	sections := pkg.ListSections()

	for _, section := range sections {
		if section.Anonymous {
			fmt.Println("")
			fmt.Println("del unnamed section")
			pkg.DelSection(section.Name)
		} else {
			fmt.Println("")
			fmt.Println("del named section")
			pkg.DelSection(section.Name)
		}
	}

	fmt.Println("list all sections")
	sections = pkg.ListSections()
	printSections(sections...)

	pkg.Commit(false)

}

func editOptionByHierarchyApi(ctx *openwrt.UciContext) {
	pkg, err := ctx.LoadPackage("hierarchy")
	if err != nil {
		panic(err)
	}
	defer pkg.Unload()

	sections := pkg.ListSections()
	var name1, name2 string

	for _, section := range sections {
		if section.Anonymous {
			name1 = section.Name
		} else {
			name2 = section.Name
		}
	}

	section := pkg.LoadSection(name1)
	fmt.Println("")
	fmt.Println("edit unnamed section")
	section.DelFromList("ports", "eth1")
	section.DelOption("name")

	fmt.Println(section.LoadOption("name"))
	printOptions(*section.LoadOption("ports"))

	section = pkg.LoadSection(name2)
	fmt.Println("")
	fmt.Println("edit named section")
	section.DelFromList("devices", "eth1")
	section.DelOption("name")

	fmt.Println(section.LoadOption("name"))
	printOptions(*section.LoadOption("devices"))

	pkg.Commit(false)
}

type Bind struct {
	Str string   `uci:"str"`
	Arr []string `uci:"arr"`
}

func marshalApi(ctx *openwrt.UciContext) {
	pkg, err := ctx.LoadPackage("hierarchy")
	if err != nil {
		panic(err)
	}
	defer pkg.Unload()

	b := &Bind{
		Str: "111",
		Arr: []string{"aaa", "bbb"},
	}

	if err = pkg.Marshal("b", "bind", b, true); err != nil {
		fmt.Println(err)
	}

}

func unmarshalApi(ctx *openwrt.UciContext) {
	pkg, err := ctx.LoadPackage("hierarchy")
	if err != nil {
		panic(err)
	}
	defer pkg.Unload()

	b := &Bind{}

	if err = pkg.Unmarshal("b", b); err != nil {
		fmt.Println(err)
	}

	fmt.Println("")
	fmt.Println("unmarshal")
	fmt.Println(b)

}

func main() {
	ctx := openwrt.NewUciContext()

	addByHierarchyApi(ctx)

	loadByHierarchyApi(ctx)

	editSectionByHierarchyApi(ctx)

	addByHierarchyApi(ctx)

	editOptionByHierarchyApi(ctx)

	marshalApi(ctx)

	unmarshalApi(ctx)

	ctx.Free()

	checkResult()
}
