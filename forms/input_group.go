package forms

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/mevansam/goutils/utils"
)

// Input types
type InputType int

const (
	String InputType = iota
	Number
	FilePath
	HttpUrl
	EmailAddress
	JsonInput
	Container
)

// Input abstraction
type Input interface {
	Name() string
	DisplayName() string
	Description() string
	LongDescription() string

	Type() InputType
	Inputs() []Input
	Enabled() bool

	getGroupId() int
}

// InputForm abstraction
type InputForm interface {
	Input

	BindFields(target interface{}) error

	AddFieldValueHint(name, hint string) error
	GetFieldValueHints(name string) ([]string, error)

	GetInputField(name string) (*InputField, error)

	GetFieldValue(name string) (*string, error)
	SetFieldValue(name string, value string) error

	InputFields() []*InputField
	InputValues() map[string]string
}

// InputField initialization attributes
type FieldAttributes struct {

	// name of the field
	Name,
	// the name to display when requesting input
	DisplayName,
	// a long description which can also be the
	// help text for the field
	Description string

	// defines a group id. all fields having
	// the same group id will be added to
	// a "Container" input group where only
	// one input in the "Container" will
	// be collected. an id of 0 flags the
	// field as not belonging to Container
	// group
	GroupID int

	// type of the input used for validation
	InputType InputType

	// if true then the input value should be
	// a file which will be read as the value
	// of the field
	ValueFromFile bool

	// a default value. nil if no default value
	DefaultValue *string

	// indicates if the field value should be masked
	Sensitive bool

	// any environment variables the value for
	// this input can be sourced from
	EnvVars []string

	// inputs that this input depends. this
	// helps define the input flow. this can
	// be of the format:
	//
	// - "field_name":
	//   this field will require input only
	//   if the given dependent field is
	//   entered
	//
	// - "field_name=value":
	//   this field will be require input
	//   only if the given dependent field has
	//   the given value
	//
	DependsOn []string

	// field value should match this regex
	InclusionFilter,
	// error message to return if inclusion
	// filter does not match
	InclusionFilterErrorMessage string

	// field value should not match this regex
	ExclusionFilter,
	// error message to return if exclusion
	// filter matches
	ExclusionFilterErrorMessage string

	// list of acceptable values for field
	AcceptedValues []string
	// error message to return none of the
	// accepted values match the field value
	AcceptedValuesErrorMessage string

	// tags which used to create field subsets
	// for input
	Tags []string
}

// This structure is a container for a collection of
// inputs which implements the InputForm abstraction
type InputGroup struct {
	name        string
	description string

	displayName string

	groupId int
	inputs  []Input

	containers   map[int]*InputGroup
	fieldNameSet map[string]Input

	fieldValueLookupHints map[string][]string
}

// Regex used to validate hints
var hintRegex = regexp.MustCompile(`^(https?:\/\/|file:\/\/\/?|field:\/\/)([a-z0-9]+([\-\.]{1}[a-z0-9]+)*(\.[a-z]{2,5})*(:[0-9]{1,5})?)(\/.*)?$`)

// in: name        - name of the container
// in: displayName - the name to display when requesting input
// in: description - a long description which can also be
//                     the help text for the container
// out: An initialized instance of an InputGroup of type "Container" structure
func (g *InputGroup) NewInputContainer(
	name, displayName, description string,
	groupId int,
) Input {

	container := &InputGroup{
		name:        name,
		description: description,
		groupId:     groupId,

		displayName: displayName,

		containers:   g.containers,
		fieldNameSet: g.fieldNameSet,

		fieldValueLookupHints: g.fieldValueLookupHints,
	}
	g.containers[groupId] = container

	return container
}

// in: attributes - The field attributes
//
// out: An initialized instance of an InputField structure
func (g *InputGroup) NewInputField(
	attributes FieldAttributes,
) (Input, error) {

	var (
		field *InputField
		err   error
	)
	if field, err = g.newInputField(
		attributes.Name,
		attributes.DisplayName,
		attributes.Description,
		attributes.GroupID,
		attributes.InputType,
		attributes.ValueFromFile,
		attributes.DefaultValue,
		attributes.Sensitive,
		attributes.EnvVars,
		attributes.DependsOn,
	); err != nil {
		return nil, err
	}
	if len(attributes.InclusionFilter) > 0 {
		if err = field.SetInclusionFilter(
			attributes.InclusionFilter,
			attributes.InclusionFilterErrorMessage,
		); err != nil {
			return nil, err
		}
	}
	if len(attributes.ExclusionFilter) > 0 {
		if err = field.SetExclusionFilter(
			attributes.ExclusionFilter,
			attributes.ExclusionFilterErrorMessage,
		); err != nil {
			return nil, err
		}
	}
	if attributes.AcceptedValues != nil {
		field.SetAcceptedValues(
			attributes.AcceptedValues,
			attributes.AcceptedValuesErrorMessage,
		)
	}

	return field, nil
}

// in: name          - name of the field
// in: displayName   - the name to display when requesting
//                     input
// in: description   - a long description which can also be
//                     the help text for the field
// in: groupId       - defines a group id. all fields having
//                     the same group id will be added to
//                     a "Container" input group where only
//                     one input in the "Container" will
//                     be collected. an id of 0 flags the
//                     field as not belonging to Container
//                     group
// in: inputType     - type of the input used for validation
// in: valueFromFile - if true then the input value should be
//                     a file which will be read as the value
//                     of the field
// in: defaultValue  - a default value. nil if no default value
// inL sensitive     - indicates if the field value should be masked
// in: envVars       - any environment variables the value for
//                     this input can be sourced from
// in: dependsOn     - inputs that this input depends. this
//                     helps define the input flow
//
// out: An initialized instance of an InputField structure
func (g *InputGroup) newInputField(
	name, displayName, description string,
	groupId int,
	inputType InputType,
	valueFromFile bool,
	defaultValue *string,
	sensitive bool,
	envVars []string,
	dependsOn []string,
) (*InputField, error) {

	var (
		err    error
		exists bool

		field *InputField
	)

	// Do not allow adding duplicate fields
	if _, exists = g.fieldNameSet[name]; exists {
		return nil, fmt.Errorf(
			"a field with name '%s' has already been added",
			name)
	}

	field = &InputField{
		InputGroup: InputGroup{
			name:        name,
			description: description,
			groupId:     groupId,

			displayName: displayName,

			containers:   g.containers,
			fieldNameSet: g.fieldNameSet,

			fieldValueLookupHints: g.fieldValueLookupHints,
		},
		inputType: inputType,

		valueFromFile: valueFromFile,
		envVars:       envVars,
		defaultValue:  defaultValue,
		sensitive:     sensitive,

		inputSet: false,
		valueRef: nil,

		postFieldConditions: []postCondition{},

		acceptedValues:  nil,
		inclusionFilter: nil,
		exclusionFilter: nil,
	}
	g.fieldNameSet[name] = field

	if dependsOn != nil && len(dependsOn) > 0 {

		// recursively add field to all
		// inputs that it depends on

		var (
			addToDepends func(
				input Input,
				names map[string]string,
			) (bool, error)

			names map[string]string
			added bool
		)

		addToDepends = func(
			input Input,
			names map[string]string,
		) (bool, error) {
			for _, i := range input.Inputs() {

				if value, exists := names[i.Name()]; exists && i.Type() != Container {
					f := i.(*InputField)
					if err = f.addInputField(field); err != nil {
						return false, err
					}
					if len(value) > 0 {
						// add post condition for field which will
						// be skipped unless dependent field equals
						// the given value
						field.postFieldConditions = append(
							field.postFieldConditions,
							postCondition{
								field:  f,
								values: strings.Split(value, "|"),
							},
						)
					}

					delete(names, i.Name())
					if len(names) == 0 {
						return true, nil
					}

				} else if len(i.Inputs()) > 0 {
					if added, err = addToDepends(i, names); added || err != nil {
						return added, err
					}
				}
			}
			return false, nil
		}

		names = make(map[string]string)
		for _, n := range dependsOn {
			tuple := strings.Split(n, "=")
			if len(tuple) == 1 {
				names[tuple[0]] = ""
			} else if len(tuple) > 2 {
				return nil,
					fmt.Errorf(
						"field '%s' has a depends that does not comfirm to format 'name[=value]': %v",
						field.name, tuple)
			} else {
				names[tuple[0]] = tuple[1]
			}
		}
		if added, err = addToDepends(g, names); !added && err == nil {
			err = fmt.Errorf(
				"unable to add field '%s' as one or more dependent fields %v not found",
				field.name, dependsOn)
		}

	} else {
		err = g.addInputField(field)
	}
	return field, err
}

func (g *InputGroup) addInputField(
	field *InputField,
) error {

	var (
		ig     *InputGroup
		exists bool
	)

	add := true
	for _, f := range g.inputs {

		if field.groupId > 0 && field.groupId == f.getGroupId() {

			if f.Type() != Container {
				return fmt.Errorf("invalid internal input container state")
			}

			// append to existing group
			ig = f.(*InputGroup)
			ig.inputs = append(ig.inputs, field)

			add = false
			break
		}
	}
	if add {

		if field.groupId > 0 {
			// retrieve group container
			// to add new field to
			if ig, exists = g.containers[field.groupId]; !exists {
				return fmt.Errorf(
					"unable to add field '%s' as its group '%d' was not found",
					field.name, field.groupId)
			}
			ig.inputs = append(ig.inputs, field)
			g.inputs = append(g.inputs, ig)
		} else {
			g.inputs = append(g.inputs, field)
		}
	}
	return nil
}

// interface: Input

// out: the name of the group
func (g *InputGroup) Name() string {
	return g.name
}

// out: the display name of the group
func (g *InputGroup) DisplayName() string {
	return g.displayName
}

// out: the description of the group
func (g *InputGroup) Description() string {
	return g.description
}

// out: the long description of the group
func (g *InputGroup) LongDescription() string {
	return g.description
}

// out: returns input type of "Container"
func (g *InputGroup) Type() InputType {
	return Container
}

// out: a list of all inputs for the group
func (g *InputGroup) Inputs() []Input {
	return g.inputs
}

// out: whether this group is enabled
func (f *InputGroup) Enabled() bool {
	return true
}

// out: return the group id
func (g *InputGroup) getGroupId() int {
	return g.groupId
}

// interface: InputForm

// in: binds the given target data structure to this input form's fields
func (g *InputGroup) BindFields(target interface{}) error {

	var (
		err  error
		ok   bool
		name string

		field *InputField
	)

	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to a struct")
	}

	t := reflect.Indirect(v).Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if name, ok = f.Tag.Lookup("form_field"); ok {

			if field, err = g.GetInputField(name); err != nil {
				return err
			}
			if err = field.SetValueRef(v.Elem().Field(i).Addr().Interface()); err != nil {
				return err
			}
		} else if f.Type.Kind() == reflect.Struct {

			if err = g.BindFields(v.Elem().Field(i).Addr().Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

// in: name - name of the field for which a hint should be added
// in: hint - the hint which is a URL with the following patterns.
//            * http://<url> - a http url from which source a list values separated by newlines
//            * file://<path> - a path to a file from which to source a list values separated by newlines
//            * field://<name>/<path> - a path to a value in a field with json content
func (g *InputGroup) AddFieldValueHint(name, hint string) error {

	var (
		exists bool
		hints  []string
	)

	if !hintRegex.Match([]byte(hint)) {
		return fmt.Errorf("hint must be a url with prefix http(s)://, file:// or field://")
	}
	if _, exists = g.fieldNameSet[name]; exists {
		if hints, exists = g.fieldValueLookupHints[name]; !exists {
			hints = []string{}
			g.fieldValueLookupHints[name] = hints
		}
		g.fieldValueLookupHints[name] = append(hints, hint)
		return nil
	}
	return fmt.Errorf("field with name '%s' not found", name)
}

// in: name - name of the field for which a hint should be retrieved
// out: array of hint values
func (g *InputGroup) GetFieldValueHints(name string) ([]string, error) {

	var (
		err   error
		value *string

		fieldData interface{}
		hintData  interface{}
	)

	hintValues := []string{}
	if hints := g.fieldValueLookupHints[name]; hints != nil && len(hints) > 0 {

		for _, hint := range hints {
			matchIndices := hintRegex.FindAllStringSubmatchIndex(hint, -1)
			protocol := hint[matchIndices[0][2]:matchIndices[0][3]]

			switch protocol {
			case "http//", "https//":
				// not implemented
				err = fmt.Errorf("not implemented")
			case "file://", "file:///":
				// not implemented
				err = fmt.Errorf("not implemented")
			case "field://":

				fieldName := hint[matchIndices[0][4]:matchIndices[0][5]]
				fieldPath := hint[matchIndices[0][12]:matchIndices[0][13]]

				if value, err = g.GetFieldValue(fieldName); err == nil && value != nil {

					if err = json.Unmarshal([]byte(*value), &fieldData); err == nil {
						if hintData, err = utils.GetValueAtPath(fieldPath, fieldData); err == nil {

							switch hintData.(type) {
							case string:
								hintValues = append(hintValues, hintData.(string))
							case []interface{}:
								for _, v := range hintData.([]interface{}) {
									hintValues = append(hintValues, fmt.Sprintf("%v", v))
								}
							default:
								hintValues = append(hintValues, fmt.Sprintf("%v", hintData))
							}
						}
					} else {
						err = fmt.Errorf(
							"error parsing json value of field '%s': %s",
							fieldName, err.Error())
					}
				}
			}
		}
	}
	return hintValues, err
}

// in: the name of the input field to retrieve
// out: the input field with the given name
func (g *InputGroup) GetInputField(name string) (*InputField, error) {

	var (
		input Input
		field *InputField
		ok    bool
	)

	if input, ok = g.fieldNameSet[name]; !ok {
		return nil, fmt.Errorf("field '%s' was not found in form", name)
	}
	if field, ok = input.(*InputField); !ok {
		return nil, fmt.Errorf("internal state error retrieving field '%s'", name)
	}
	return field, nil
}

// in: the name of the input field whose value should be retrieved
// out: a reference to the value of the input field
func (g *InputGroup) GetFieldValue(name string) (*string, error) {

	var (
		err   error
		field *InputField
	)

	if field, err = g.GetInputField(name); err != nil {
		return nil, err
	}
	return field.Value(), nil
}

// in: the name of the input field to set the value of
// in: a reference to the value to set. if nil the value is cleared
func (g *InputGroup) SetFieldValue(name string, value string) error {

	var (
		err   error
		field *InputField
	)

	if field, err = g.GetInputField(name); err != nil {
		return err
	}
	return field.SetValue(&value)
}

// out: a list of all fields for the group
func (g *InputGroup) InputFields() []*InputField {
	return g.inputFields(make(map[string]bool))
}

// in: set of added fields
// out: a list of all fields for the group
func (g *InputGroup) inputFields(added map[string]bool) []*InputField {

	fields := []*InputField{}
	for _, f := range g.inputs {
		if f.Type() == Container {
			// recursively retrieve fields from groups
			ig := f.(*InputGroup)
			fields = append(fields, ig.inputFields(added)...)

		} else if _, exists := added[f.Name()]; !exists {
			fields = append(fields, f.(*InputField))
			added[f.Name()] = true

			ig := f.(*InputField)
			// recursively retrieve fields from groups
			fields = append(fields, ig.inputFields(added)...)
		}
	}
	return fields
}

// out: map of name-values of all inputs entered
func (g *InputGroup) InputValues() map[string]string {

	var (
		val *string
	)

	valueMap := make(map[string]string)
	inputFields := g.InputFields()

	for _, f := range inputFields {
		if f.InputSet() {
			val = f.Value()
			valueMap[f.Name()] = *val
		}
	}
	return valueMap
}
