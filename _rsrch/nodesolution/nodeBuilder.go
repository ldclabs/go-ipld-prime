package ipld

// NodeAssembler is the interface that describes all the ways we can set values
// in a node that's under construction.
//
// To create a Node, you should start with a NodeBuilder (which contains a
// superset of the NodeAssembler methods, and can return the finished Node
// from its `Build` method).
//
// Why do both this and the NodeBuilder interface exist?
// When creating trees of nodes, recursion works over the NodeAssembler interface.
// This is important to efficient library internals, because avoiding the
// requirement to be able to return a Node at any random point in the process
// relieves internals from needing to implement 'freeze' features.
// (This is useful in turn because implementing those 'freeze' features in a
// language without first-class/compile-time support for them (as golang is)
// would tend to push complexity and costs to execution time; we'd rather not.)
type NodeAssembler interface {
	BeginMap(sizeHint int) (MapNodeAssembler, error)
	BeginList(sizeHint int) (ListNodeAssembler, error)
	AssignNull() error
	AssignBool(bool) error
	AssignInt(int) error
	AssignFloat(float64) error
	AssignString(string) error
	AssignBytes([]byte) error
	AssignLink(Link) error

	AssignNode(Node) error // if you already have a completely constructed subtree, this method puts the whole thing in place at once.

	// Style returns a NodeStyle describing what kind of value we're assembling.
	//
	// You often don't need this (because you should be able to
	// just feed data and check errors), but it's here.
	//
	// Using `this.Style().NewBuilder()` to produce a new `Node`,
	// then giving that node to `this.AssignNode(n)` should always work.
	// (Note that this is not necessarily an _exclusive_ statement on what
	// sort of values will be accepted by `this.AssignNode(n)`.)
	Style() NodeStyle
}

// MapNodeAssembler assembles a map node!  (You guessed it.)
//
// Methods on MapNodeAssembler must be called in a valid order:
// assemble a key, then assemble a value, then loop as long as desired;
// when finished, call 'Finish'.
//
// Incorrect order invocations will panic.
// Calling AssembleKey twice in a row will panic;
// calling AssembleValue before finishing using the NodeAssembler from AssembleKey will panic;
// calling AssembleValue twice in a row will panic;
// etc.
//
// Note that the NodeAssembler yielded from AssembleKey has additional behavior:
// if the node assembled there matches a key already present in the map,
// that assembler will emit the error!
type MapNodeAssembler interface {
	AssembleKey() NodeAssembler   // must be followed by call to AssembleValue.
	AssembleValue() NodeAssembler // must be called immediately after AssembleKey.

	AssembleDirectly(k string) (NodeAssembler, error) // shortcut combining AssembleKey and AssembleValue into one step; valid when the key is a string kind.

	Finish() error

	// KeyStyle returns a NodeStyle that knows how to build keys of a type this map uses.
	//
	// You often don't need this (because you should be able to
	// just feed data and check errors), but it's here.
	//
	// For all Data Model maps, this will answer with a basic concept of "string".
	// For Schema typed maps, this may answer with a more complex type (potentially even a struct type).
	KeyStyle() NodeStyle

	// ValueStyle returns a NodeStyle that knows how to build values this map can contain.
	//
	// You often don't need this (because you should be able to
	// just feed data and check errors), but it's here.
	//
	// ValueStyle requires a parameter describing the key in order to say what
	// NodeStyle will be acceptable as a value for that key, because when using
	// struct types (or union types) from the Schemas system, they behave as maps
	// but have different acceptable types for each field (or member, for unions).
	// For plain maps (that is, not structs or unions masquerading as maps),
	// the empty string can be used as a parameter, and the returned NodeStyle
	// can be assumed applicable for all values.
	// Using an empty string for a struct or union will return a nil NodeStyle.
	// (Design note: a string is sufficient for the parameter here rather than
	// a full Node, because the only cases where the value types vary are also
	// cases where the keys may not be complex.)
	ValueStyle(k string) NodeStyle
}

type ListNodeAssembler interface {
	AssembleValue() NodeAssembler

	Finish() error

	// ValueStyle returns a NodeStyle that knows how to build values this map can contain.
	//
	// You often don't need this (because you should be able to
	// just feed data and check errors), but it's here.
	//
	// In contrast to the `MapNodeAssembler.ValueStyle(key)` function,
	// to determine the ValueStyle for lists we need no parameters;
	// lists always contain one value type (even if it's "any").
	ValueStyle() NodeStyle
}

type NodeBuilder interface {
	NodeAssembler

	// Build returns the new value after all other assembly has been completed.
	//
	// A method on the NodeAssembler that finishes assembly of the data must
	// be called first (e.g., any of the "Assign*" methods, or "Finish" if
	// the assembly was for a map or a list); that finishing method still has
	// all responsibility for validating the assembled data and returning
	// any errors from that process.
	// (Correspondingly, there is no error return from this method.)
	Build() Node

	// Resets the builder.  It can hereafter be used again.
	// Reusing a NodeBuilder can reduce allocations and improve performance.
	//
	// Only call this if you're going to reuse the builder.
	// (Otherwise, it's unnecessary, and may cause an unwanted allocation).
	Reset()
}
