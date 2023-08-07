package radix_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/livebud/router/internal/radix"
	"github.com/matryer/is"
	"github.com/matthewmueller/diff"
)

func insertEqual(t *testing.T, tree *radix.Tree, route string, expected string) {
	t.Helper()
	t.Run(route, func(t *testing.T) {
		t.Helper()
		if err := tree.Insert(route); err != nil {
			if err.Error() == expected {
				return
			}
			t.Fatal(err)
		}
		actual := strings.TrimSpace(tree.String())
		expected = strings.ReplaceAll(strings.TrimSpace(expected), "\t", "")
		if actual == expected {
			return
		}
		var b bytes.Buffer
		b.WriteString("\n\x1b[4mExpected\x1b[0m:\n")
		b.WriteString(expected)
		b.WriteString("\n\n")
		b.WriteString("\x1b[4mActual\x1b[0m: \n")
		b.WriteString(actual)
		b.WriteString("\n\n")
		b.WriteString("\x1b[4mDifference\x1b[0m: \n")
		b.WriteString(diff.String(expected, actual))
		b.WriteString("\n")
		t.Fatal(b.String())
	})
}

// https://en.wikipedia.org/wiki/Radix_tree#Insertion
func TestWikipediaInsert(t *testing.T) {
	tree := radix.New()
	insertEqual(t, tree, "/test", `
		/test [routable=/test]
	`)
	insertEqual(t, tree, "/slow", `
		/
		•test [routable=/test]
		•slow [routable=/slow]
	`)
	insertEqual(t, tree, "/water", `
		/
		•test [routable=/test]
		•slow [routable=/slow]
		•water [routable=/water]
	`)
	insertEqual(t, tree, "/slower", `
		/
		•test [routable=/test]
		•slow [routable=/slow]
		•••••er [routable=/slower]
		•water [routable=/water]
	`)
	tree = radix.New()
	insertEqual(t, tree, "/tester", `
		/tester [routable=/tester]
	`)
	insertEqual(t, tree, "/test", `
		/test [routable=/test]
		•••••er [routable=/tester]
	`)
	tree = radix.New()
	insertEqual(t, tree, "/test", `
		/test [routable=/test]
	`)
	insertEqual(t, tree, "/team", `
		/te
		•••st [routable=/test]
		•••am [routable=/team]
	`)
	insertEqual(t, tree, "/toast", `
		/t
		••e
		•••st [routable=/test]
		•••am [routable=/team]
		••oast [routable=/toast]
	`)
}

func TestSampleInsert(t *testing.T) {
	tree := radix.New()
	insertEqual(t, tree, "/hello/{name}", `
		/hello/{name} [routable=/hello/{name}]
	`)
	insertEqual(t, tree, "/howdy/{name}/", `
		/h
		••ello/{name} [routable=/hello/{name}]
		••owdy/{name} [routable=/howdy/{name}]
	`)
	insertEqual(t, tree, "/hello/{name}/elsewhere", `
		/h
		••ello/{name} [routable=/hello/{name}]
		•••••••••••••/elsewhere [routable=/hello/{name}/elsewhere]
		••owdy/{name} [routable=/howdy/{name}]
	`)
	insertEqual(t, tree, "/hello/{name}/admin/", `
		/h
		••ello/{name} [routable=/hello/{name}]
		•••••••••••••/
		••••••••••••••elsewhere [routable=/hello/{name}/elsewhere]
		••••••••••••••admin [routable=/hello/{name}/admin]
		••owdy/{name} [routable=/howdy/{name}]
	`)
	insertEqual(t, tree, "/hello/{name}/else/", `
		/h
		••ello/{name} [routable=/hello/{name}]
		•••••••••••••/
		••••••••••••••else [routable=/hello/{name}/else]
		••••••••••••••••••where [routable=/hello/{name}/elsewhere]
		••••••••••••••admin [routable=/hello/{name}/admin]
		••owdy/{name} [routable=/howdy/{name}]
	`)
}

func TestEquals(t *testing.T) {
	tree := radix.New()
	insertEqual(t, tree, "/hello/{name}", `
		/hello/{name} [routable=/hello/{name}]
	`)
	insertEqual(t, tree, "/hello/{name}", `route already exists: "/hello/{name}"`)
	insertEqual(t, tree, "/hello", `
		/hello [routable=/hello]
		••••••/{name} [routable=/hello/{name}]
	`)
	insertEqual(t, tree, "/hello", `route already exists: "/hello"`)
	tree = radix.New()
	insertEqual(t, tree, "/{name}", `
		/{name} [routable=/{name}]
	`)
	insertEqual(t, tree, "/{title}", `route already exists: "/{name}"`)
}

func TestDifferentSlots(t *testing.T) {
	tree := radix.New()
	insertEqual(t, tree, "/{name}", `
		/{name} [routable=/{name}]
	`)
	insertEqual(t, tree, "/{first}/{last}", `
		/{name} [routable=/{name}]
		•••••••/{last} [routable=/{first}/{last}]
	`)
	insertEqual(t, tree, "/{first}/else", `
		/{name} [routable=/{name}]
		•••••••/
		••••••••{last} [routable=/{first}/{last}]
		••••••••else [routable=/{first}/else]
	`)
	tree = radix.New()
	insertEqual(t, tree, "/{name}", `
		/{name} [routable=/{name}]
	`)
	insertEqual(t, tree, "/else", `
		/
		•{name} [routable=/{name}]
		•else [routable=/else]
	`)
}

func TestPathAfter(t *testing.T) {
	tree := radix.New()
	insertEqual(t, tree, "/{name}", `
		/{name} [routable=/{name}]
	`)
	insertEqual(t, tree, "/", `
		/ [routable=/]
		•{name} [routable=/{name}]
	`)
	insertEqual(t, tree, "/first/{name}", `
		/ [routable=/]
		•{name} [routable=/{name}]
		•first/{name} [routable=/first/{name}]
	`)
	insertEqual(t, tree, "/first", `
		/ [routable=/]
		•{name} [routable=/{name}]
		•first [routable=/first]
		••••••/{name} [routable=/first/{name}]
	`)
}

func TestOptionals(t *testing.T) {
	tree := radix.New()
	insertEqual(t, tree, "/{name?}", `
		/ [routable=/]
		•{name} [routable=/{name}]
	`)
	insertEqual(t, tree, "/first/{last?}", `
		/ [routable=/]
		•{name} [routable=/{name}]
		•first [routable=/first]
		••••••/{last} [routable=/first/{last}]
	`)
	insertEqual(t, tree, "/{first}/{last}", `
		/ [routable=/]
		•{name} [routable=/{name}]
		•••••••/{last} [routable=/{first}/{last}]
		•first [routable=/first]
		••••••/{last} [routable=/first/{last}]
	`)
	insertEqual(t, tree, "/first/else", `
		/ [routable=/]
		•{name} [routable=/{name}]
		•••••••/{last} [routable=/{first}/{last}]
		•first [routable=/first]
		••••••/
		•••••••{last} [routable=/first/{last}]
		•••••••else [routable=/first/else]
	`)
}

func TestWildcards(t *testing.T) {
	tree := radix.New()
	insertEqual(t, tree, "/{name*}", `
		/ [routable=/]
		•{name*} [routable=/{name*}]
	`)
	insertEqual(t, tree, "/first/{last*}", `
		/ [routable=/]
		•{name*} [routable=/{name*}]
		•first [routable=/first]
		••••••/{last*} [routable=/first/{last*}]
	`)
	insertEqual(t, tree, "/{first}/{last}", `
		/ [routable=/]
		•{name*} [routable=/{name*}]
		••••••••/{last} [routable=/{first}/{last}]
		•first [routable=/first]
		••••••/{last*} [routable=/first/{last*}]
	`)
	insertEqual(t, tree, "/first/else", `
		/ [routable=/]
		•{name*} [routable=/{name*}]
		••••••••/{last} [routable=/{first}/{last}]
		•first [routable=/first]
		••••••/
		•••••••{last*} [routable=/first/{last*}]
		•••••••else [routable=/first/else]
	`)
}

// TODO: test regexps

func TestRootSwap(t *testing.T) {
	tree := radix.New()
	insertEqual(t, tree, "/hello", `
		/hello [routable=/hello]
	`)
	insertEqual(t, tree, "/", `
		/ [routable=/]
		•hello [routable=/hello]
	`)
}

func matchEqual(t *testing.T, tree *radix.Tree, path string, expect string) {
	t.Helper()
	t.Run(path, func(t *testing.T) {
		t.Helper()
		match, err := tree.Match(path)
		if err != nil {
			if err.Error() == expect {
				return
			}
			t.Fatal(err.Error())
		}
		actual := match.String()
		if actual == expect {
			return
		}
		var b bytes.Buffer
		b.WriteString("\n\x1b[4mExpected\x1b[0m:\n")
		b.WriteString(expect)
		b.WriteString("\n\n")
		b.WriteString("\x1b[4mActual\x1b[0m: \n")
		b.WriteString(actual)
		b.WriteString("\n\n")
		b.WriteString("\x1b[4mDifference\x1b[0m: \n")
		b.WriteString(diff.String(expect, actual))
		b.WriteString("\n")
		t.Fatal(b.String())
	})
}

func TestMatch(t *testing.T) {
	is := is.New(t)
	tree := radix.New()
	is.NoErr(tree.Insert("/hello"))
	matchEqual(t, tree, "/hello", `/hello`)
	matchEqual(t, tree, "/hello/world", `no match for "/hello/world"`)
	matchEqual(t, tree, "/", `no match for "/"`)
	matchEqual(t, tree, "/hello/", `/hello`)
	is.NoErr(tree.Insert("/"))
	matchEqual(t, tree, "/hello", `/hello`)
	matchEqual(t, tree, "/hello/world", `no match for "/hello/world"`)
	matchEqual(t, tree, "/", `/`)
	matchEqual(t, tree, "/hello/", `/hello`)
}
