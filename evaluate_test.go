package grules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestEvaluate(t *testing.T) {
	var testJSON = `
	{
		"name": {"first": "anakin", "last": "skywalker"},
		"age": 22,
		"children": ["luke", "leia"],
		"order": "jedi",
		"friends": [
			{"first": "r2d2",  "last": "droid",      "order": "republic", "age": 13, "episodes": [1,2,3,4,5,6,7,8,9]},
			{"first": "ben",   "last": "kenobi",     "order": "jedi",     "age": 38, "episodes": [1,2,3,4,5,6]},
			{"first": "c3po",  "last": "droid",      "order": "republic", "age": 13, "episodes": [1,2,3,4,5,6,7,8,9]},
			{"first": "sheev", "last": "palpatine",  "order": "sith",     "age": 63, "episodes": [1,2,3,5,6,9]}
		]
	}
	`

	testCases := []struct {
		desc     string
		rule     string
		expected bool
	}{
		{
			desc: "standard evaluate",
			rule: `
			{
				"comparer": "eq",
				"path": "name.first",
				"value": "anakin"
			}
			`,
			expected: true,
		},
		{
			desc: "fail evaluation",
			rule: `
			{
				"comparer": "eq",
				"path": "name.first",
				"value": "ANAKIN"
			}
			`,
			expected: false,
		},
		{
			desc: "arrays",
			rule: `
			{
				"comparer": "eq",
				"operator": "or",
				"path": "children",
				"value": "luke"
			}
			`,
			expected: true,
		},
		{
			desc: "arrays all contain",
			rule: `
			{
				"comparer": "contains",
				"operator": "and",
				"path": "friends.#.episodes",
				"value": 1
			}
			`,
			expected: true,
		},
		{
			desc: "multiple rules with 'and' operator",
			rule: `
			{
				"operator": "and",
				"rules": [
					{
						"path": "name.first",
						"comparer": "eq",
						"value": "anakin"
					},
					{
						"path": "age",
						"comparer": "gt",
						"value": 20
					}
				]
			}
			`,
			expected: true,
		},
		{
			desc: "multi nested complicated example",
			rule: `
			{
				"operator": "or",
				"rules": [
					{
						"operator": "and",
						"rules": [
							{
								"path": "name.first",
								"comparer": "eq",
								"value": "darth"
							},
							{
								"path": "name.last",
								"comparer": "eq",
								"value": "vader"
							}
						]
					},
					{
						"operator": "or",
						"rules": [
							{
								"path": "order",
								"comparer": "eq",
								"value": "first world order"
							},
							{
								"operator": "or",
								"path": "friends.#.order",
								"comparer": "contains",
								"value": "sith"
							}
						]
					}
				]
			}
			`,
			expected: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, _ := Evaluate(testJSON, tc.rule)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestEvaluateObject(t *testing.T) {
	testCases := []struct {
		desc     string
		object   gjson.Result
		rule     Rule
		expected bool
	}{
		{
			desc: "standard evaluate",
			object: gjson.Parse(`
			{
				"person": {
					"firstName": "stephen", 
					"lastName": "stanton"
				}
			}
			`),
			rule: Rule{
				Comparer: "eq",
				Path:     "person.firstName",
				Value:    "stephen",
			},
			expected: true,
		},
		{
			desc: "path does not exist",
			object: gjson.Parse(`
			{
				"person": {
					"firstName": "stephen", 
					"lastName": "stanton"
				}
			}
			`),
			rule: Rule{
				Comparer: "eq",
				Path:     "person.age",
				Value:    29,
			},
			expected: false,
		},
		{
			desc: "no comparer found",
			object: gjson.Parse(`
			{
				"person": {
					"firstName": "stephen", 
					"lastName": "stanton"
				}
			}
			`),
			rule: Rule{
				Comparer: "fubar",
				Path:     "person.age",
				Value:    29,
			},
			expected: false,
		},
		{
			desc: "with array",
			object: gjson.Parse(`
			{
				"names": ["stephen","david","stanton"]
			}
			`),
			rule: Rule{
				Comparer: "eq",
				Operator: Or,
				Path:     "names",
				Value:    "stephen",
			},
			expected: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := evaluateObject(tc.object, tc.rule)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestEvaluateMultiRule(t *testing.T) {
	testCases := []struct {
		desc     string
		object   gjson.Result
		rules    []Rule
		operator Operator
		expected bool
	}{
		{
			desc: "evaluate multiple or rules -- true",
			object: gjson.Parse(`
			{
				"first": "luke",
				"last": "skywalker"
			}
			`),
			rules: []Rule{
				{
					Path:     "first",
					Comparer: "eq",
					Value:    "anakin",
				},
				{
					Path:     "last",
					Comparer: "eq",
					Value:    "skywalker",
				},
			},
			operator: Or,
			expected: true,
		},
		{
			desc: "evaluate multiple or rules -- false",
			object: gjson.Parse(`
			{
				"first": "sheev",
				"last": "palpatine"
			}
			`),
			rules: []Rule{
				{
					Path:     "first",
					Comparer: "eq",
					Value:    "anakin",
				},
				{
					Path:     "last",
					Comparer: "eq",
					Value:    "skywalker",
				},
			},
			operator: Or,
			expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := evaluateMultiRule(tc.object, tc.rules, tc.operator)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestEvaluatePrimitive(t *testing.T) {
	testCases := []struct {
		desc     string
		object   gjson.Result
		rule     Rule
		compare  Compare // compare function should eval to true if values passed in correctly
		expected bool
	}{
		{
			desc:   "standard string",
			object: gjson.Parse(`{"name": "stephen"}`),
			rule: Rule{
				Path:  "name",
				Value: "stanton",
			},
			compare: func(a, b interface{}) bool {
				return a == "stephen" && b == "stanton"
			},
			expected: true,
		},
		{
			desc:   "standard number",
			object: gjson.Parse(`{"age": 21}`),
			rule: Rule{
				Path:  "age",
				Value: float64(42),
			},
			compare: func(a, b interface{}) bool {
				return a == float64(21) && b == float64(42)
			},
			expected: true,
		},
		{
			desc:   "standard boolean",
			object: gjson.Parse(`{"isCool": false}`),
			rule: Rule{
				Path:  "isCool",
				Value: true,
			},
			compare: func(a, b interface{}) bool {
				return a == false && b == true
			},
			expected: true,
		},
		{
			desc:   "dealing with arrays",
			object: gjson.Parse(`{"greatMovies": ["tenet", "inception", "interstellar"]}`),
			rule: Rule{
				Path:  "greatMovies",
				Value: "tenet",
			},
			compare:  contains,
			expected: true,
		},
		{
			desc:   "non primitive rule value",
			object: gjson.Parse(`{"isCool": false}`),
			rule: Rule{
				Path:  "isCool",
				Value: struct{ ID int }{ID: 42},
			},
			compare: func(a, b interface{}) bool {
				gotWhatWeWanted := a == false && b == struct{ ID int }{ID: 42}

				return !gotWhatWeWanted // inverse if we didn't get what we want
			},
			expected: false,
		},
		{
			desc:   "passing in a non primitive",
			object: gjson.Parse(`{"person": {"firstName": "stephen"}}`),
			rule: Rule{
				Path:  "person", // this will be an object
				Value: "stephen",
			},
			compare: func(a, b interface{}) bool {
				// since we say the value type is a string, gjson will pass in
				// the string version of the person, which will just be an empty string
				//lint:ignore S1008 this is more expressive
				if a == "" {
					return false // we expect this
				}

				return true // if we get this, something went wrong
			},
			expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			value := tc.object.Get(tc.rule.Path)
			result := evaluatePrimitive(value, tc.rule, tc.compare)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestEvaluateArrayOfPrimitives(t *testing.T) {
	testCases := []struct {
		desc     string
		object   gjson.Result
		rule     Rule
		compare  Compare
		expected bool
	}{
		{
			desc: "standard 'or' -- true",
			object: gjson.Parse(`
			{
				"family": [
					{
						"firstName": "ben",
						"lastName": "kenobi"
					},
					{
						"firstName": "luke",
						"lastName": "skywalker"
					},
					{
						"firstName": "mace",
						"lastName": "windu"
					}
				]
			}
			`),
			rule: Rule{
				Operator: Or,
				Path:     "family.#.lastName",
				Value:    "windu",
			},
			compare:  equal,
			expected: true,
		},
		{
			desc: "standard 'and' -- false",
			object: gjson.Parse(`
			{
				"family": [
					{
						"firstName": "ben",
						"lastName": "kenobi"
					},
					{
						"firstName": "luke",
						"lastName": "skywalker"
					},
					{
						"firstName": "mace",
						"lastName": "windu"
					}
				]
			}
			`),
			rule: Rule{
				Operator: Or,
				Path:     "family.#.lastName",
				Value:    "palpatine",
			},
			compare:  equal,
			expected: false,
		},
		{
			desc: "standard 'and' -- true",
			object: gjson.Parse(`
			{
				"family": [
					{
						"firstName": "anakin",
						"lastName": "skywalker"
					},
					{
						"firstName": "luke",
						"lastName": "skywalker"
					},
					{
						"firstName": "leia",
						"lastName": "skywalker"
					}
				]
			}
			`),
			rule: Rule{
				Operator: And,
				Path:     "family.#.lastName",
				Value:    "skywalker",
			},
			compare:  equal,
			expected: true,
		},
		{
			desc: "standard 'and' -- false",
			object: gjson.Parse(`
			{
				"family": [
					{
						"firstName": "anakin",
						"lastName": "skywalker"
					},
					{
						"firstName": "luke",
						"lastName": "skywalker"
					},
					{
						"firstName": "leia",
						"lastName": "organa"
					}
				]
			}
			`),
			rule: Rule{
				Operator: And,
				Path:     "family.#.lastName",
				Value:    "skywalker",
			},
			compare:  equal,
			expected: false,
		},
		{
			desc: "default to 'and' -- true",
			object: gjson.Parse(`
			{
				"family": [
					{
						"firstName": "anakin",
						"lastName": "skywalker"
					},
					{
						"firstName": "luke",
						"lastName": "skywalker"
					},
					{
						"firstName": "leia",
						"lastName": "skywalker"
					}
				]
			}
			`),
			rule: Rule{
				// No Operator
				Path:  "family.#.lastName",
				Value: "skywalker",
			},
			compare:  equal,
			expected: true,
		},
		{
			desc: "default to 'and' -- false",
			object: gjson.Parse(`
			{
				"family": [
					{
						"firstName": "anakin",
						"lastName": "skywalker"
					},
					{
						"firstName": "luke",
						"lastName": "skywalker"
					},
					{
						"firstName": "leia",
						"lastName": "organa"
					}
				]
			}
			`),
			rule: Rule{
				// No Operator
				Path:  "family.#.lastName",
				Value: "skywalker",
			},
			compare:  equal,
			expected: false,
		},
		{
			desc: "work with normal arrays too",
			object: gjson.Parse(`
			{
				"lightsaberColors": ["red","green","black","purple"]
			}
			`),
			rule: Rule{
				Operator: Or,
				Path:     "lightsaberColors",
				Value:    "purple",
			},
			compare:  equal,
			expected: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			value := tc.object.Get(tc.rule.Path)
			assert.True(t, value.IsArray(), "test is broken, path in rule does not return an array")

			result := evaluateArrayOfPrimitives(value.Array(), tc.rule, tc.compare)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTypeMatches(t *testing.T) {
	testCases := []struct {
		desc           string
		rule           Rule
		input          gjson.Result
		expectedResult bool
	}{
		{
			desc: "same number type",
			rule: Rule{
				Value: float64(42),
			},
			input:          gjson.Get(`{"age":42}`, "age"),
			expectedResult: true,
		},
		{
			desc: "different number type",
			rule: Rule{
				Value: float64(42),
			},
			input:          gjson.Get(`{"age":"42"}`, "age"),
			expectedResult: false,
		},
		{
			desc: "same string type",
			rule: Rule{
				Value: "stephen",
			},
			input:          gjson.Get(`{"name":"stephen"}`, "name"),
			expectedResult: true,
		},
		{
			desc: "different string type",
			rule: Rule{
				Value: "42",
			},
			input:          gjson.Get(`{"name":42}`, "name"),
			expectedResult: false,
		},
		{
			desc: "same boolean type",
			rule: Rule{
				Value: false,
			},
			input:          gjson.Get(`{"isCool":false}`, "isCool"),
			expectedResult: true,
		},
		{
			desc: "different boolean type",
			rule: Rule{
				Value: false,
			},
			input:          gjson.Get(`{"isCool":0}`, "isCool"),
			expectedResult: false,
		},
		{
			desc: "string version of bool",
			rule: Rule{
				Value: true,
			},
			input:          gjson.Get(`{"isCool":"true"}`, "foo"),
			expectedResult: false,
		},
		{
			desc:           "default false",
			input:          gjson.Get(`{"isCool":1}`, "isCool"),
			expectedResult: false,
		},
		{
			desc: "same types for array",
			rule: Rule{
				Value: float64(42),
			},
			input: gjson.Get(`
			{
				"members": [
					{"age":42},
					{"age":43},
					{"age":44}
				]
			}
			`, "members.#.age"),
			expectedResult: true,
		},
		{
			desc: "different types for array",
			rule: Rule{
				Value: float64(42),
			},
			input: gjson.Get(`
			{
				"members": [
					{"age":42},
					{"age":43},
					{"age":"44"}
				]
			}
			`, "members.#.age"),
			expectedResult: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := typeMatches(tc.input, tc.rule)

			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestTransformGJSONArrayToSlice(t *testing.T) {
	testCases := []struct {
		desc     string
		value    gjson.Result
		expected []interface{}
	}{
		{
			desc:     "array of strings",
			value:    gjson.Get(`{"strings":["foo","bar"]`, "strings"),
			expected: []interface{}{"foo", "bar"},
		},
		{
			desc:     "array of numbers",
			value:    gjson.Get(`{"numbers":[1,2,3]`, "numbers"),
			expected: []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			desc:     "array of bools",
			value:    gjson.Get(`{"bools":[true, false, true]`, "bools"),
			expected: []interface{}{true, false, true},
		},
		{
			desc:     "non array",
			value:    gjson.Get(`{"fizz":"buzz"`, "fizz"),
			expected: []interface{}{"buzz"}, // this is what will happen if you don't pass in an actual array
		},
		{
			desc:     "mixed types",
			value:    gjson.Get(`{"mixedTypes":[true, 42, "foo"]`, "mixedTypes"),
			expected: []interface{}{true, float64(42), "foo"},
		},
		{
			desc:     "array of arrays",
			value:    gjson.Get(`{"arrayOfArrays":[["foo", "bar"], ["fizz","buzz"]]`, "arrayOfArrays"),
			expected: []interface{}{[]interface{}{"foo", "bar"}, []interface{}{"fizz", "buzz"}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := transformGJSONArrayToSlice(tc.value.Array())

			assert.Equal(t, tc.expected, result)
		})
	}
}
