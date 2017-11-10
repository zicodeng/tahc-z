package indexes

import (
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestSearch(t *testing.T) {

	cases := []struct {
		name                 string
		keys                 []string
		prefix               string
		resultsLimit         int
		expectedResultLength int
	}{
		{
			"keys with shared prefix",
			[]string{"do", "dog", "dope", "door", "desk", "cat"},
			"do",
			20,
			4,
		},
		{
			"keys with no shared prefix",
			[]string{"love", "big", "small"},
			"b",
			20,
			1,
		},
		{
			"empty prefix",
			[]string{"love", "big", "small"},
			"",
			20,
			0,
		},
		{
			"empty trie",
			[]string{},
			"",
			20,
			0,
		},
		{
			"exceeds results limit",
			[]string{"do", "dog", "dope", "door", "desk", "cat"},
			"d",
			3,
			3,
		},
		{
			"duplicated keys",
			[]string{"dog", "dog", "dog", "door", "desk", "cat"},
			"do",
			4,
			4,
		},
		{
			"duplicated keys with results limit",
			[]string{"dog", "dog", "dog", "door", "desk", "cat"},
			"do",
			2,
			2,
		},
		{
			"different casting",
			[]string{"Dog", "DOG", "dog", "door", "deSk", "cat"},
			"d",
			20,
			5,
		},
	}

	for _, c := range cases {
		// For each case, construct a new trie.
		trie := NewTrie()

		// Build our test trie.
		for _, key := range c.keys {
			trie.Insert(key, bson.NewObjectId())
		}

		result := trie.Search(c.resultsLimit, c.prefix)
		if len(result) != c.expectedResultLength {
			t.Errorf("\ncase: %v\ngot: %v\nwant: %v", c.name, len(result), c.expectedResultLength)
		}
	}

	// A trie might store identical user ID but with different keys.
	specialCases := []struct {
		name                 string
		keys                 []string
		prefix               string
		expectedResultLength int
	}{
		{
			"different keys have same values",
			[]string{"dog", "do", "dope"},
			"do",
			1,
		},
	}

	for _, c := range specialCases {
		trie := NewTrie()

		userID := bson.NewObjectId()
		for _, key := range c.keys {
			trie.Insert(key, userID)
		}

		result := trie.Search(20, c.prefix)
		if len(result) != c.expectedResultLength {
			t.Errorf("\ncase: %v\ngot: %v\nwant: %v", c.name, len(result), c.expectedResultLength)
		}
	}
}

func TestRemove(t *testing.T) {

	// Fake values that represent user ID.
	values := []bson.ObjectId{
		bson.NewObjectId(),
		bson.NewObjectId(),
		bson.NewObjectId(),
		bson.NewObjectId(),
		bson.NewObjectId(),
	}

	cases := []struct {
		name                 string
		keys                 []string
		key                  string
		value                bson.ObjectId
		expectedResultLength int
	}{
		{
			"target node has child nodes",
			[]string{"dog", "do", "dope", "cat"},
			"do",
			values[1],
			2,
		},
		{
			"target node has no child nodes",
			[]string{"dog", "do", "dope", "cat"},
			"dog",
			values[0],
			0,
		},
		{
			"target node has multiple values",
			[]string{"do", "do", "do", "dog", "dope"},
			"do",
			values[0],
			4,
		},
		{
			"case-insensitive remove",
			[]string{"do", "do", "do", "dog", "dope"},
			"DO",
			values[0],
			4,
		},
		{
			"remove empty key",
			[]string{"do", "dooog"},
			"",
			values[0],
			0,
		},
		{
			"empty trie",
			[]string{},
			"do",
			values[1],
			0,
		},
	}

	for _, c := range cases {
		trie := NewTrie()

		for i, key := range c.keys {
			trie.Insert(key, values[i])
		}

		trie.Remove(c.key, c.value)

		result := trie.Search(20, c.key)
		if len(result) != c.expectedResultLength {
			t.Errorf("\ncase: %v\ngot: %v\nwant: %v", c.name, len(result), c.expectedResultLength)
		}
	}

	// Test removing dangling nodes.
	danglingCases := []struct {
		name           string
		keys           []string
		key            string
		value          bson.ObjectId
		testKey        string
		expectedOutput int // Expected child nodes length.
	}{
		{
			"remove the entry together with its node",
			[]string{"do", "dog"},
			"dog",
			values[1],
			"do",
			0,
		},
		{
			"remove multiple dangling nodes",
			[]string{"do", "dooog"},
			"dooog",
			values[1],
			"do",
			0,
		},
		{
			"remove multiple dangling nodes when parent node has multiple child nodes",
			[]string{"do", "dooog", "dot", "dog"},
			"dooog",
			values[1],
			"do",
			2,
		},
	}

	for _, c := range danglingCases {
		trie := NewTrie()

		for i, key := range c.keys {
			trie.Insert(key, values[i])
		}

		trie.Remove(c.key, c.value)

		// Find the node pointing to the last character in the test key.
		curNode := trie.root
		for _, char := range c.testKey {
			_, hasKey := curNode.children[char]
			if !hasKey {
				t.Error("error finding node")
			}
			curNode = curNode.children[char]
		}

		if len(curNode.children) != c.expectedOutput {
			t.Errorf("\ncase: %v\ngot: %v\nwant: %v", c.name, len(curNode.children), c.expectedOutput)
		}
	}
}
