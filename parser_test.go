package parsec

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsify(t *testing.T) {

	t.Run("strings", func(t *testing.T) {
		require.Equal(t, "ff", Parsify("ff")(InputString("ffooo")).Token)
	})

	t.Run("parsers", func(t *testing.T) {
		require.Equal(t, "ff", Parsify(Chars("f"))(InputString("ffooo")).Token)
	})

	t.Run("parser funcs", func(t *testing.T) {
		node := Parsify(func(p *State) *Node {
			return &Node{Token: "hello"}
		})(InputString("ffooo"))

		require.Equal(t, "hello", node.Token)
	})

	t.Run("*parsers", func(t *testing.T) {
		var parser Parser
		parserfied := Parsify(&parser)
		parser = Chars("f")

		node := parserfied(InputString("ffooo"))
		require.Equal(t, "ff", node.Token)
	})

	require.Panics(t, func() {
		Parsify(1)
	})
}

func TestParsifyAll(t *testing.T) {
	parsers := ParsifyAll("ff", "gg")

	result := parsers[0](InputString("ffooo"))
	require.Equal(t, "ff", result.Token)

	result = parsers[1](InputString("ffooo"))
	require.Nil(t, result)
}

func TestExact(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		node, ps := runParser("foobar", Exact("fo"))
		require.Equal(t, "fo", node.Token)
		require.Equal(t, "obar", ps.Get())
	})

	t.Run("error", func(t *testing.T) {
		_, ps := runParser("foobar", Exact("bar"))
		require.Equal(t, "bar", ps.Error.Expected)
		require.Equal(t, 0, ps.Pos)
	})
}

func TestChars(t *testing.T) {
	t.Run("full match", func(t *testing.T) {
		node, ps := runParser("foobar", Chars("a-z"))
		require.Equal(t, "foobar", node.Token)
		require.Equal(t, "", ps.Get())
		require.False(t, ps.Errored())
	})

	t.Run("partial match", func(t *testing.T) {
		node, ps := runParser("a1b2c3d4efg", Chars("1-4d-a"))
		require.Equal(t, "a1b2c3d4", node.Token)
		require.Equal(t, "efg", ps.Get())
		require.False(t, ps.Errored())
	})

	t.Run("limited match", func(t *testing.T) {
		node, ps := runParser("a1b2c3d4efg", Chars("1-4d-a", 1, 2))
		require.Equal(t, "a1", node.Token)
		require.Equal(t, "b2c3d4efg", ps.Get())
		require.False(t, ps.Errored())
	})

	t.Run("no match", func(t *testing.T) {
		_, ps := runParser("ffffff", Chars("0-9"))
		require.Equal(t, "offset 0: Expected 0-9", ps.Error.Error())
		require.Equal(t, 0, ps.Pos)
	})

	t.Run("no match with min", func(t *testing.T) {
		_, ps := runParser("ffffff", Chars("0-9", 4))
		require.Equal(t, "0-9", ps.Error.Expected)
		require.Equal(t, 0, ps.Pos)
	})

	t.Run("test exact matches", func(t *testing.T) {
		node, ps := runParser("aaff", Chars("abcd"))
		require.Equal(t, "aa", node.Token)
		require.Equal(t, 2, ps.Pos)
		require.False(t, ps.Errored())
	})

	t.Run("test not matches", func(t *testing.T) {
		node, ps := runParser("aaff", NotChars("ff"))
		require.Equal(t, "aa", node.Token)
		require.Equal(t, 2, ps.Pos)
		require.False(t, ps.Errored())
	})

	require.Panics(t, func() {
		Chars("a-b", 1, 2, 3)
	})
}

func TestParseString(t *testing.T) {
	Y := Map("hello", func(n *Node) *Node { return &Node{Result: n.Token} })
	t.Run("partial match", func(t *testing.T) {
		result, remaining, err := ParseString(Y, "hello world")
		require.Equal(t, "hello", result)
		require.Equal(t, " world", remaining)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		result, remaining, err := ParseString(Y, "world")
		require.Nil(t, result)
		require.Equal(t, "world", remaining)
		require.Error(t, err)
		require.Equal(t, "offset 0: Expected hello", err.Error())
	})
}

func TestString(t *testing.T) {
	t.Run("test basic match", func(t *testing.T) {
		result, p := runParser(`"hello"`, String('"'))
		require.Equal(t, `hello`, result.Token)
		require.Equal(t, "", p.Get())
	})

	t.Run("test non match", func(t *testing.T) {
		_, p := runParser(`1`, String('"'))
		require.Equal(t, `"`, p.Error.Expected)
		require.Equal(t, `1`, p.Get())
	})

	t.Run("test unterminated string", func(t *testing.T) {
		_, p := runParser(`"hello `, String('"'))
		require.Equal(t, `"`, p.Error.Expected)
		require.Equal(t, `"hello `, p.Get())
	})

	t.Run("test escaping", func(t *testing.T) {
		result, p := runParser(`"hello \"world\""`, String('"'))
		require.Equal(t, `hello "world"`, result.Token)
		require.Equal(t, ``, p.Get())
	})
}

func TestWS(t *testing.T) {
	t.Run("consumes all whitespace", func(t *testing.T) {
		result, p := runParser("    asdf", WS)
		require.Nil(t, result)
		require.Equal(t, "asdf", p.Get())
		require.False(t, p.Errored())
	})

	t.Run("never errors", func(t *testing.T) {
		result, p := runParser("asdf", WS)
		require.Nil(t, result)
		require.Equal(t, "asdf", p.Get())
		require.False(t, p.Errored())
	})
}

func runParser(input string, parser Parser) (*Node, *State) {
	ps := InputString(input)
	result := parser(ps)
	return result, ps
}
