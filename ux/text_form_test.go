package ux_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/mevansam/goforms/forms"
	"github.com/mevansam/goforms/ux"
	"github.com/mevansam/goutils/logger"
	"github.com/mevansam/goutils/term"
	"github.com/mevansam/goutils/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	test_data "github.com/mevansam/goforms/test/data"
)

var _ = Describe("Text Formatting tests", func() {

	var (
		err error

		origStdin, stdInWriter,
		origStdout, stdOutReader,
		origStderr *os.File

		inputGroup  *forms.InputGroup
		inputValues map[string]*string
	)

	BeforeEach(func() {

		// pipe output to be written to by form output
		origStdout = os.Stdout
		stdOutReader, os.Stdout, err = os.Pipe()
		Expect(err).ToNot(HaveOccurred())

		// redirect all output to stderr to new stdout
		origStderr = os.Stderr
		os.Stderr = os.Stdout

		// pipe input to be read in by form input
		origStdin = os.Stdin
		os.Stdin, stdInWriter, err = os.Pipe()
		Expect(err).ToNot(HaveOccurred())

		// Bind fields to map of values so
		// that form values can be saved
		inputGroup = test_data.NewTestInputCollection().Group("input-form")

		inputValues = make(map[string]*string)
		for _, f := range inputGroup.InputFields() {
			s := new(string)
			inputValues[f.Name()] = s
			err = f.SetValueRef(s)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		stdOutReader.Close()
		os.Stdout = origStdout
		os.Stderr = origStderr
		stdInWriter.Close()
		os.Stdin = origStdin
	})

	Context("Output", func() {

		var testFormOutput = func(fieldShowOption ux.FieldShowOption, expected string) {

			// channel to signal when getting form input is done
			out := make(chan string)

			go func() {

				var (
					output bytes.Buffer
				)

				tf, err := ux.NewTextForm(
					"Input Data Form for 'input-form'",
					"CONFIGURATION DATA INPUT",
					inputGroup,
				)
				Expect(err).NotTo(HaveOccurred())
				tf.ShowInputReference(fieldShowOption, 2, 2, 80)

				// close piped output
				os.Stdout.Close()
				io.Copy(&output, stdOutReader)

				// signal end
				out <- output.String()
			}()

			// wait until signal is received

			output := <-out
			logger.DebugMessage("\n%s\n", output)
			Expect(output).To(Equal(expected))
		}

		It("outputs a detailed input data form reference", func() {
			testFormOutput(ux.DescAndDefaults, testFormReferenceOutput)
		})

		It("outputs a detailed input data form with field values", func() {

			inputGroup.SetFieldValue("attrib12", "value for attrib12")
			inputGroup.SetFieldValue("attrib122", "value for attrib122")
			inputGroup.SetFieldValue("attrib1221", "value for attrib1221")
			inputGroup.SetFieldValue("attrib131", "value for attrib131")
			inputGroup.SetFieldValue("attrib1311", "value for attrib1311")
			inputGroup.SetFieldValue("attrib1312", "value for attrib1311")
			inputGroup.SetFieldValue("attrib14", "value for attrib14")

			testFormOutput(ux.DescAndValues, testFormOutputWithValues)
		})
	})

	Context("Input", func() {

		var testFormInput = func(testFormInputPrompts string, expectedValues map[string]string) {

			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				defer GinkgoRecover()

				tf, err := ux.NewTextForm(
					"Input Data Form for 'input-form'",
					"CONFIGURATION DATA INPUT",
					inputGroup,
				)
				if err == nil {
					err = tf.GetInput(2, 80)
				}
				Expect(err).NotTo(HaveOccurred())
			}()

			outputReader := bufio.NewScanner(stdOutReader)
			expectReader := bufio.NewScanner(bytes.NewBufferString(testFormInputPrompts))

			actual := ""
			read := true
			readOutput := func(expected string) {
				if !outputReader.Scan() {
					Fail(fmt.Sprintf("TextFrom GetInput() did not output expected string '%s'.", expected))
				}
				actual = outputReader.Text()
				logger.TraceMessage("expect> %s\n", actual)
			}

			for expectReader.Scan() {
				expected := expectReader.Text()
				if i := strings.Index(expected, "<<"); i != -1 {

					prompt := expected[:i]
					input := expected[i+2:]
					stdInWriter.WriteString(input + "\n")
					if read {
						readOutput(expected)
					}

					if strings.HasPrefix(actual, prompt) {
						actual = actual[len(prompt):]
						read = false

					} else {
						Fail(
							fmt.Sprintf(
								"actual line read does not contain prompt as prefix: '%s' !> '%s",
								actual, prompt,
							),
						)
					}

				} else {
					if read {
						readOutput(expected)
					}

					Expect(actual).To(Equal(expected))
					actual = ""
					read = true
				}
			}

			values := inputGroup.InputValues()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(values)).To(Equal(len(expectedValues)))
			Expect(reflect.DeepEqual(expectedValues, values)).To(BeTrue())

			Expect(utils.WaitTimeout(&wg, time.Second)).To(BeTrue())
		}

		It("gathers input for the form from stdin #1", func() {

			expectedValues := map[string]string{
				"attrib12":   "value for attrib12",
				"attrib122":  "value for attrib122",
				"attrib1221": "value for attrib1221",
				"attrib131":  "value for attrib131",
				"attrib1311": "value for attrib1311",
				"attrib1312": "value for attrib1312",
				"attrib14":   "value for attrib14",
			}

			testFormInput(testFormInputPrompts1, expectedValues)
		})

		It("gathers input for the form from stdin #2 - with conditional field", func() {

			expectedValues := map[string]string{
				"attrib12":   "value for attrib12 - A",
				"attrib121":  "value for attrib121",
				"attrib131":  "value for attrib131",
				"attrib1311": "value for attrib1311",
				"attrib1312": "value for attrib1312",
				"attrib14":   "value for attrib14 - X",
				"attrib141":  "value for attrib141",
			}

			testFormInput(testFormInputPrompts2, expectedValues)
		})
	})
})

const testFormReferenceOutput = term.BOLD + `  Input Data Form for 'input-form'
  ================================` + term.NC + `

  test group description

` + term.ITALIC + `  CONFIGURATION DATA INPUT` + term.NC + `

  * Provide one of the following for:

    description for group 1

    * Attrib 11 - description for attrib11. It will be sourced from the
                  environment variables ATTRIB11_ENV1, ATTRIB11_ENV2,
                  ATTRIB11_ENV3 if not provided.

    OR

    * Attrib 12 - description for attrib12. It will be sourced from the
                  environment variable ATTRIB12_ENV1 if not provided.

    * Provide one of the following for:

      description for group 2

      * Attrib 121 - description for attrib121.

      OR

      * Attrib 122  - description for attrib122.
      * Attrib 1221 - description for attrib1221.

    * Attrib 131  - description for attrib131.
    * Attrib 1311 - description for attrib1311.
    * Attrib 1312 - description for attrib1312.

    OR

    * Attrib 13   - description for attrib13. It will be sourced from the
                    environment variables ATTRIB13_ENV1, ATTRIB13_ENV2 if not
                    provided.
    * Attrib 131  - description for attrib131.
    * Attrib 1311 - description for attrib1311.
    * Attrib 1312 - description for attrib1312.

    * Provide one of the following for:

      description for group 3

      * Attrib 132 - description for attrib132. It will be sourced from the
                     environment variable ATTRIB132 if not provided.

      OR

      * Attrib 133 - description for attrib133.
                     (Default value = 'default value for attrib133')

  * Attrib 14 - description for attrib14.
                (Default value = 'default value for attrib14')`

const testFormOutputWithValues = term.BOLD + `  Input Data Form for 'input-form'
  ================================` + term.NC + `

  test group description

` + term.ITALIC + `  CONFIGURATION DATA INPUT` + term.NC + `

  * Provide one of the following for:

    description for group 1

    * Attrib 11 = [no data]
                  ` + term.DIM + `description for attrib11. It will be sourced from the
                  environment variables ATTRIB11_ENV1, ATTRIB11_ENV2,
                  ATTRIB11_ENV3 if not provided.` + term.NC + `

    OR

    * Attrib 12 = value for attrib12
                  ` + term.DIM + `description for attrib12. It will be sourced from the
                  environment variable ATTRIB12_ENV1 if not provided.` + term.NC + `

    * Provide one of the following for:

      description for group 2

      * Attrib 121 = [no data]
                     ` + term.DIM + `description for attrib121.` + term.NC + `

      OR

      * Attrib 122  = value for attrib122
                      ` + term.DIM + `description for attrib122.` + term.NC + `
      * Attrib 1221 = value for attrib1221
                      ` + term.DIM + `description for attrib1221.` + term.NC + `

    * Attrib 131  = value for attrib131
                    ` + term.DIM + `description for attrib131.` + term.NC + `
    * Attrib 1311 = value for attrib1311
                    ` + term.DIM + `description for attrib1311.` + term.NC + `
    * Attrib 1312 = value for attrib1311
                    ` + term.DIM + `description for attrib1312.` + term.NC + `

    OR

    * Attrib 13   = [no data]
                    ` + term.DIM + `description for attrib13. It will be sourced from the
                    environment variables ATTRIB13_ENV1, ATTRIB13_ENV2 if not
                    provided.` + term.NC + `
    * Attrib 131  = value for attrib131
                    ` + term.DIM + `description for attrib131.` + term.NC + `
    * Attrib 1311 = value for attrib1311
                    ` + term.DIM + `description for attrib1311.` + term.NC + `
    * Attrib 1312 = value for attrib1311
                    ` + term.DIM + `description for attrib1312.` + term.NC + `

    * Provide one of the following for:

      description for group 3

      * Attrib 132 = [no data]
                     ` + term.DIM + `description for attrib132. It will be sourced from the
                     environment variable ATTRIB132 if not provided.` + term.NC + `

      OR

      * Attrib 133 = default value for attrib133
                     ` + term.DIM + `description for attrib133.` + term.NC + `

  * Attrib 14 = value for attrib14
                ` + term.DIM + `description for attrib14.` + term.NC

const testFormInputPrompts1 = term.BOLD + `Input Data Form for 'input-form'
================================` + term.NC + `

test group description

` + term.ITALIC + `CONFIGURATION DATA INPUT` + term.NC + `
================================================================================

description for group 1
================================================================================
1. Attrib 11 - description for attrib11. It will be sourced from the environment
               variables ATTRIB11_ENV1, ATTRIB11_ENV2, ATTRIB11_ENV3 if not
               provided.
--------------------------------------------------------------------------------
2. Attrib 12 - description for attrib12. It will be sourced from the environment
               variable ATTRIB12_ENV1 if not provided.
--------------------------------------------------------------------------------
3. Attrib 13 - description for attrib13. It will be sourced from the environment
               variables ATTRIB13_ENV1, ATTRIB13_ENV2 if not provided.
--------------------------------------------------------------------------------
Please select one of the above ? <<2
--------------------------------------------------------------------------------
Attrib 12 : <<value for attrib12

description for group 2
================================================================================
1. Attrib 121 - description for attrib121.
--------------------------------------------------------------------------------
2. Attrib 122 - description for attrib122.
--------------------------------------------------------------------------------
Please select one of the above ? <<2
--------------------------------------------------------------------------------
Attrib 122 : <<value for attrib122

Attrib 1221 - description for attrib1221.
--------------------------------------------------------------------------------
: <<value for attrib1221

Attrib 131 - description for attrib131.
--------------------------------------------------------------------------------
: <<value for attrib131

Attrib 1311 - description for attrib1311.
--------------------------------------------------------------------------------
: <<value for attrib1311

Attrib 1312 - description for attrib1312.
--------------------------------------------------------------------------------
: <<value for attrib1312

Attrib 14 - description for attrib14.
--------------------------------------------------------------------------------
: <<value for attrib14
`

const testFormInputPrompts2 = term.BOLD + `Input Data Form for 'input-form'
================================` + term.NC + `

test group description

` + term.ITALIC + `CONFIGURATION DATA INPUT` + term.NC + `
================================================================================

description for group 1
================================================================================
1. Attrib 11 - description for attrib11. It will be sourced from the environment
               variables ATTRIB11_ENV1, ATTRIB11_ENV2, ATTRIB11_ENV3 if not
               provided.
--------------------------------------------------------------------------------
2. Attrib 12 - description for attrib12. It will be sourced from the environment
               variable ATTRIB12_ENV1 if not provided.
--------------------------------------------------------------------------------
3. Attrib 13 - description for attrib13. It will be sourced from the environment
               variables ATTRIB13_ENV1, ATTRIB13_ENV2 if not provided.
--------------------------------------------------------------------------------
Please select one of the above ? <<2
--------------------------------------------------------------------------------
Attrib 12 : <<value for attrib12 - A

Attrib 121 - description for attrib121.
--------------------------------------------------------------------------------
: <<value for attrib121

Attrib 131 - description for attrib131.
--------------------------------------------------------------------------------
: <<value for attrib131

Attrib 1311 - description for attrib1311.
--------------------------------------------------------------------------------
: <<value for attrib1311

Attrib 1312 - description for attrib1312.
--------------------------------------------------------------------------------
: <<value for attrib1312

Attrib 14 - description for attrib14.
--------------------------------------------------------------------------------
: <<value for attrib14 - X

Attrib 141 - description for attrib141.
--------------------------------------------------------------------------------
: <<value for attrib141
`
