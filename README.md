> [!IMPORTANT]
> What you are seeing here is an in-progress implementation of our thesis paper. This is not usable as of the moment and it remains unanswered if this is effective while no testing has been conducted yet as of the moment.

# ErrgoEngine
A contextualized programming error analysis translation engine.

It is a Go package that analyzes the programming error message in order to build a contextualized data which contains the error type, the data extracted from the error message text and the files from the extracted stack trace data. The data is then used to generate an enhanced form of programming error message which contains an explanation and bug fix suggestions which is local to the said error and at the same time the program's codebase.

Our aim here is to be able to assist novice programmers in enhancing their debugging skills by providing more information that will help them fully understand the programming errors presented to them.

## Programming languages supported
ErrgoEngine uses the [tree-sitter](https://github.com/tree-sitter/tree-sitter) parser to extract information from the source code from different programming languages in a unified way. This enables us to support a variety of programming languages as long as there is support for it. At the moment, we only support for **Java** and **Python** but we will be able to support more programming languages, and in return more types of programming errors, in the near future with improved interface for langauge support.

## Dependencies
ErrgoEngine only relies on the third-party [go-tree-sitter](https://github.com/smacker/go-tree-sitter) package for parsing but the rest of the code relies on the Go standard library so it should not be a problem at all when importing this package to your application.

## TODO
- [ ] Implementation of error templates
- [ ] Tests

## Paper
- TODO

## Copyright and license
(c) 2023 by the [ErrgoEngine Authors](https://github.com/nedpals/errgoengine/graphs/contributors). Code released under the [MIT License](https://github.com/nedpals/errgoengine/blob/main/LICENSE).
