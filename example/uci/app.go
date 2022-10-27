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
	Bool     bool     `uci:"bool"`
	Int      int      `uci:"int"`
	Int32    int32    `uci:"int32"`
	Int64    int64    `uci:"int64"`
	Uint     uint     `uci:"uint"`
	Uint32   uint32   `uci:"uint32"`
	Uint64   uint64   `uci:"uint64"`
	Str      string   `uci:"str"`
	BoolList []bool   `uci:"bool_list"`
	IntList  []int    `uci:"int_list"`
	UintList []uint   `uci:"uint_list"`
	StrList  []string `uci:"str_list"`
	Inner    struct {
		Int8      int8    `uci:"int8"`
		Int16     int16   `uci:"int16"`
		Uint8     uint8   `uci:"uint8"`
		Uint16    uint16  `uci:"uint16"`
		Int32List []int32 `uci:"int32_list"`
	}
}

func marshalApi(ctx *openwrt.UciContext) {
	pkg, err := ctx.LoadPackage("hierarchy")
	if err != nil {
		panic(err)
	}
	defer pkg.Unload()

	b := &Bind{
		Bool:     true,
		Int:      1,
		Int32:    2,
		Int64:    3,
		Uint:     4,
		Uint32:   5,
		Uint64:   6,
		Str:      "string",
		BoolList: []bool{true, false},
		IntList:  []int{7, 8},
		UintList: []uint{9, 10},
		StrList:  []string{"string1", "string2"},
	}

	b.Inner.Int8 = 11
	b.Inner.Int16 = 12
	b.Inner.Uint8 = 13
	b.Inner.Uint16 = 14
	b.Inner.Int32List = []int32{15, 16}

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

func saveData(ctx *openwrt.UciContext) {
	client, err := openwrt.NewUciClient(ctx, "ng_test")
	if err != nil {
		panic(err)
	}
	defer client.Free()

	b := &Bind{
		Bool:     true,
		Int:      1,
		Int32:    2,
		Int64:    3,
		Uint:     4,
		Uint32:   5,
		Uint64:   6,
		Str:      "string",
		BoolList: []bool{true, false},
		IntList:  []int{7, 8},
		UintList: []uint{9, 10},
		StrList:  []string{"string1", "string2"},
	}

	b.Inner.Int8 = 11
	b.Inner.Int16 = 12
	b.Inner.Uint8 = 13
	b.Inner.Uint16 = 14
	b.Inner.Int32List = []int32{15, 16}

	fragment := openwrt.UciFragment{
		SectionName: "fragment_name",
		SectionType: "fragment_type",
		Content:     b,
	}

	client.Save(&fragment)

	var cmd openwrt.UciCommand
	cmd = &openwrt.UciCmd_AddSection{
		SectionName: "section_name_01",
		SectionType: "section_type_01",
	}
	client.Exec(cmd)

	cmd = &openwrt.UciCmd_SetOption{
		SectionName: "section_name_01",
		OptionName:  "option_01",
		OptionValue: "value_01",
	}
	client.Exec(cmd)

	cmd = &openwrt.UciCmd_SetOption{
		SectionName: "section_name_01",
		OptionName:  "option_02",
		OptionValue: "value_02",
	}
	client.Exec(cmd)

	cmd = &openwrt.UciCmd_AddListOption{
		SectionName:  "section_name_01",
		OptionName:   "option_03",
		OptionValue:  "list_value_01",
		OptionValues: []string{"list_value_02", "list_value_03"},
	}
	client.Exec(cmd)

	cmd = &openwrt.UciCmd_AddListOption{
		SectionName:  "section_name_01",
		OptionName:   "option_04",
		OptionValue:  "list_value_01",
		OptionValues: []string{"list_value_02", "list_value_03"},
	}
	client.Exec(cmd)

	addCmd := &openwrt.UciCmd_AddSection{
		SectionType: "section_type_02",
	}
	client.Exec(addCmd)
	section := addCmd.Section

	cmd = &openwrt.UciCmd_SetOption{
		Section:     section,
		OptionName:  "option_01",
		OptionValue: "value_01",
	}
	client.Exec(cmd)

	cmd = &openwrt.UciCmd_SetOption{
		Section:     section,
		OptionName:  "option_02",
		OptionValue: "value_02",
	}
	client.Exec(cmd)

	cmd = &openwrt.UciCmd_AddListOption{
		Section:      section,
		OptionName:   "option_03",
		OptionValue:  "list_value_01",
		OptionValues: []string{"list_value_02", "list_value_03"},
	}
	client.Exec(cmd)

	cmd = &openwrt.UciCmd_AddListOption{
		Section:      section,
		OptionName:   "option_04",
		OptionValue:  "list_value_01",
		OptionValues: []string{"list_value_02", "list_value_03"},
	}
	client.Exec(cmd)

	cmd = &openwrt.UciCmd_AddSection{
		SectionName: "section_name_03",
		SectionType: "section_type_03",
	}
	client.Exec(cmd)
}

func delData(ctx *openwrt.UciContext) {
	client, err := openwrt.NewUciClient(ctx, "ng_test")
	if err != nil {
		panic(err)
	}
	defer client.Free()

	var cmd openwrt.UciCommand
	cmd = &openwrt.UciCmd_DelSection{
		SectionName: "section_name_03",
	}
	client.Exec(cmd)

	addCmd := &openwrt.UciCmd_AddSection{
		SectionType: "section_type_03",
	}
	client.Exec(cmd)

	cmd = &openwrt.UciCmd_DelSection{
		Section: addCmd.Section,
	}
	client.Exec(cmd)

	sections := client.QuerySectionByType("section_type_01")
	for _, s := range sections {
		cmd = &openwrt.UciCmd_DelOption{
			SectionName: s.Name,
			OptionName:  "option_02",
		}
		client.Exec(cmd)

		cmd = &openwrt.UciCmd_DelFromList{
			SectionName: s.Name,
			OptionName:  "option_03",
			OptionValue: "list_value_02",
		}
		client.Exec(cmd)
	}
}

func queryData(ctx *openwrt.UciContext) {
	client, err := openwrt.NewUciClient(ctx, "ng_test")
	if err != nil {
		panic(err)
	}
	defer client.Free()

	b := &Bind{}
	fragment := openwrt.UciFragment{
		SectionName: "fragment_name",
		SectionType: "fragment_type",
		Content:     b,
	}

	client.Load(&fragment)

	fmt.Println("load fragment:")
	fmt.Println(fragment.Content)

	sections := client.QuerySectionByOption("option_01", "value_01")
	fmt.Println("")
	fmt.Println("sections by option:")
	fmt.Println(sections)

	sections = client.QuerySectionByTypeAndOption("fragment_type", "option_01", "value_01")
	fmt.Println("")
	fmt.Println("sections by type and option:")
	fmt.Println(sections)
}

func queryNetwork() {
	client, err := openwrt.NewUciClient(nil, "firewall")
	if err != nil {
		panic(err)
	}
	defer client.Free()

	sections := client.QuerySectionByTypeAndOption("zone", "input", "ACCEPT")
	fmt.Println(sections)

	for _, s := range sections {
		options := s.ListOptions()
		for _, o := range options {
			fmt.Println(o.Name, o.Value, o.Values)
		}
	}
}

func main() {
	// basic apiuci
	// ctx := openwrt.NewUciContext()
	// addByHierarchyApi(ctx)
	// loadByHierarchyApi(ctx)
	// editSectionByHierarchyApi(ctx)
	// addByHierarchyApi(ctx)
	// editOptionByHierarchyApi(ctx)
	// marshalApi(ctx)
	// unmarshalApi(ctx)
	// ctx.Free()

	// advanced api

	// ctx := openwrt.NewUciContext()
	// defer ctx.Free()

	// ctx.DelPackage("ng_test")

	// saveData(ctx)

	// delData(ctx)

	// queryData(ctx)

	queryNetwork()

	// checkResult()
}
