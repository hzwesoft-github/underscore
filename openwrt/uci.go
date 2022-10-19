package openwrt

type UciClient struct {
	Context *UciContext
	Package *UciPackage

	shouldCommit bool
}

func NewUciClient(packageName string) (*UciClient, error) {
	ctx := NewUciContext()

	pkg, err := ctx.AddPackage(packageName)
	if err != nil {
		ctx.Free()
		return nil, err
	}

	return &UciClient{ctx, pkg, false}, nil
}

func (client *UciClient) Free() {
	defer client.Context.Free()
	defer client.Package.Unload()

	client.Package.Commit(false)
}

func (client *UciClient) Save(fragment *UciFragment) error {
	// TODO
	return nil
}

func (client *UciClient) Exec(command *UciCommand) error {
	// TODO
	return nil
}

func (client *UciClient) LoadSectionByName(name string) *UciSection {
	return client.Package.LoadSection(name)
}

func (client *UciClient) QuerySectionByType(typ string) []UciSection {
	return client.Package.QuerySection(func(section *UciSection) bool {
		return section.Type == typ
	})
}

func (client *UciClient) QuerySectionIncludeOption(name string, value string) []UciSection {
	return client.Package.QuerySection(func(section *UciSection) bool {
		option := section.LoadOption(name)
		if option == nil {
			return false
		}

		switch option.Type {
		case UCI_TYPE_STRING:
			return option.Value == value
		case UCI_TYPE_LIST:
			for _, v := range option.Values {
				if v == value {
					return true
				}
			}

			return false
		default:
			return false
		}
	})
}

func (client *UciClient) QuerySection(cb SectionFilter) []UciSection {
	return client.Package.QuerySection(cb)
}

type UciCommand interface {
	Exec(client *UciClient) error
}

type UciCmd_AddSection struct {
	Section     *UciSection
	SectionName string
	SectionType string
}

func (c *UciCmd_AddSection) Exec(client *UciClient) error {
	// TODO
	return nil
}

type UciCmd_DelSection struct {
	Section     *UciSection
	SectionName string
	SectionType string
}

func (c *UciCmd_DelSection) Exec(client *UciClient) error {
	// TODO
	return nil
}

type UciCmd_SetOption struct {
}

func (c *UciCmd_SetOption) Exec(client *UciClient) error {
	// TODO
	return nil
}

type UciCmd_AddListOption struct {
}

func (c *UciCmd_AddListOption) Exec(client *UciClient) error {
	// TODO
	return nil
}

type UciCmd_DelOption struct {
}

func (c *UciCmd_DelOption) Exec(client *UciClient) error {
	// TODO
	return nil
}

type UciCmd_DelFromList struct {
}

func (c *UciCmd_DelFromList) Exec(client *UciClient) error {
	// TODO
	return nil
}

// UCI Fragment
type UciFragment struct {
	Section     *UciSection
	SectionName string
	SectionType string
	Content     any
}
