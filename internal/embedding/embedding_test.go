package embedding

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a, b []float32
		want float32
	}{
		{
			name: "identical vectors",
			a:    []float32{1, 0, 0},
			b:    []float32{1, 0, 0},
			want: 1.0,
		},
		{
			name: "opposite vectors",
			a:    []float32{1, 0, 0},
			b:    []float32{-1, 0, 0},
			want: -1.0,
		},
		{
			name: "orthogonal vectors",
			a:    []float32{1, 0, 0},
			b:    []float32{0, 1, 0},
			want: 0.0,
		},
		{
			name: "similar vectors",
			a:    []float32{1, 1, 0},
			b:    []float32{1, 0, 0},
			want: float32(1 / math.Sqrt(2)),
		},
		{
			name: "empty vectors",
			a:    []float32{},
			b:    []float32{},
			want: 0.0,
		},
		{
			name: "mismatched lengths",
			a:    []float32{1, 2},
			b:    []float32{1, 2, 3},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CosineSimilarity(tt.a, tt.b)
			if diff := math.Abs(float64(got - tt.want)); diff > 0.0001 {
				t.Errorf("CosineSimilarity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCosineDistance(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{1, 0, 0}

	dist := CosineDistance(a, b)
	if dist != 0 {
		t.Errorf("CosineDistance() = %v, want 0", dist)
	}

	c := []float32{-1, 0, 0}
	dist = CosineDistance(a, c)
	if math.Abs(float64(dist-2)) > 0.0001 {
		t.Errorf("CosineDistance() = %v, want 2", dist)
	}
}

func TestSerializeDeserialize(t *testing.T) {
	original := []float32{1.5, -2.3, 0, 42.0, -0.001}

	serialized := SerializeFloat32(original)
	if len(serialized) != len(original)*4 {
		t.Errorf("SerializeFloat32() produced %d bytes, want %d", len(serialized), len(original)*4)
	}

	deserialized := DeserializeFloat32(serialized)
	if len(deserialized) != len(original) {
		t.Errorf("DeserializeFloat32() produced %d elements, want %d", len(deserialized), len(original))
	}

	for i, val := range deserialized {
		if val != original[i] {
			t.Errorf("DeserializeFloat32()[%d] = %v, want %v", i, val, original[i])
		}
	}
}

func TestDeserializeInvalidData(t *testing.T) {
	// Not a multiple of 4 bytes
	result := DeserializeFloat32([]byte{1, 2, 3})
	if result != nil {
		t.Errorf("DeserializeFloat32() = %v, want nil for invalid data", result)
	}
}
