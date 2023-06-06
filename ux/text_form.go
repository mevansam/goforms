package ux

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/peterh/liner"

	"github.com/mevansam/goforms/forms"
	"github.com/mevansam/goutils/logger"
	"github.com/mevansam/goutils/term"
	"github.com/mevansam/goutils/utils"
)

// Field show options
type FieldShowOption int

const (
	DescOnly FieldShowOption = iota
	DescAndValues
	DescAndDefaults
)

type TextForm struct {
	title,
	heading string

	inputGroup *forms.InputGroup
}

func GetFormInput(
	inputForm forms.InputForm,
	title, heading string,
	indentSpaces, width int,
	tags ...string,
) error {

	var (
		err      error
		textForm *TextForm
	)

	fmt.Println()
	if textForm, err = NewTextForm(
		title, heading,
		inputForm); err != nil {
		// if this happens there is an internal
		// error and it is most likely a bug
		panic(err)
	}
	if err = textForm.GetInput(
		indentSpaces,
		width,
		tags...,
	); err != nil {

		if err == liner.ErrPromptAborted {
			fmt.Println(
				color.Red.Render("\nConfiguration input aborted.\n"),
			)
			os.Exit(1)
		} else {
			return err
		}
	}
	return nil
}

func NewTextForm(
	title, heading string,
	input forms.Input,
) (*TextForm, error) {

	var (
		ok         bool
		inputGroup *forms.InputGroup
	)

	if inputGroup, ok = input.(*forms.InputGroup); !ok {
		return nil, fmt.Errorf("input is not of type forms.InputGroup: %#v", input)
	}

	return &TextForm{
		title:   title,
		heading: heading,

		inputGroup: inputGroup,
	}, nil
}

func (tf *TextForm) GetInput(
	indentSpaces, width int,
	tags ...string,
) error {

	var (
		err    error
		exists bool

		nameLen, l, j int

		cursor     *forms.InputCursor
		inputField *forms.InputField
		input      forms.Input

		doubleDivider, singleDivider,
		prompt, response, suggestion,
		envVal string

		value *string

		valueFromFile bool
		filePaths,
		hintValues,
		fieldHintValues []string
	)

	line := liner.NewLiner()
	line.SetCtrlCAborts(true)

	defer func() {
		line.Close()
	}()

	doubleDivider = strings.Repeat("=", width)
	singleDivider = strings.Repeat("-", width)

	tf.printFormHeader("", width)
	fmt.Println(doubleDivider)
	fmt.Println()

	promptInput := func(input forms.Input) {

		fmt.Println(tf.getInputLongDescription(
			input,
			DescOnly,
			"", "",
			0, width, len(input.DisplayName()),
		))
		fmt.Println(singleDivider)
		prompt = ": "
	}

	cursor = forms.NewInputCursor(tf.inputGroup, tags...)
	cursor = cursor.NextInput()

	for cursor != nil {
		if input, err = cursor.GetCurrentInput(); err != nil {
			return err
		}

		if input.Type() == forms.Container {

			inputs := input.EnabledInputs(true, tags...)
			if len(inputs) > 1 {
				fmt.Println(input.Description())
				fmt.Println(doubleDivider)

				// normalize display name length
				// of all input group fields
				nameLen = 0
				for _, ii := range input.Inputs() {
					l = len(ii.DisplayName())
					if nameLen < l {
						nameLen = l
					}
				}

				// show list of possible inputs and prompt
				// which input should be requested
				options := make([]string, len(inputs))

				for i, ii := range inputs {

					options[i] = strconv.Itoa(i + 1)
					fmt.Println(tf.getInputLongDescription(
						ii,
						DescOnly,
						"", fmt.Sprintf("%s. ", options[i]),
						0, width, l,
					))
					fmt.Println(singleDivider)
				}

				line.SetCompleter(func(line string) (c []string) {
					// allow selection of options using tab
					return options
				})
				for {
					if response, err = line.Prompt("Please select one of the above ? "); err != nil {
						return err
					}
					if j, err = strconv.Atoi(response); err == nil {
						break
					}
				}

				fmt.Println(singleDivider)
				input = inputs[j-1]
				prompt = input.DisplayName() + " : "

			} else if len(inputs) == 1 {
				// If only a single input is available within
				// then skip showing the container input options
				input = inputs[0]
				promptInput(input)

			} else {
				input = nil
			}

		} else if input.Enabled(true, tags...) {
			promptInput(input)
			prompt = ": "

		} else {
			input = nil
		}

		if input != nil {
			inputField = input.(*forms.InputField)
			value = inputField.Value()

			valueFromFile, filePaths = inputField.ValueFromFile()
			if valueFromFile {

				// if value for the field is sourced from a file then
				// create a list of auto-completion hints with default
				// values from the environment
				if value != nil {
					hintValues = append(filePaths, []string{"", "[saved]"}...)
				} else {
					hintValues = append(filePaths, "")
				}
				suggestion = hintValues[len(hintValues)-1]

			} else {

				if values := inputField.AcceptedValues(); values != nil {
					// if values are restrcted to a given list then
					// create a list of auto-completion hints only
					// with those values
					hintValues = values
					if value != nil {
						suggestion = *value
					} else {
						suggestion = ""
					}

				} else {
					// create a list of auto-completion hints from
					// the environment variable associated with the
					// input field along with any values retrieved
					// from any field hints set in the input group.
					hintValues = []string{}

					// set of added values used to ensure
					// the same values are not added twice
					valueSet := map[string]bool{"": true}
					if value != nil && len(*value) > 0 {
						valueSet[*value] = true
					}

					// add values sourced from environment to completion list
					for _, e := range inputField.EnvVars() {
						if envVal, exists = os.LookupEnv(e); exists {
							if _, exists = valueSet[envVal]; !exists {
								hintValues = append(hintValues, envVal)
								valueSet[envVal] = true
							}
						}
					}

					// add values sourced from hints to completion list
					if fieldHintValues, err = tf.inputGroup.GetFieldValueHints(input.Name()); err != nil {
						logger.DebugMessage(
							"Error retrieving hint values for field '%s': '%s'",
							input.Name(), err.Error())
					}
					hintValues = append(append(hintValues, fieldHintValues...), "")
					if value != nil {
						hintValues = append(hintValues, *value)
					}
					suggestion = hintValues[len(hintValues)-1]
				}
			}

			line.SetCompleter(func(line string) []string {
				filteredHintValues := []string{}
				for _, v := range hintValues {
					if strings.HasPrefix(v, strings.ToLower(line)) {
						filteredHintValues = append(filteredHintValues, v)
					}
				}
				return filteredHintValues
			})
			if response, err = line.PromptWithSuggestion(prompt, suggestion, -1); err != nil {
				return err
			}

			// set input with entered value
			if valueFromFile && response == "[saved]" {
				if cursor, err = cursor.SetDefaultInput(input.Name()); err != nil {
					return err
				}
			} else {
				if cursor, err = cursor.SetInput(input.Name(), response); err != nil {
					return err
				}
			}
			if valueFromFile && !inputField.Sensitive() {
				fmt.Printf("Value from file: \n%s\n", *inputField.Value())
			}

			fmt.Println()
		}

		cursor = cursor.NextInput()
	}

	return nil
}

func (tf *TextForm) ShowInputReference(
	fieldShowOption FieldShowOption,
	startIndent, indentSpaces, width int,
	tags ...string,
) {

	var (
		padding    string
		printInput func(level int, input forms.Input)

		evalFieldDeps bool
		fieldLengths  map[string]*int
	)

	padding = strings.Repeat(" ", startIndent)
	evalFieldDeps = fieldShowOption == DescAndValues

	fieldLengths = make(map[string]*int)
	tf.calcNameLengths(tf.inputGroup, fieldLengths, nil, true, tags...)

	printInput = func(level int, input forms.Input) {

		var (
			levelIndent string

			inputs []forms.Input
			ii     forms.Input
			i, l   int
		)

		// skip if input is disabled
		if !input.Enabled(evalFieldDeps, tags...) {
			return
		}

		fmt.Println()
		if input.Type() == forms.Container {

			// output description of a group of
			// inputs which are mutually exclusive

			fmt.Print(padding)
			utils.RepeatString(" ", level*indentSpaces, os.Stdout)
			fmt.Printf("* Provide one of the following for:\n\n")

			levelIndent = strings.Repeat(" ", (level+1)*indentSpaces)

			fmt.Print(padding)
			fmt.Print(levelIndent)
			fmt.Print(input.Description())

		} else {

			fmt.Print(tf.getInputLongDescription(
				input,
				fieldShowOption,
				padding, "* ",
				level*indentSpaces, width, *fieldLengths[input.DisplayName()],
			))
		}

		inputs = input.Inputs()
		for i, ii = range inputs {

			if input.Type() == forms.Container {
				if i > 0 {
					fmt.Print("\n\n")
					fmt.Print(padding)
					fmt.Print(levelIndent)
					fmt.Print("OR\n")
				} else {
					fmt.Print("\n")
				}
				printInput(level+1, ii)

			} else {
				if ii.Type() == forms.Container {
					fmt.Print("\n")
				}
				printInput(level, ii)
			}
		}
		if input.Type() == forms.Container {

			inputs = ii.Inputs()
			l = len(inputs)

			// end group with new line. handle case where if last input
			// also had a container at the end of its inputs two newlines
			// will be output when only on newline should have been output
			if l == 0 || inputs[l-1].Type() != forms.Container {
				fmt.Print("\n")
			}
		}
	}

	tf.printFormHeader(padding, width)
	for _, i := range tf.inputGroup.Inputs() {
		printInput(0, i)
	}
}

func (tf *TextForm) printFormHeader(
	padding string,
	width int,
) {
	var (
		text strings.Builder
	)

	text.WriteString(padding)
	text.WriteString(tf.title)
	text.Write(term.LineFeedB)
	text.WriteString(padding)
	utils.RepeatString("=", len(tf.title), &text)

	fmt.Print(color.OpBold.Render(text.String()))
	fmt.Print("\n\n")

	fmt.Print(padding)
	fmt.Print(tf.inputGroup.Description())
	fmt.Print("\n\n")

	if len(tf.heading) > 0 {
		l := len(padding)
		s, _ := utils.FormatMultilineString(tf.heading, l, width-l, true, true)
		fmt.Print(color.OpItalic.Render(s))
		fmt.Println()
	}
}

func (tf *TextForm) getInputLongDescription(
	input forms.Input,
	fieldShowOption FieldShowOption,
	padding, bullet string,
	indent, width, nameLen int,
) string {

	var (
		ok bool

		out strings.Builder

		name string
		l    int

		field  *forms.InputField
		value  *string
		output string
	)

	out.WriteString(padding)
	utils.RepeatString(" ", indent, &out)
	out.WriteString(bullet)

	name = input.DisplayName()
	out.WriteString(name)

	utils.RepeatString(" ", nameLen-len(name), &out)
	if fieldShowOption == DescAndValues && input.Type() != forms.Container {
		out.WriteString(" = ")
		l = len(out.String())

		if field, ok = input.(*forms.InputField); ok {
			if value = field.Value(); value != nil {
				if field.Sensitive() {
					out.WriteString("****")
				} else {
					output, _ = utils.FormatMultilineString(*value, l, width-l, false, true)
					out.WriteString(output)
				}
			} else {
				out.WriteString("[no data]")
			}
		}

		out.WriteString("\n")
		description, _ := utils.FormatMultilineString(
			color.OpFuzzy.Render(input.LongDescription()),
			l, width-l, true, true)
		out.WriteString(description)
	} else {
		out.WriteString(" - ")

		l = len(out.String())
		description, _ := utils.FormatMultilineString(
			input.LongDescription(), 
			l, width-l, false, true)
		out.WriteString(description)

		if fieldShowOption == DescAndDefaults && input.Type() != forms.Container {
			if field, ok = input.(*forms.InputField); ok {
				if value = field.DefaultValue(); value != nil {
					out.WriteString("\n")

					if field.Sensitive() {
						output, _ = utils.FormatMultilineString(
							"(Default value = '****')", 
							l, width-l, true, true)
						out.WriteString(output)
					} else {
						output, _ = utils.FormatMultilineString(
							fmt.Sprintf("(Default value = '%s')", *value), 
							l, width-l, true, true)
						out.WriteString(output)
					}
				}
			}
		}
	}

	return out.String()
}

func (tf *TextForm) calcNameLengths(
	input forms.Input,
	fieldLengths map[string]*int,
	length *int,
	isRoot bool,
	tags ...string,
) {

	if length == nil {
		ll := 0
		length = &ll
	}

	for _, i := range input.EnabledInputs(false, tags...) {

		if i.Type() != forms.Container {

			if !isRoot && input.Type() == forms.Container {
				// reset length for each input in a container
				ll := 0
				length = &ll
			}

			name := i.DisplayName()
			fieldLengths[name] = length

			ii := i.Inputs()
			if len(ii) > 0 {
				tf.calcNameLengths(i, fieldLengths, length, false)
			}

			l := len(name)
			if l > *length {
				*length = l
			}
		} else {

			// reset length when a container is encountered
			ll := 0
			length = &ll

			tf.calcNameLengths(i, fieldLengths, nil, false)
		}
	}
}
