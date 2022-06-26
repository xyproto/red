package mode

// Mode is a per-filetype mode, like for Markdown
type Mode int

const (
	Blank          = iota // No file mode found
	Git                   // Git commits and interactive rebases
	Markdown              // Markdown
	Doc                   // asciidoctor, rst files, sdoc etc
	Makefile              // Makefiles
	Shell                 // Shell scripts and PKGBUILD files
	Config                // Config like yaml, yml, toml, and ini files
	Assembly              // Assembly
	GoAssembly            // Go-style Assembly
	Go                    // Go
	Haskell               // Haskell
	OCaml                 // OCaml
	StandardML            // Standard ML
	Python                // Python
	Text                  // plain text documents
	CMake                 // CMake files
	Vim                   // Vim or NeoVim configuration, or .vim scripts
	V                     // V programming language
	Clojure               // Clojure
	Lisp                  // Common Lisp and Emacs Lisp
	Zig                   // Zig
	Kotlin                // Kotlin
	Java                  // Java
	Gradle                // Gradle
	HIDL                  // Android-related: Hardware Abstraction Layer Interface Definition Language
	AIDL                  // Android-related: Android Interface Definition Language
	SQL                   // Structured Query Language
	Oak                   // Oak
	Rust                  // Rust
	Lua                   // Lua
	Crystal               // Crystal
	Erlang                // Erlang
	Nim                   // Nim
	ObjectPascal          // Object Pascal and Delphi
	Bat                   // DOS and Windows batch files
	Cpp                   // C++
	C                     // C
	Ada                   // Ada
	HTML                  // HTML
	Odin                  // Odin
	Hare                  // Hare
	Jakt                  // Jakt
	XML                   // XML
	PolicyLanguage        // SE Linux configuration files
	Nroff                 // editing man pages
	Scala                 // Scala
	JSON                  // JSON and iPython notebooks
	Battlestar            // Battlestar
	CS                    // C#
	JavaScript            // JavaScript
	TypeScript            // TypeScript
	ManPage               // viewing man pages
	Amber                 // Amber templates
	Bazel                 // Bazel and Starlark
	D                     // D
	Perl                  // Perl
	Prolog                // Prolog
	M4                    // M4 macros
	Agda                  // Agda
	Basic                 // FreeBasic, Gambas 3
	Log                   // All sorts of log files
	Teal                  // Teal
	Email                 // For using o with ie. Mutt
)

// String will return a short lowercase string representing the given editor mode
func (mode Mode) String() string {
	// TODO: Sort the cases alphabetically
	// TODO: Add a test that makes sure every mode has a string
	switch mode {
	case Blank:
		return "-"
	case Teal:
		return "Teal"
	case Git:
		return "Git"
	case Markdown:
		return "Markdown"
	case Makefile:
		return "Make"
	case Shell:
		return "Shell"
	case Config:
		return "Configuration"
	case Assembly:
		return "Assembly"
	case GoAssembly:
		return "Go-style Assembly"
	case Go:
		return "Go"
	case Haskell:
		return "Haskell"
	case OCaml:
		return "Ocaml"
	case StandardML:
		return "Standard ML"
	case Python:
		return "Python"
	case Text:
		return "Text"
	case CMake:
		return "Cmake"
	case Vim:
		return "ViM"
	case Clojure:
		return "Clojure"
	case Lisp:
		return "Lisp"
	case Zig:
		return "Zig"
	case Kotlin:
		return "Kotlin"
	case Java:
		return "Java"
	case Gradle:
		return "Gradle"
	case HIDL:
		return "HIDL"
	case AIDL:
		return "AIDL"
	case SQL:
		return "SQL"
	case Oak:
		return "Oak"
	case Rust:
		return "Rust"
	case Lua:
		return "Lua"
	case Crystal:
		return "Crystal"
	case Erlang:
		return "Erlang"
	case Nim:
		return "Nim"
	case ObjectPascal:
		return "Pas"
	case Bat:
		return "Bat"
	case Cpp:
		return "C++"
	case C:
		return "C"
	case Ada:
		return "Ada"
	case HTML:
		return "HTML"
	case Odin:
		return "Odin"
	case Hare:
		return "Hare"
	case Jakt:
		return "Jakt"
	case Perl:
		return "Perl"
	case Prolog:
		return "Prolog"
	case XML:
		return "XML"
	case PolicyLanguage:
		return "SELinux"
	case Nroff:
		return "Nroff"
	case Scala:
		return "Scala"
	case JSON:
		return "JSON"
	case Battlestar:
		return "Battlestar"
	case CS:
		return "C#"
	case TypeScript:
		return "TypeScript"
	case JavaScript:
		return "JavaScript"
	case ManPage:
		return "Man"
	case Amber:
		return "Amber"
	case Bazel:
		return "Bazel"
	case D:
		return "D"
	case V:
		return "V"
	case M4:
		return "M4"
	case Agda:
		return "Agda"
	case Basic:
		return "Basic"
	case Log:
		return "Log"
	case Email:
		return "E-mail"
	case Doc:
		return "Document"
	default:
		return "?"
	}
}
