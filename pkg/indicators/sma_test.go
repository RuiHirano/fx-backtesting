package indicators

import (
	"math"
	"testing"
)

func TestNewSMA(t *testing.T) {
	sma := NewSMA(5)
	
	if sma.period != 5 {
		t.Errorf("Expected period 5, got %d", sma.period)
	}
	
	if len(sma.values) != 0 {
		t.Errorf("Expected empty values slice, got length %d", len(sma.values))
	}
}

func TestSMA_Update(t *testing.T) {
	sma := NewSMA(3)
	
	// Add first value
	value1 := sma.Update(10.0)
	if value1 != 10.0 {
		t.Errorf("Expected first value 10.0, got %v", value1)
	}
	
	// Add second value
	value2 := sma.Update(20.0)
	expected2 := (10.0 + 20.0) / 2.0
	if value2 != expected2 {
		t.Errorf("Expected second value %v, got %v", expected2, value2)
	}
	
	// Add third value (period reached)
	value3 := sma.Update(30.0)
	expected3 := (10.0 + 20.0 + 30.0) / 3.0
	if value3 != expected3 {
		t.Errorf("Expected third value %v, got %v", expected3, value3)
	}
	
	// Add fourth value (sliding window)
	value4 := sma.Update(40.0)
	expected4 := (20.0 + 30.0 + 40.0) / 3.0
	if value4 != expected4 {
		t.Errorf("Expected fourth value %v, got %v", expected4, value4)
	}
}

func TestSMA_GetValue(t *testing.T) {
	sma := NewSMA(3)
	
	// Before any updates
	if sma.GetValue() != 0.0 {
		t.Errorf("Expected initial value 0.0, got %v", sma.GetValue())
	}
	
	sma.Update(10.0)
	sma.Update(20.0)
	sma.Update(30.0)
	
	expected := 20.0
	if sma.GetValue() != expected {
		t.Errorf("Expected value %v, got %v", expected, sma.GetValue())
	}
}

func TestSMA_IsReady(t *testing.T) {
	sma := NewSMA(3)
	
	// Not ready initially
	if sma.IsReady() {
		t.Error("Expected SMA not to be ready initially")
	}
	
	sma.Update(10.0)
	if sma.IsReady() {
		t.Error("Expected SMA not to be ready after 1 value")
	}
	
	sma.Update(20.0)
	if sma.IsReady() {
		t.Error("Expected SMA not to be ready after 2 values")
	}
	
	sma.Update(30.0)
	if !sma.IsReady() {
		t.Error("Expected SMA to be ready after 3 values")
	}
}

func TestSMA_Reset(t *testing.T) {
	sma := NewSMA(3)
	
	sma.Update(10.0)
	sma.Update(20.0)
	sma.Update(30.0)
	
	if !sma.IsReady() {
		t.Error("Expected SMA to be ready before reset")
	}
	
	sma.Reset()
	
	if sma.IsReady() {
		t.Error("Expected SMA not to be ready after reset")
	}
	
	if len(sma.values) != 0 {
		t.Errorf("Expected empty values slice after reset, got length %d", len(sma.values))
	}
	
	if sma.GetValue() != 0.0 {
		t.Errorf("Expected value 0.0 after reset, got %v", sma.GetValue())
	}
}

func TestSMA_FloatingPointPrecision(t *testing.T) {
	sma := NewSMA(3)
	
	sma.Update(1.0/3.0)
	sma.Update(2.0/3.0)
	sma.Update(1.0)
	
	expected := (1.0/3.0 + 2.0/3.0 + 1.0) / 3.0
	actual := sma.GetValue()
	
	if math.Abs(actual-expected) > 1e-10 {
		t.Errorf("Expected value %v, got %v", expected, actual)
	}
}