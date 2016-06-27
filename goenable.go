package main // import "github.com/bashgo/goenable"

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bashgo/goenable/bash"
)

var (
	input = flag.String("input", "", "input go file name")
)

const (
	hFile = `
/*
Most of this content is copied from bash shell source code.
http://git.savannah.gnu.org/cgit/bash.git
*/

/* A structure which represents a word. */
typedef struct word_desc {
  char *word;           /* Zero terminated string. */
  int flags;            /* Flags associated with this word. */
} WORD_DESC;

/* A linked list of words. */
typedef struct word_list {
  struct word_list *next;
  WORD_DESC *word;
} WORD_LIST;

typedef int sh_builtin_func_t (WORD_LIST *);

/* The thing that we build the array of builtins out of. */
struct builtin {
  char *name;                   /* The name that the user types. */
  sh_builtin_func_t *function;  /* The address of the invoked function. */
  int flags;                    /* One of the #defines above. */
  char * const *long_doc;       /* NULL terminated array of strings. */
  const char *short_doc;        /* Short version of documentation. */
  char *handle;                 /* for future use */
};

#define EXECUTION_SUCCESS 0
#define EXECUTION_FAILURE 1

#define BUILTIN_ENABLED 0x01

`

	cFile = `
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "{{.Name}}_builtins.h"

/*
Defs that a normally created in a .h file by the go build -buildmode=c-shared command.
Here for convinience
*/

typedef long long GoInt64;
typedef GoInt64 GoInt;
typedef struct { void *data; GoInt len; GoInt cap; } GoSlice;
typedef struct { char *p; GoInt n; } GoString;

{{range .Builtins}}
extern struct builtin {{.Name}}_struct;
extern char *{{.Name}}_doc[];
extern int {{.Name}}_builtin (WORD_LIST *list);
extern int {{.Name}} (GoSlice slice);
{{end}}

/*
Much of this content is copied from bash shell source code
http://git.savannah.gnu.org/cgit/bash.git
*/

int n_word_list(list)
WORD_LIST *list;
{
    int i;
    WORD_LIST *l;
    for (l = list, i = 0; l; l = l->next, i++);
    return i;
}

void word_list_to_string_slice(slice, list)
GoSlice *slice;
WORD_LIST *list;
{
    int i;
    WORD_LIST *l;
    int c = n_word_list(list);
    GoString * sliceData = calloc(c, sizeof(GoString));
    slice->len = c;
    slice->cap = c;
    slice->data = sliceData;
    for ( l = list, i = 0; l; l = l->next, i++ ) {
        sliceData[i].n = strlen(l->word->word);
        sliceData[i].p = l->word->word;
    }
}

{{range .Builtins}}

/* built in fuction for cmd: {{.Name}} */

int {{.Name}}_builtin(list)
WORD_LIST *list;
{
    int ret = EXECUTION_SUCCESS;
    int i;
    int c;
    WORD_LIST *l;
    GoSlice slice;
    c = n_word_list(list);
    word_list_to_string_slice(&slice, list);
    ret = {{.Name}}(slice);
    cfree(slice.data);
    fflush(stdout);
    fflush(stderr);

    return ret;
}

char *{{.Name}}_doc[] = {
    {{.ShortDoc}},
    {{range .LongDoc}}
    {{.}},
    {{end}}
    (char*) NULL
};

struct builtin {{.Name}}_struct = {
    "{{.Name}}",
    {{.Name}}_builtin,
    BUILTIN_ENABLED,
    {{.Name}}_doc,
    {{.ShortDoc}},
    0
};

{{end}}
`
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
}

func parseFile(path string) {
	c := template.Must(template.New("cfile").Parse(cFile))
	h := template.Must(template.New("hFile").Parse(hFile))
	var text interface{}
	fs := token.NewFileSet()
	astFile, err := parser.ParseFile(fs, *input, text, 0)
	if err != nil {
		log.Fatal(err)
	}
	astName := astFile.Name.Name
	headerName, err := filepath.Abs(*input)
	if err != nil {
		log.Fatal(err)
	}
	headerName = filepath.Dir(headerName)
	headerName = filepath.Base(headerName) + ".h"
	builtins := make([]bash.Enable, 0)
	for _, v := range astFile.Scope.Objects {
		if v.Kind != ast.Var {
			continue
		}

		if mTO, ok := v.Decl.(*ast.ValueSpec); ok {
			for _, val := range mTO.Values {
				if complit, ok := val.(*ast.CompositeLit); ok {
					docArr := make([]string, 0)
					var name string
					var shortDoc string
					elts := complit.Elts
					if selector, ok := complit.Type.(*ast.SelectorExpr); ok {
						if selector.Sel.Name == "Enable" {
							name = strings.Replace(elts[0].(*ast.BasicLit).Value, "\"", "", -1)
							for _, doc := range elts[1].(*ast.CompositeLit).Elts {
								docArr = append(docArr, doc.(*ast.BasicLit).Value)
							}
							shortDoc = elts[2].(*ast.BasicLit).Value
							builtins = append(builtins, bash.Enable{
								Name:     name,
								ShortDoc: shortDoc,
								LongDoc:  docArr,
							})
						}
					}
					data := struct {
						Name       string
						HeaderName string
						Builtins   []bash.Enable
					}{
						astName,
						headerName,
						builtins,
					}
					cout, err := os.Create(astName + "_builtins.c")
					if err != nil {
						log.Fatal(err)
					}
					defer cout.Close()
					hout, err := os.Create(astName + "_builtins.h")
					if err != nil {
						log.Fatal(err)
					}
					defer hout.Close()

					c.Execute(cout, data)
					h.Execute(hout, data)
				}
			}
		}
	}
}

// this is a comment for main function
func main() {
	log.SetFlags(0)
	log.SetPrefix("goenable")
	flag.Usage = Usage
	flag.Parse()
	flag.Args()
	if *input == "" {
		flag.Usage()
		os.Exit(-2)
	}
	parseFile(*input)
}
