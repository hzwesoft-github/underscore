package openwrt

import (
	"errors"
	"fmt"

	"github.com/hzwesoft-github/underscore/lang"
)

type UciClient struct {
	Context *UciContext
	Package *UciPackage

	shouldCommit  bool
	externContext bool
}

func NewUciClient(context *UciContext, packageName string) (*UciClient, error) {
	externCtx := context != nil
	if context == nil {
		context = NewUciContext()
	}

	pkg, err := context.AddPackage(packageName)
	if err != nil {
		context.Free()
		return nil, err
	}

	return &UciClient{context, pkg, false, externCtx}, nil
}

func (client *UciClient) Flush() error {
	if !client.shouldCommit {
		return nil
	}
	return client.Package.Commit(false)
}

func (client *UciClient) Free() {
	if !client.externContext {
		defer client.Context.Free()
	}

	defer client.Package.Unload()

	if !client.shouldCommit {
		return
	}

	client.Package.Commit(false)
}

func (client *UciClient) Remove() error {
	return client.Context.DelPackage(client.Package.Name)
}

func (client *UciClient) Save(fragment *UciFragment) error {
	if fragment.Section == nil && lang.IsBlank(fragment.SectionType) {
		return errors.New("ng: fragment section must be specified")
	}

	if fragment.Section != nil {
		if err := client.Package.MarshalSection(fragment.Section, fragment.Content, false); err != nil {
			return err
		}
	} else {
		if err := client.Package.Marshal(fragment.SectionName, fragment.SectionType, fragment.Content, false); err != nil {
			return err
		}
	}

	client.shouldCommit = true
	return nil
}

func (client *UciClient) Load(fragment *UciFragment) error {
	if fragment.Section == nil && lang.IsBlank(fragment.SectionName) {
		return errors.New("ng: fragment section must be specified")
	}

	if fragment.Section != nil {
		return client.Package.UnmarshalSection(fragment.Section, fragment.Content)
	} else {
		return client.Package.Unmarshal(fragment.SectionName, fragment.Content)
	}
}

func (client *UciClient) Exec(command UciCommand) error {
	return command.Exec(client)
}

func (client *UciClient) LoadSectionByName(name string) *UciSection {
	return client.Package.LoadSection(name)
}

func (client *UciClient) QuerySectionByType(typ string) []UciSection {
	return client.Package.QuerySection(func(section *UciSection) bool {
		return section.Type == typ
	})
}

func (client *UciClient) QueryOneByOption(name string, value string) *UciSection {
	sections := client.QuerySectionByOption(name, value)
	if len(sections) == 0 {
		return nil
	}

	return &sections[0]
}

func (client *UciClient) QuerySectionByOption(name string, value string) []UciSection {
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

func (client *UciClient) QueryOneByTypeAndOption(typ string, name string, value string) *UciSection {
	sections := client.QuerySectionByTypeAndOption(typ, name, value)
	if len(sections) == 0 {
		return nil
	}

	return &sections[0]
}

func (client *UciClient) QuerySectionByTypeAndOption(typ string, name string, value string) []UciSection {
	return client.Package.QuerySection(func(section *UciSection) bool {
		if section.Type != typ {
			return false
		}

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
	if lang.IsBlank(c.SectionType) {
		return errors.New("ng: section type must be specified")
	}

	if lang.IsBlank(c.SectionName) {
		section, err := client.Package.AddUnnamedSection(c.SectionType)
		if err != nil {
			return err
		}

		c.Section = section
	} else {
		if err := client.Package.AddSection(c.SectionName, c.SectionType); err != nil {
			return err
		}

		c.Section = client.Package.LoadSection(c.SectionName)
	}

	client.shouldCommit = true
	return nil
}

type UciCmd_DelSection struct {
	Section     *UciSection
	SectionName string
}

func (c *UciCmd_DelSection) Exec(client *UciClient) error {
	if c.Section == nil && lang.IsBlank(c.SectionName) {
		return errors.New("ng: cmd section must be specified")
	}

	if c.Section != nil {
		if c.Section.Anonymous {
			if err := client.Package.DelUnnamedSection(c.Section); err != nil {
				return err
			}
		} else {
			if err := client.Package.DelSection(c.Section.Name); err != nil {
				return err
			}
		}
	} else {
		if err := client.Package.DelSection(c.SectionName); err != nil {
			return err
		}
	}

	client.shouldCommit = true
	return nil
}

type UciCmd_SetOption struct {
	Section     *UciSection
	SectionName string
	OptionName  string
	OptionValue string
}

func (c *UciCmd_SetOption) Exec(client *UciClient) error {
	if c.Section == nil && lang.IsBlank(c.SectionName) {
		return errors.New("ng: cmd section must be specified")
	}
	if lang.IsBlank(c.OptionName) {
		return errors.New("ng: option name must be specified")
	}

	if c.Section != nil {
		if err := c.Section.SetStringOption(c.OptionName, c.OptionValue); err != nil {
			return err
		}
	} else {
		section := client.Package.LoadSection(c.SectionName)
		if section == nil {
			return fmt.Errorf("ng: section %s is not exist", c.SectionName)
		}
		return section.SetStringOption(c.OptionName, c.OptionValue)
	}

	client.shouldCommit = true
	return nil
}

type UciCmd_AddListOption struct {
	Section      *UciSection
	SectionName  string
	OptionName   string
	OptionValue  string
	OptionValues []string
}

func (c *UciCmd_AddListOption) Exec(client *UciClient) error {
	if c.Section == nil && lang.IsBlank(c.SectionName) {
		return errors.New("ng: cmd section must be specified")
	}
	if lang.IsBlank(c.OptionName) {
		return errors.New("ng: option name must be specified")
	}

	if c.Section != nil {
		if c.OptionValue != "" {
			if err := c.Section.AddListOption(c.OptionName, c.OptionValue); err != nil {
				return err
			}
		}
		if len(c.OptionValues) >= 0 {
			if err := c.Section.AddListOption(c.OptionName, c.OptionValues...); err != nil {
				return err
			}
		}
	} else {
		section := client.Package.LoadSection(c.SectionName)
		if section == nil {
			return fmt.Errorf("ng: section %s is not exist", c.SectionName)
		}

		if c.OptionValue != "" {
			if err := section.AddListOption(c.OptionName, c.OptionValue); err != nil {
				return err
			}
		}
		if len(c.OptionValues) >= 0 {
			if err := section.AddListOption(c.OptionName, c.OptionValues...); err != nil {
				return err
			}
		}
	}

	client.shouldCommit = true
	return nil
}

type UciCmd_DelOption struct {
	Section     *UciSection
	SectionName string
	OptionName  string
}

func (c *UciCmd_DelOption) Exec(client *UciClient) error {
	if c.Section == nil && lang.IsBlank(c.SectionName) {
		return errors.New("ng: cmd section must be specified")
	}
	if lang.IsBlank(c.OptionName) {
		return errors.New("ng: option name must be specified")
	}

	if c.Section != nil {
		if err := c.Section.DelOption(c.OptionName); err != nil {
			return err
		}
	} else {
		section := client.Package.LoadSection(c.SectionName)
		if section == nil {
			return fmt.Errorf("ng: section %s is not exist", c.SectionName)
		}

		if err := section.DelOption(c.OptionName); err != nil {
			return err
		}
	}

	client.shouldCommit = true
	return nil
}

type UciCmd_DelFromList struct {
	Section     *UciSection
	SectionName string
	OptionName  string
	OptionValue string
}

func (c *UciCmd_DelFromList) Exec(client *UciClient) error {
	if c.Section == nil && lang.IsBlank(c.SectionName) {
		return errors.New("ng: cmd section must be specified")
	}
	if lang.IsBlank(c.OptionName) {
		return errors.New("ng: option name must be specified")
	}

	if c.Section != nil {
		if err := c.Section.DelFromList(c.OptionName, c.OptionValue); err != nil {
			return err
		}
	} else {
		section := client.Package.LoadSection(c.SectionName)
		if section == nil {
			return fmt.Errorf("ng: section %s is not exist", c.SectionName)
		}

		if err := section.DelFromList(c.OptionName, c.OptionValue); err != nil {
			return err
		}
	}

	client.shouldCommit = true
	return nil
}

// UCI Fragment
type UciFragment struct {
	Section     *UciSection
	SectionName string
	SectionType string
	Content     any
}
