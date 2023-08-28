package testutils

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	lib "github.com/nedpals/errgoengine"
)

// func init() {
// 	testing.Init()
// 	flag.Parse()
// }

func IsTestRun() bool {
	// return flag.Lookup("test.v").Value.(flag.Getter).Get().(bool)
	return true
}

type TestCase struct {
	Name             string
	Input            string
	ExpectedLanguage string
	ExpectedTemplate string
	ExpectedOutput   string
	Files            fstest.MapFS
}

type TestCases struct {
	engine  *lib.ErrgoEngine
	entries map[string][]TestCase
}

type SetupTestConfig struct {
	DirName        string
	TemplateLoader func(*lib.ErrorTemplates)
}

func SetupTest(tb testing.TB, cfg SetupTestConfig) TestCases {
	wd, err := os.Getwd()
	if err != nil {
		tb.Fatal(err)
	}

	name := filepath.Base(wd)
	if name != cfg.DirName {
		tb.Fatalf("exp dirname %s, got %s", cfg.DirName, name)
	}

	// load error templates
	engine := lib.New()
	cfg.TemplateLoader(&engine.ErrorTemplates)

	// load tests
	testFilesDirPath := filepath.Join(wd, "test_files")
	testFilePaths, err := fs.Glob(os.DirFS(testFilesDirPath), "**/test.txt")
	if err != nil {
		tb.Fatal(err)
	} else if len(testFilePaths) == 0 {
		tb.Fatalf("no test files found (wd=%s)", wd)
	}

	p := NewParser()

	cases := TestCases{
		engine:  engine,
		entries: map[string][]TestCase{},
	}

	for _, testPath := range testFilePaths {
		// parse test.txt
		fullTestPath := filepath.Join(testFilesDirPath, testPath)
		testContents, err := os.ReadFile(fullTestPath)
		if err != nil {
			tb.Error(err)
			continue
		}

		in, exp, err := p.ParseInputExpected(fullTestPath, string(testContents))
		if err != nil {
			tb.Error(err)
			continue
		}

		// get template
		expTemp := engine.ErrorTemplates.Find(in.Language, in.Template)
		if expTemp == nil {
			tb.Errorf(
				"no error template found for `%s` (language=`%s`, name=`%s`)",
				testPath,
				in.Language,
				in.Template)
			continue
		}

		// get test files
		testFiles := fstest.MapFS{}
		testDirPath := filepath.Dir(fullTestPath)
		err = filepath.WalkDir(testDirPath, func(path string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}

			if expTemp.Language.MatchPath(path) {
				fileContent, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				relFilePath, _ := filepath.Rel(testDirPath, path)
				testFiles[relFilePath] = &fstest.MapFile{
					Data: fileContent,
				}
			}

			return nil
		})
		if err != nil {
			// just let it slide...
			tb.Error(err)
		}

		tmpKey := lib.TemplateKey(in.Language, in.Template)
		if _, entryExists := cases.entries[tmpKey]; !entryExists {
			cases.entries[tmpKey] = []TestCase{}
		}

		cases.entries[tmpKey] = append(cases.entries[tmpKey], TestCase{
			Input:            in.Output,
			Name:             in.Name,
			ExpectedLanguage: exp.Language,
			ExpectedTemplate: exp.Template,
			ExpectedOutput:   strings.TrimSpace(exp.Output),
			Files:            testFiles,
		})
	}

	return cases
}

func (cases TestCases) Execute(t *testing.T) {
	for tmpName := range cases.engine.ErrorTemplates {
		t.Run(tmpName, func(t *testing.T) {
			tCases, exists := cases.entries[tmpName]
			if !exists {
				t.Fatal("Test case not implemented")
			}

			for _, tCase := range tCases {
				caseName := tCase.Name
				if len(caseName) == 0 {
					caseName = "Simple"
				}

				t.Run(caseName, func(t *testing.T) {
					cases.engine.FS = tCase.Files
					template, data, err := cases.engine.Analyze("", tCase.Input)
					if err != nil {
						t.Fatal(err)
					} else if template.Name != tCase.ExpectedTemplate || template.Language.Name != tCase.ExpectedLanguage {
						t.Fatalf(
							"\nExpected: (Language: %s, Template: %s)\nGot:      (Language: %s, Template: %s)",
							tCase.ExpectedLanguage, tCase.ExpectedTemplate,
							template.Language.Name, template.Name,
						)
					}

					output := cases.engine.Translate(template, data)
					if output != tCase.ExpectedOutput {
						t.Errorf("\nExpected: %s\nGot:      %s", tCase.ExpectedOutput, output)
					}
				})
			}
		})
	}
}

func Equals[V comparable](tb testing.TB, exp V, got V) {
	if got != exp {
		tb.Fatalf("\nexp: %v\ngot: %v", exp, got)
	}
}

func ExpectError(tb testing.TB, err error, exp string) {
	if err == nil {
		tb.Fatalf("expected error, got nil")
	}

	Equals(tb, err.Error(), exp)
}
