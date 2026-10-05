package jenkins

import (
	"encoding/xml"
	"fmt"
	"regexp"
)

type folder struct {
	XMLName       xml.Name         `xml:"com.cloudbees.hudson.plugins.folder.Folder"`
	Description   string           `xml:"description"`
	DisplayName   string           `xml:"displayName,omitempty"`
	Properties    folderProperties `xml:"properties"`
	FolderViews   xmlRawProperty   `xml:"folderViews"`
	HealthMetrics xmlRawProperty   `xml:"healthMetrics"`
}

type folderProperties struct {
	Security *folderSecurity  `xml:"com.cloudbees.hudson.plugins.folder.properties.AuthorizationMatrixProperty,omitempty"`
	Other    []xmlRawProperty `xml:",any"`
}

type folderSecurity struct {
	InheritanceStrategy folderPermissionInheritanceStrategy `xml:"inheritanceStrategy"`
	Permission          []string                            `xml:"permission"`
}

type folderPermissionInheritanceStrategy struct {
	Class string `xml:"class,attr"`
}

type xmlRawProperty struct {
	XMLName xml.Name
	Plugin  string `xml:"plugin,attr,omitempty"`
	Raw     string `xml:",innerxml"`
}

func parseFolder(config string) (*folder, error) {
	ret := &folder{}

	doc := handleXml(config)
	if err := xml.Unmarshal(doc, &ret); err != nil {
		return ret, fmt.Errorf("could not parse job XML: %w", err)
	}

	return ret, nil
}

func (j *folder) Render() ([]byte, error) {
	return xml.MarshalIndent(j, "", "\t")
}

// Go has no XML 1.1 support and rejects the declaration outright. Jenkins emits 1.1,
// single-quoted on disk but double-quoted over the REST API, so match either form.
// Safe as long as Jenkins uses no 1.1-only constructs.
var xmlVersion11Declaration = regexp.MustCompile(`^(\s*<\?xml\s[^>]*?version=["'])1\.1(["'])`)

func handleXml(def string) []byte {
	return []byte(xmlVersion11Declaration.ReplaceAllString(def, "${1}1.0${2}"))
}
