package forms_test

import (
	"os"
	"path/filepath"

	"github.com/mevansam/goforms/forms"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	test_data "github.com/mevansam/goforms/test/data"
)

var _ = Describe("Input Fields", func() {

	var (
		err error
		ic  *forms.InputCollection
		ig  *forms.InputGroup
	)

	BeforeEach(func() {
		ic = test_data.NewTestInputCollection()
		ig = ic.Group("input-form")
	})

	Context("value retrieval from file", func() {

		It("sources field value from a file with path sourced from environment", func() {

			var (
				attrib132Value string
				value          *string
			)

			field, err := ig.GetInputField("attrib132")
			Expect(err).NotTo(HaveOccurred())
			err = field.SetValueRef(&attrib132Value)
			Expect(err).NotTo(HaveOccurred())

			attrib132FilePath, err := filepath.Abs(workingDirectory + "/../test/fixtures/forms/attrib132")
			Expect(err).NotTo(HaveOccurred())
			os.Setenv("ATTRIB132", attrib132FilePath)

			valueFromFile, paths := field.ValueFromFile()
			Expect(valueFromFile).To(BeTrue())
			Expect(len(paths)).To(Equal(1))
			Expect(paths[0]).To(Equal(attrib132FilePath))

			err = field.SetValue(&paths[0])
			Expect(err).NotTo(HaveOccurred())
			value = field.Value()
			Expect(*value).To(Equal(`{"attrib132":"value for attrib132 from file"}`))
		})

		It("sources field value from a file with path sourced from environment", func() {

			var (
				attrib132Value string
				attrib133Value string
				value          *string
			)

			field, err := ig.GetInputField("attrib132")
			Expect(err).NotTo(HaveOccurred())
			err = field.SetValueRef(&attrib132Value)
			Expect(err).NotTo(HaveOccurred())

			field, err = ig.GetInputField("attrib133")
			Expect(err).NotTo(HaveOccurred())
			err = field.SetValueRef(&attrib133Value)
			Expect(err).NotTo(HaveOccurred())
			Expect(attrib133Value).To(Equal("default value for attrib133"))

			// hint is value of parsed json context of field 'attrib132' having key 'attrib132'
			err = ig.AddFieldValueHint("attrib133", "field://attrib132/attrib132")
			Expect(err).NotTo(HaveOccurred())

			attrib132FilePath, err := filepath.Abs(workingDirectory + "/../test/fixtures/forms/attrib132")
			Expect(err).NotTo(HaveOccurred())
			err = ig.SetFieldValue("attrib132", attrib132FilePath)
			Expect(err).NotTo(HaveOccurred())

			hintValues, err := ig.GetFieldValueHints("attrib133")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(hintValues)).To(Equal(1))

			err = ig.SetFieldValue("attrib133", hintValues[0])
			Expect(err).NotTo(HaveOccurred())
			value, err = ig.GetFieldValue("attrib133")
			Expect(err).NotTo(HaveOccurred())
			Expect(*value).To(Equal("value for attrib132 from file"))
		})
	})

	Context("input field validation", func() {

		BeforeEach(func() {

			// Bind fields to map of values so
			// that form values can be saved
			inputValues := make(map[string]*string)
			for _, f := range ig.InputFields() {
				s := new(string)
				inputValues[f.Name()] = s
				err = f.SetValueRef(s)
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("restricts field values to a list of accepted values", func() {

			var (
				field *forms.InputField
				value string
			)

			field, err = ig.GetInputField("attrib11")
			Expect(err).NotTo(HaveOccurred())

			acceptedValues := []string{"aa", "bb", "cc"}
			field.SetAcceptedValues(&acceptedValues, "error")

			value = "dd"
			err = field.SetValue(&value)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))

			value = "bb"
			err = field.SetValue(&value)
			Expect(err).ToNot(HaveOccurred())
		})

		It("validates field values using an inclusion filter", func() {

			var (
				field *forms.InputField
				value string
			)

			field, err = ig.GetInputField("attrib11")
			Expect(err).NotTo(HaveOccurred())

			err = field.SetInclusionFilter("(gopher){2}", "error")
			Expect(err).ToNot(HaveOccurred())

			value = "gopher"
			err = field.SetValue(&value)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))

			value = "gophergophergopher"
			err = field.SetValue(&value)
			Expect(err).ToNot(HaveOccurred())
		})

		It("validates field values using an exclusion filter", func() {

			var (
				field *forms.InputField
				value string
			)

			field, err = ig.GetInputField("attrib11")
			Expect(err).NotTo(HaveOccurred())

			err = field.SetExclusionFilter("(gopher){2}", "error")
			Expect(err).ToNot(HaveOccurred())

			value = "gophergophergopher"
			err = field.SetValue(&value)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))

			value = "gopher"
			err = field.SetValue(&value)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
