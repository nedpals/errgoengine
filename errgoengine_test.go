package errgoengine_test

import (
	"testing"
	"testing/fstest"

	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/languages/java"
	"github.com/nedpals/errgoengine/languages/python"
)

func Setup(lang *lib.Language, workingPath string, targetPos lib.Position) *lib.ContextData {
	// setup context data
	contextData := lib.NewContextData(lib.NewEmptyStore(), ".")
	contextData.Analyzer = lang.AnalyzerFactory(contextData)
	contextData.TraceStack = lib.TraceStack{}

	// add dummy stack trace item
	contextData.TraceStack.Add("main", lib.Location{
		DocumentPath: workingPath,
		StartPos:     targetPos,
		EndPos:       targetPos,
	})

	return contextData
}

func TestParseFromStackTrace(t *testing.T) {
	t.Run("Simple/Java", func(t *testing.T) {
		currentLang := java.Language
		contextData := Setup(currentLang, "Main.java", lib.Position{Line: 4})
		files := fstest.MapFS{
			"Main.java": &fstest.MapFile{
				Data: []byte(`public class Main {
	public static void main(String[] args) {
		int a = 1;
		System.out.println(a/0);
	}
}`),
			},
		}

		err := lib.ParseFromStackTrace(contextData, currentLang, files)
		if err != nil {
			t.Fatal(err)
		}

		// check if the document is parsed
		doc, ok := contextData.Documents["Main.java"]
		if !ok {
			t.Error("Main.java document not found")
		}

		// check if doc language is same as currentLang
		if doc.Language == nil {
			t.Error("Language is nil")
		}

		if doc.Language != currentLang {
			t.Errorf("expected language %s, got %s", currentLang.Name, doc.Language.Name)
		}

		// check if tree is present
		if doc.Tree == nil {
			t.Error("Tree is nil")
		}

		// check if the content is also present
		if doc.Contents == "" {
			t.Error("Content is empty")
		}

		// check if the tree is parsed
		if doc.Tree.RootNode() == nil {
			t.Error("Tree is not parsed")
		}
	})

	t.Run("Simple/Python", func(t *testing.T) {
		currentLang := python.Language
		contextData := Setup(currentLang, "main.py", lib.Position{Line: 2})
		files := fstest.MapFS{
			"main.py": &fstest.MapFile{
				Data: []byte(`def main():
	a = 1
	print(a/0)
`),
			},
		}

		err := lib.ParseFromStackTrace(contextData, currentLang, files)
		if err != nil {
			t.Fatal(err)
		}

		// check if the document is parsed
		doc, ok := contextData.Documents["main.py"]
		if !ok {
			t.Error("main.py document not found")
		}

		// check if doc language is same as currentLang
		if doc.Language == nil {
			t.Error("Language is nil")
		}

		if doc.Language != currentLang {
			t.Errorf("expected language %s, got %s", currentLang.Name, doc.Language.Name)
		}

		// check if tree is present
		if doc.Tree == nil {
			t.Error("Tree is nil")
		}

		// check if the content is also present
		if doc.Contents == "" {
			t.Error("Content is empty")
		}

		// check if the tree is parsed
		if doc.Tree.RootNode() == nil {
			t.Error("Tree is not parsed")
		}
	})

	t.Run("FileNotFound", func(t *testing.T) {
		currentLang := java.Language
		contextData := Setup(currentLang, "Test.java", lib.Position{Line: 2})
		files := fstest.MapFS{}

		err := lib.ParseFromStackTrace(contextData, currentLang, files)
		if err != nil {
			t.Fatal(err)
		}

		// document should not be present
		if _, ok := contextData.Documents["Test.java"]; ok {
			t.Error("Test.java document is present")
		}
	})

	t.Run("LanguageNotMatched", func(t *testing.T) {
		currentLang := python.Language
		contextData := Setup(currentLang, "main.java", lib.Position{Line: 2})
		files := fstest.MapFS{
			"main.java": &fstest.MapFile{
				Data: []byte(`public class Main {
	public static void main(String[] args) {
		int a = 1;
		System.out.println(a/0);
	}
}`),
			},
		}

		err := lib.ParseFromStackTrace(contextData, currentLang, files)
		if err == nil {
			t.Error("expected error, got nil")
		}

		if err.Error() != "no language found for main.java" {
			t.Errorf("expected error message 'no language found for main.java', got %s", err.Error())
		}

		// check if the document is not parsed
		if _, ok := contextData.Documents["main.java"]; ok {
			t.Error("main.java document is parsed")
		}
	})

	t.Run("ExistingDoc", func(t *testing.T) {
		currentLang := python.Language
		contextData := Setup(currentLang, "main.py", lib.Position{Line: 2})
		files := fstest.MapFS{
			"main.py": &fstest.MapFile{
				Data: []byte(`def main():
	a = 1
	print(a/0)
`),
			},
		}

		// parse the document first
		err := lib.ParseFromStackTrace(contextData, currentLang, files)
		if err != nil {
			t.Fatal(err)
		}

		// check if the document is parsed
		doc, ok := contextData.Documents["main.py"]
		if !ok {
			t.Error("main.py document not found")
		}

		// check if doc language is same as currentLang
		if doc.Language == nil {
			t.Error("Language is nil")
		}

		if doc.Language != currentLang {
			t.Errorf("expected language %s, got %s", currentLang.Name, doc.Language.Name)
		}

		// check if tree is present
		if doc.Tree == nil {
			t.Error("Tree is nil")
		}

		// check if the content is also present
		if doc.Contents == "" {
			t.Error("Content is empty")
		}

		newFiles := fstest.MapFS{
			"main.py": &fstest.MapFile{
				Data: []byte(`def main():
	print("Hello, World!")
`),

				// change the file modification time
				Mode: 0,
			},
		}

		oldContent := doc.Contents

		// parse the document again
		err = lib.ParseFromStackTrace(contextData, currentLang, newFiles)
		if err != nil {
			t.Fatal(err)
		}

		// check if the document is parsed
		doc, ok = contextData.Documents["main.py"]
		if !ok {
			t.Error("main.py document not found")
		}

		// check if doc language is same as currentLang
		if doc.Language == nil {
			t.Error("Language is nil")
		}

		if doc.Language != currentLang {
			t.Errorf("expected language %s, got %s", currentLang.Name, doc.Language.Name)
		}

		// check if the content is also present
		if doc.Contents == "" {
			t.Error("Content is empty")
		}

		// check if the content is changed
		if doc.Contents == oldContent {
			t.Error("Content is not changed")
		}
	})
}
