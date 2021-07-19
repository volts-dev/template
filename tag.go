package template

/* Incomplete:
   -----------

   verbatim (only the "name" argument is missing for verbatim)

   Reconsideration:
   ----------------

   debug (reason: not sure what to output yet)
   regroup / Grouping on other properties (reason: maybe too python-specific; not sure how useful this would be in Go)

   Following built-in tags wont be added:
   --------------------------------------

   csrf_token (reason: web-framework specific)
   load (reason: python-specific)
   url (reason: web-framework specific)
*/

import (
	"fmt"
)

// This is the function signature of the tag's parser you will have
// to implement in order to create a new tag.
//
// 'doc' is providing access to the whole document while 'arguments'
// is providing access to the user's arguments to the tag:
//
//     {% your_tag_name some "arguments" 123 %}
//
// start_token will be the *Token with the tag's name in it (here: your_tag_name).
//
// Please see the Parser documentation on how to use the parser.
// See RegisterTag()'s documentation for more information about
// writing a tag as well.
type TagParser func(doc *Parser, start *TToken, arguments *Parser) (INodeTag, *Error)

type tag struct {
	name   string
	parser TagParser
}

var tags map[string]*tag

func init() {
	tags = make(map[string]*tag)
}

// Registers a new tag. If there's already a tag with the same
// name, RegisterTag will panic. You usually want to call this
// function in the tag's init() function:
// http://golang.org/doc/effective_go.html#init
//
// See http://www.florian-schlachter.de/post/pongo2/ for more about
// writing filters and tags.
func RegisterTag(name string, parserFn TagParser) {
	_, existing := tags[name]
	if existing {
		panic(fmt.Sprintf("Tag with name '%s' is already registered.", name))
	}
	tags[name] = &tag{
		name:   name,
		parser: parserFn,
	}
}

// Replaces an already registered tag with a new implementation. Use this
// function with caution since it allows you to change existing tag behaviour.
func ReplaceTag(name string, parserFn TagParser) {
	_, existing := tags[name]
	if !existing {
		panic(fmt.Sprintf("Tag with name '%s' does not exist (therefore cannot be overridden).", name))
	}
	tags[name] = &tag{
		name:   name,
		parser: parserFn,
	}
}
