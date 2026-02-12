package x

import (
	"testing"
)

func TestMin_Int32(t *testing.T) {
	tests := []struct {
		a, b, want int32
	}{
		{1, 2, 1},
		{2, 1, 1},
		{0, 0, 0},
		{-1, 1, -1},
		{-5, -3, -5},
	}
	for _, tt := range tests {
		if got := Min(tt.a, tt.b); got != tt.want {
			t.Errorf("Min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMin_Int64(t *testing.T) {
	tests := []struct {
		a, b, want int64
	}{
		{1, 2, 1},
		{2, 1, 1},
		{1000000000000, 2000000000000, 1000000000000},
	}
	for _, tt := range tests {
		if got := Min(tt.a, tt.b); got != tt.want {
			t.Errorf("Min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMin_Float32(t *testing.T) {
	tests := []struct {
		a, b, want float32
	}{
		{1.5, 2.5, 1.5},
		{2.5, 1.5, 1.5},
		{-1.5, 1.5, -1.5},
	}
	for _, tt := range tests {
		if got := Min(tt.a, tt.b); got != tt.want {
			t.Errorf("Min(%f, %f) = %f, want %f", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMin_Float64(t *testing.T) {
	tests := []struct {
		a, b, want float64
	}{
		{1.5, 2.5, 1.5},
		{3.14159, 2.71828, 2.71828},
	}
	for _, tt := range tests {
		if got := Min(tt.a, tt.b); got != tt.want {
			t.Errorf("Min(%f, %f) = %f, want %f", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMax_Int32(t *testing.T) {
	tests := []struct {
		a, b, want int32
	}{
		{1, 2, 2},
		{2, 1, 2},
		{0, 0, 0},
		{-1, 1, 1},
		{-5, -3, -3},
	}
	for _, tt := range tests {
		if got := Max(tt.a, tt.b); got != tt.want {
			t.Errorf("Max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMax_Int64(t *testing.T) {
	tests := []struct {
		a, b, want int64
	}{
		{1, 2, 2},
		{2, 1, 2},
		{1000000000000, 2000000000000, 2000000000000},
	}
	for _, tt := range tests {
		if got := Max(tt.a, tt.b); got != tt.want {
			t.Errorf("Max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMax_Float32(t *testing.T) {
	tests := []struct {
		a, b, want float32
	}{
		{1.5, 2.5, 2.5},
		{2.5, 1.5, 2.5},
		{-1.5, 1.5, 1.5},
	}
	for _, tt := range tests {
		if got := Max(tt.a, tt.b); got != tt.want {
			t.Errorf("Max(%f, %f) = %f, want %f", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMax_Float64(t *testing.T) {
	tests := []struct {
		a, b, want float64
	}{
		{1.5, 2.5, 2.5},
		{3.14159, 2.71828, 3.14159},
	}
	for _, tt := range tests {
		if got := Max(tt.a, tt.b); got != tt.want {
			t.Errorf("Max(%f, %f) = %f, want %f", tt.a, tt.b, got, tt.want)
		}
	}
}
