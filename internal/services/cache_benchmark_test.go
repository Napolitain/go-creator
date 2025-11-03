package services

import (
	"testing"
	"time"
)

// Benchmark cache set operation
func BenchmarkCacheService_Set(b *testing.B) {
	service := NewCacheService(5*time.Minute, 10*time.Minute)
	value := "test value for cache benchmark"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key_" + string(rune(i))
		service.Set(key, value)
	}
}

// Benchmark cache get operation with hit
func BenchmarkCacheService_Get_Hit(b *testing.B) {
	service := NewCacheService(5*time.Minute, 10*time.Minute)
	key := "test_key"
	value := "test value for cache benchmark"
	service.Set(key, value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.Get(key)
	}
}

// Benchmark cache get operation with miss
func BenchmarkCacheService_Get_Miss(b *testing.B) {
	service := NewCacheService(5*time.Minute, 10*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "nonexistent_key_" + string(rune(i))
		_, _ = service.Get(key)
	}
}

// Benchmark cache delete operation
func BenchmarkCacheService_Delete(b *testing.B) {
	service := NewCacheService(5*time.Minute, 10*time.Minute)
	value := "test value for cache benchmark"

	// Pre-populate cache
	for i := 0; i < b.N; i++ {
		key := "key_" + string(rune(i))
		service.Set(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key_" + string(rune(i))
		service.Delete(key)
	}
}

// Benchmark cache with expiration
func BenchmarkCacheService_WithExpiration(b *testing.B) {
	service := NewCacheService(1*time.Millisecond, 2*time.Millisecond)
	value := "test value for cache benchmark"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key_" + string(rune(i))
		service.Set(key, value)
		time.Sleep(2 * time.Millisecond) // Wait for expiration
		_, _ = service.Get(key)          // Should miss due to expiration
	}
}

// Benchmark mixed cache operations
func BenchmarkCacheService_MixedOperations(b *testing.B) {
	service := NewCacheService(5*time.Minute, 10*time.Minute)
	value := "test value for cache benchmark"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key_" + string(rune(i%100)) // Reuse keys to get some hits
		
		// Set
		service.Set(key, value)
		
		// Get
		_, _ = service.Get(key)
		
		// Delete occasionally
		if i%10 == 0 {
			service.Delete(key)
		}
	}
}
