package forms_test

import (
	"github.com/mevansam/goforms/forms"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	test_data "github.com/mevansam/goforms/test/data"
)

var _ = Describe("Input Groups", func() {

	var (
		err error
		ic  *forms.InputCollection
		ig  *forms.InputGroup
	)

	BeforeEach(func() {
		ic = test_data.NewTestInputCollection()
		ig = ic.Group("input-form")
	})

	Context("input group data retrieval and updates", func() {

		It("can retrieve fields from group as expected", func() {

			var (
				field *forms.InputField
			)

			field, err = ig.GetInputField("attrib121")
			Expect(err).NotTo(HaveOccurred())
			Expect(field.Name()).To(Equal("attrib121"))

			field, err = ig.GetInputField("attrib1221")
			Expect(err).NotTo(HaveOccurred())
			Expect(field.Name()).To(Equal("attrib1221"))

			field, err = ig.GetInputField("attrib1312")
			Expect(err).NotTo(HaveOccurred())
			Expect(field.Name()).To(Equal("attrib1312"))

			_, err = ig.GetInputField("attrib1411")
			Expect(err).To(HaveOccurred())
		})

		It("gets and sets fields bound to an external data structure", func() {

			var (
				field *forms.InputField
				value *string

				newValue string
			)
			attrib11Value := "attrib11 #1"

			data := struct {
				attrib11 *string
				attrib12 string
				attrib13 *string
			}{
				attrib11: &attrib11Value,
				attrib12: "attrib12 #2",
				attrib13: nil,
			}

			field, err = ig.GetInputField("attrib11")
			Expect(err).NotTo(HaveOccurred())
			err = field.SetValueRef(&data.attrib11)
			Expect(err).NotTo(HaveOccurred())
			value = field.Value()
			Expect(*value).To(Equal("attrib11 #1"))

			// value update in struct should reflect
			// when retrieved via InputForm
			attrib11Value = "attrib11 #2"
			value = field.Value()
			Expect(*value).To(Equal("attrib11 #2"))

			// value update in input form
			// should reflect in struct
			newValue = "attrib11 #3"
			err = field.SetValue(&newValue)
			Expect(*data.attrib11).To(Equal("attrib11 #3"))

			field, err = ig.GetInputField("attrib12")
			Expect(err).NotTo(HaveOccurred())
			err = field.SetValueRef(&data.attrib12)
			Expect(err).NotTo(HaveOccurred())
			value = field.Value()
			Expect(*value).To(Equal("attrib12 #2"))

			// value update in struct should reflect
			// when retrieved via InputForm
			data.attrib12 = "attrib12 #3"
			value = field.Value()
			Expect(*value).To(Equal("attrib12 #3"))

			// value update in input form
			// should reflect in struct
			newValue = "attrib12 #3"
			err = field.SetValue(&newValue)
			Expect(data.attrib12).To(Equal("attrib12 #3"))

			field, err = ig.GetInputField("attrib13")
			Expect(err).NotTo(HaveOccurred())
			err = field.SetValueRef(&data.attrib13)
			Expect(err).NotTo(HaveOccurred())
			value = field.Value()
			Expect(value).To(BeNil())

			// value update in input form
			// should reflect in struct
			newValue = "attrib13 #1"
			err = field.SetValue(&newValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(*data.attrib13).To(Equal("attrib13 #1"))

			data.attrib13 = nil
			value = field.Value()
			Expect(value).To(BeNil())

			// value update in struct should reflect
			// when retrieved via InputForm
			newValue = "attrib13 #2"
			data.attrib13 = &newValue
			value = field.Value()
			Expect(*value).To(Equal("attrib13 #2"))
		})

		It("binds an external data structure to the form", func() {

			var (
				field *forms.InputField
				value *string
			)
			attrib11Value := "attrib11 #1"

			data := struct {
				Attrib11 *string `form_field:"attrib11"`
				Attrib12 *string `form_field:"attrib12"`

				Group2 struct {
					Attrib121 string `form_field:"attrib121"`
					Attrib122 string `form_field:"attrib122"`
				}
			}{
				Attrib11: &attrib11Value,
				Attrib12: nil,

				Group2: struct {
					Attrib121 string `form_field:"attrib121"`
					Attrib122 string `form_field:"attrib122"`
				}{
					Attrib121: "attrib121 #1",
					Attrib122: "attrib122 #1",
				},
			}

			err = ig.BindFields(&data)
			Expect(err).NotTo(HaveOccurred())

			field, err = ig.GetInputField("attrib11")
			Expect(err).NotTo(HaveOccurred())
			value = field.Value()
			Expect(*value).To(Equal("attrib11 #1"))

			// value update in struct should reflect
			// when retrieved via InputForm
			attrib11Value = "attrib11 #2"
			value = field.Value()
			Expect(*value).To(Equal("attrib11 #2"))

			// value update in input form
			// should reflect in struct
			newValue1 := "attrib11 #3"
			err = field.SetValue(&newValue1)
			Expect(*data.Attrib11).To(Equal("attrib11 #3"))

			field, err = ig.GetInputField("attrib12")
			Expect(err).NotTo(HaveOccurred())
			value = field.Value()
			Expect(value).To(BeNil())

			// value update in input form
			// should reflect in struct
			newValue2 := "attrib12 #1"
			err = field.SetValue(&newValue2)
			Expect(err).NotTo(HaveOccurred())
			Expect(*data.Attrib12).To(Equal("attrib12 #1"))

			data.Attrib12 = nil
			value = field.Value()
			Expect(value).To(BeNil())

			// value update in struct should reflect
			// when retrieved via InputForm
			newValue3 := "attrib12 #2"
			data.Attrib12 = &newValue3
			value = field.Value()
			Expect(*value).To(Equal("attrib12 #2"))

			// Validate fields in nested struct
			field, err = ig.GetInputField("attrib121")
			Expect(err).NotTo(HaveOccurred())
			value = field.Value()
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal("attrib121 #1"))

			newValue4 := "attrib121 #2"
			err = field.SetValue(&newValue4)
			value = field.Value()
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal("attrib121 #2"))

			field, err = ig.GetInputField("attrib122")
			Expect(err).NotTo(HaveOccurred())
			value = field.Value()
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal("attrib122 #1"))
		})
	})
})
