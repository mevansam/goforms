package forms

import (
	"github.com/mevansam/goforms/forms"
	"github.com/mevansam/goutils/utils"

	. "github.com/onsi/gomega"
)

func NewTestInputCollection() *forms.InputCollection {

	var (
		err error
		ic  *forms.InputCollection
		ig  *forms.InputGroup
	)

	ic = forms.NewInputCollection()

	ig = ic.NewGroup("input-form", "test group description")
	ic.NewGroup("input-form2", "input form 2 description")
	ic.NewGroup("input-form3", "input form 3 description")

	// Input Paths (name group)
	//
	// attrib11 1 -> X
	// attrib12 1 -> attrib121 2 -> X
	//            -> attrib122 2 -> attrib1221 -> 0 X
	//            -> attrib131 0 -> attrib1311 -> 0 X
	//                           -> attrib1312 -> 0 X
	// attrib13 1 -> attrib131 0 -> attrib1311 -> 0 X
	//                           -> attrib1312 -> 0 X
	//            -> attrib132 3 -> X
	//            -> attrib133 3 -> X
	// attrib14 0 -> attrib141 0

	ig.NewInputContainer(
		/* name */ "group1",
		/* displayName */ "Group 1",
		/* description */ "description for group 1",
		/* groupId */ 1,
	)
	ig.NewInputContainer(
		/* name */ "group2",
		/* displayName */ "Group 2",
		/* description */ "description for group 2",
		/* groupId */ 2,
	)
	ig.NewInputContainer(
		/* name */ "group3",
		/* displayName */ "Group 3",
		/* description */ "description for group 3",
		/* groupId */ 3,
	)

	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:        "attrib11",
		DisplayName: "Attrib 11",
		Description: "description for attrib11.",
		GroupID:     1,
		InputType:   forms.String,
		EnvVars: []string{
			"ATTRIB11_ENV1",
			"ATTRIB11_ENV2",
			"ATTRIB11_ENV3",
		},
		Tags: []string{"tag1"},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:        "attrib12",
		DisplayName: "Attrib 12",
		Description: "description for attrib12.",
		GroupID:     1,
		InputType:   forms.String,
		EnvVars: []string{
			"ATTRIB12_ENV1",
		},
		Tags: []string{"tag1"},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:        "attrib13",
		DisplayName: "Attrib 13",
		Description: "description for attrib13.",
		GroupID:     1,
		InputType:   forms.String,
		EnvVars: []string{
			"ATTRIB13_ENV1",
			"ATTRIB13_ENV2",
		},
		Tags: []string{"tag2"},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:         "attrib14",
		DisplayName:  "Attrib 14",
		Description:  "description for attrib14.",
		InputType:    forms.String,
		DefaultValue: utils.PtrToStr("default value for attrib14"),
		EnvVars:      []string{},
		Tags:         []string{"tag1"},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:        "attrib121",
		DisplayName: "Attrib 121",
		Description: "description for attrib121.",
		GroupID:     2,
		InputType:   forms.String,
		EnvVars:     []string{},
		DependsOn:   []string{"attrib12=value for attrib12|value for attrib12 - A"},
		Tags:        []string{"tag1"},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:        "attrib122",
		DisplayName: "Attrib 122",
		Description: "description for attrib122.",
		GroupID:     2,
		InputType:   forms.String,
		EnvVars:     []string{},
		DependsOn:   []string{"attrib12=value for attrib12|value for attrib12 - B"},
		Tags:        []string{"tag1"},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:        "attrib131",
		DisplayName: "Attrib 131",
		Description: "description for attrib131.",
		InputType:   forms.String,
		EnvVars:     []string{},
		DependsOn:   []string{"attrib12", "attrib13"},
		Tags:        []string{"tag1"},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:          "attrib132",
		DisplayName:   "Attrib 132",
		Description:   "description for attrib132.",
		GroupID:       3,
		InputType:     forms.String,
		ValueFromFile: true,
		EnvVars:       []string{"ATTRIB132"},
		DependsOn:     []string{"attrib13"},
		Tags:          []string{"tag1"},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:         "attrib133",
		DisplayName:  "Attrib 133",
		Description:  "description for attrib133.",
		GroupID:      3,
		InputType:    forms.String,
		DefaultValue: utils.PtrToStr("default value for attrib133"),
		EnvVars:      []string{},
		DependsOn:    []string{"attrib13"},
		Tags:         []string{"tag1"},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:        "attrib141",
		DisplayName: "Attrib 141",
		Description: "description for attrib141.",
		InputType:   forms.String,
		EnvVars:     []string{},
		DependsOn:   []string{"attrib14=value for attrib14 - X"},
		Tags:        []string{"tag1"},
	})
	Expect(err).NotTo(HaveOccurred())

	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:        "attrib1221",
		DisplayName: "Attrib 1221",
		Description: "description for attrib1221.",
		InputType:   forms.String,
		EnvVars:     []string{},
		DependsOn:   []string{"attrib122"},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:        "attrib1311",
		DisplayName: "Attrib 1311",
		Description: "description for attrib1311.",
		InputType:   forms.String,
		EnvVars:     []string{},
		DependsOn:   []string{"attrib131"},
	})
	Expect(err).NotTo(HaveOccurred())
	_, err = ig.NewInputField(forms.FieldAttributes{
		Name:        "attrib1312",
		DisplayName: "Attrib 1312",
		Description: "description for attrib1312.",
		InputType:   forms.String,
		EnvVars:     []string{},
		DependsOn:   []string{"attrib131"},
	})
	Expect(err).NotTo(HaveOccurred())

	return ic
}
