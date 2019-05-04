package help

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/go-joe/joe"
	"go.uber.org/zap"
)

type helper struct {
	logger  *zap.Logger
	command *regexp.Regexp
	help    []helpInfo
}

type helpInfo struct {
	pattern string
	descr   string
}

func Adapter() joe.Module {
	return joe.ModuleFunc(func(conf *joe.Config) error {
		h := &helper{
			logger:  conf.Logger(""),
			command: regexp.MustCompile(`^help(\s+.+)?$`),
		}

		conf.RegisterHandler(h.registerCommand)
		conf.RegisterHandler(h.helpCommand)

		return nil
	})
}

func (h *helper) registerCommand(evt joe.RegisterCommandEvent) {
	if evt.Expression == "" {
		return
	}

	if evt.Expression[0] == '^' {
		evt.Expression = evt.Expression[1:]
	}
	if evt.Expression[len(evt.Expression)-1] == '$' {
		evt.Expression = evt.Expression[:len(evt.Expression)-1]
	}

	pc := reflect.ValueOf(evt.Function).Pointer()
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		h.logger.Error("Failed to lookup function pointer to add the help text")
		return
	}

	funName := funcName(fn)
	file, _ := fn.FileLine(pc)

	// Try to parse the file to extract the documentation
	// TODO: this breaks if we have a binary only right?
	// -> We could use code generation to work around this issue
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		h.logger.Error("Failed to parse file to lookup help text",
			zap.String("file", file),
			zap.Error(err),
		)
		return
	}

	var descr string
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.File:
			// Function declarations are children of the File AST node so we
			// have to go deeper.
			return true

		case *ast.FuncDecl:
			if x.Name.Name == funName && x.Doc != nil {
				descr = strings.TrimPrefix(x.Doc.List[0].Text, "// ")
			}

			return false

		default:
			// Do not inspect any other AST notes.
			return false
		}
	})

	descr = strings.ReplaceAll(descr, "\n", " ")
	h.help = append(h.help, helpInfo{
		pattern: evt.Expression,
		descr:   strings.TrimSpace(descr),
	})
}

// The helpCommand prints a helpful description for each command the bot
// responds to.
func (h *helper) helpCommand(msg joe.ReceiveMessageEvent) error {
	if !h.command.MatchString(msg.Text) {
		return nil
	}

	var filter string
	matches := h.command.FindStringSubmatch(msg.Text)
	if len(matches) > 1 {
		filter = strings.TrimSpace(matches[1])
	}

	for _, h := range h.help {
		if filter == "" || strings.Contains(h.pattern, filter) {
			msg.Respond("%s: %s", h.pattern, h.descr)
		}
	}
	return nil
}

func funcName(fun *runtime.Func) string {
	splitFuncName := strings.Split(fun.Name(), ".")
	name := splitFuncName[len(splitFuncName)-1]

	nameParts := strings.SplitN(name, "-", 2)
	return nameParts[0]
}
