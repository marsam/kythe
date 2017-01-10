// Package comment contains comments.
package comment

//- Pkg.node/kind package
//- PkgDoc.node/kind doc
//- PkgDoc documents Pkg
//- PkgDoc.text "Package comment contains comments."

//- @+3"foo" defines/binding Foo?

/* foo comment */
var foo int

//- Foo.node/kind variable
//- FooDoc.node/kind doc
//- FooDoc.text "foo comment"
//- FooDoc documents Foo

// General constants.
const (
	//- @+2"bar" defines/binding Bar

	bar = 5

	//- @+3"baz" defines/binding Baz

	// baz comment
	baz = "quux"
)

// Names without their own comments inherit the block comment, if present.
// This may result in duplicating the comment text to different nodes.
//
//- Bar.node/kind constant
//- BarComment.node/kind doc
//- BarComment.text "General constants."
//- BarComment documents Bar

// Names with their own comments prefer them.
//- Baz.node/kind constant
//- BazComment.node/kind doc
//- BazComment.text "baz comment"
//- BazComment documents Baz

//- @+3"alpha" defines/binding Alpha

// alpha is a function.
func alpha() {}

//- Alpha.node/kind function
//- AlphaDoc.node/kind doc
//- AlphaDoc.text "alpha is a function."
//- AlphaDoc documents Alpha

//- @+3"widget" defines/binding Widget

// widget comment
type widget struct {}

// TODO(fromberger): Handle fields.

//- Widget.node/kind record
//- WidgetComment.node/kind doc
//- WidgetComment.text "widget comment"
//- WidgetComment documents Widget
