package build

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseSimpleEvals tests parsing valid evals from JSON
func TestParseSimpleEvals(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-parse-evals-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	evalsJSON := `[
		{
			"name": "test1",
			"input": {"prompt": "hello"},
			"expected": {"response": "hi"}
		},
		{
			"name": "test2",
			"input": {"prompt": "bye"},
			"expected": {"response": "goodbye"}
		}
	]`

	evalsPath := tmpDir + "/evals.json"
	err = os.WriteFile(evalsPath, []byte(evalsJSON), 0644)
	require.NoError(t, err)

	evalResults, err := ParseSimpleEvals(evalsPath)
	require.NoError(t, err)
	assert.Len(t, evalResults, 2)
	assert.Equal(t, "test1", evalResults[0].Name)
	assert.Equal(t, "test2", evalResults[1].Name)
}

// TestParseSimpleEvalsEmpty tests parsing empty evals array
func TestParseSimpleEvalsEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-empty-evals-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	evalsJSON := `[]`

	evalsPath := tmpDir + "/evals.json"
	err = os.WriteFile(evalsPath, []byte(evalsJSON), 0644)
	require.NoError(t, err)

	evalResults, err := ParseSimpleEvals(evalsPath)
	require.NoError(t, err)
	assert.Len(t, evalResults, 0)
}

// TestParseSimpleEvalsInvalidJSON tests parsing invalid JSON
func TestParseSimpleEvalsInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-invalid-evals-json-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	invalidJSON := `this is not valid JSON [`

	evalsPath := tmpDir + "/evals.json"
	err = os.WriteFile(evalsPath, []byte(invalidJSON), 0644)
	require.NoError(t, err)

	_, err = ParseSimpleEvals(evalsPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse evals JSON")
}

// TestParseSimpleEvalsFileNotFound tests parsing from non-existent file
func TestParseSimpleEvalsFileNotFound(t *testing.T) {
	_, err := ParseSimpleEvals("/non/existent/evals.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read evals file")
}

// TestRunSimpleEval tests running a simple evaluation
func TestRunSimpleEval(t *testing.T) {
	eval := SimpleEval{
		Name: "test1",
		Input: map[string]any{
			"prompt": "hello",
		},
		Expected: map[string]any{
			"response": "hi",
		},
	}

	actual := map[string]any{
		"response": "hi",
	}

	result := RunSimpleEval(eval, actual)
	assert.True(t, result.Passed)
	assert.Nil(t, result.Error)
	assert.Equal(t, eval.Name, result.Eval.Name)
	assert.Equal(t, actual, result.Actual)
}

// TestRunSimpleEvalFailure tests running an evaluation that fails
func TestRunSimpleEvalFailure(t *testing.T) {
	eval := SimpleEval{
		Name: "test1",
		Input: map[string]any{
			"prompt": "hello",
		},
		Expected: map[string]any{
			"response": "hi",
		},
	}

	actual := map[string]any{
		"response": "bye",
	}

	result := RunSimpleEval(eval, actual)
	assert.False(t, result.Passed)
	assert.Nil(t, result.Error)
}

// TestRunSimpleEvalEmptyMaps tests running evaluation with empty maps
func TestRunSimpleEvalEmptyMaps(t *testing.T) {
	eval := SimpleEval{
		Name:     "empty-test",
		Input:    map[string]any{},
		Expected: map[string]any{},
	}

	actual := map[string]any{}

	result := RunSimpleEval(eval, actual)
	assert.True(t, result.Passed)
}

// TestRunSimpleEvalMissingKey tests evaluation when actual is missing a key
func TestRunSimpleEvalMissingKey(t *testing.T) {
	eval := SimpleEval{
		Name: "missing-key-test",
		Input: map[string]any{
			"prompt": "test",
		},
		Expected: map[string]any{
			"response": "answer",
			"extra":    "data",
		},
	}

	actual := map[string]any{
		"response": "answer",
	}

	result := RunSimpleEval(eval, actual)
	assert.False(t, result.Passed)
}

// TestRunSimpleEvalExtraKey tests evaluation when actual has extra key
func TestRunSimpleEvalExtraKey(t *testing.T) {
	eval := SimpleEval{
		Name: "extra-key-test",
		Input: map[string]any{
			"prompt": "test",
		},
		Expected: map[string]any{
			"response": "answer",
		},
	}

	actual := map[string]any{
		"response": "answer",
		"extra":    "data",
	}

	result := RunSimpleEval(eval, actual)
	assert.False(t, result.Passed)
}

// TestDeepEqualMatching tests deepEqual with matching maps
func TestDeepEqualMatching(t *testing.T) {
	a := map[string]any{
		"name": "test",
		"value": 123,
		"flag": true,
	}

	b := map[string]any{
		"name": "test",
		"value": 123,
		"flag": true,
	}

	result := deepEqual(a, b)
	assert.True(t, result)
}

// TestDeepEqualDifferentValues tests deepEqual with different values
func TestDeepEqualDifferentValues(t *testing.T) {
	a := map[string]any{
		"name": "test",
		"value": 123,
	}

	b := map[string]any{
		"name": "test",
		"value": 456,
	}

	result := deepEqual(a, b)
	assert.False(t, result)
}

// TestDeepEqualDifferentLengths tests deepEqual with different map lengths
func TestDeepEqualDifferentLengths(t *testing.T) {
	a := map[string]any{
		"name": "test",
		"value": 123,
	}

	b := map[string]any{
		"name": "test",
	}

	result := deepEqual(a, b)
	assert.False(t, result)
}

// TestDeepEqualMissingKey tests deepEqual when one map is missing a key
func TestDeepEqualMissingKey(t *testing.T) {
	a := map[string]any{
		"name": "test",
		"value": 123,
	}

	b := map[string]any{
		"name": "test",
		"other": 456,
	}

	result := deepEqual(a, b)
	assert.False(t, result)
}

// TestDeepEqualNestedMaps tests deepEqual with nested maps
func TestDeepEqualNestedMaps(t *testing.T) {
	a := map[string]any{
		"name": "test",
		"nested": map[string]any{
			"key1": "value1",
			"key2": 123,
		},
	}

	b := map[string]any{
		"name": "test",
		"nested": map[string]any{
			"key1": "value1",
			"key2": 123,
		},
	}

	result := deepEqual(a, b)
	assert.True(t, result)
}

// TestDeepEqualNestedMapsDifferent tests deepEqual with different nested maps
func TestDeepEqualNestedMapsDifferent(t *testing.T) {
	a := map[string]any{
		"name": "test",
		"nested": map[string]any{
			"key1": "value1",
		},
	}

	b := map[string]any{
		"name": "test",
		"nested": map[string]any{
			"key1": "value2",
		},
	}

	result := deepEqual(a, b)
	assert.False(t, result)
}

// TestDeepEqualSlices tests deepEqual with slices
func TestDeepEqualSlices(t *testing.T) {
	a := map[string]any{
		"items": []any{"a", "b", "c"},
	}

	b := map[string]any{
		"items": []any{"a", "b", "c"},
	}

	result := deepEqual(a, b)
	assert.True(t, result)
}

// TestDeepEqualSlicesDifferent tests deepEqual with different slices
func TestDeepEqualSlicesDifferent(t *testing.T) {
	a := map[string]any{
		"items": []any{"a", "b", "c"},
	}

	b := map[string]any{
		"items": []any{"a", "b", "d"},
	}

	result := deepEqual(a, b)
	assert.False(t, result)
}

// TestDeepEqualEmptyMaps tests deepEqual with empty maps
func TestDeepEqualEmptyMaps(t *testing.T) {
	a := map[string]any{}
	b := map[string]any{}

	result := deepEqual(a, b)
	assert.True(t, result)
}

// TestDeepEqualArrays tests deepEqual with arrays
func TestDeepEqualArrays(t *testing.T) {
	a := map[string]any{
		"nums": []any{1, 2, 3},
	}

	b := map[string]any{
		"nums": []any{1, 2, 3},
	}

	result := deepEqual(a, b)
	assert.True(t, result)
}
