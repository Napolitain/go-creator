package services

import (
	"context"
	"testing"

	"gocreator/internal/mocks"

	"github.com/stretchr/testify/mock"
)

// Benchmark single translation
func BenchmarkTranslationService_Translate(b *testing.B) {
	mockClient := new(mocks.MockOpenAIClient)
	logger := &mockLogger{}
	service := NewTranslationService(mockClient, logger)

	text := "This is a test text for translation benchmark"
	targetLang := "Spanish"
	ctx := context.Background()

	// Mock API response
	mockClient.On("ChatCompletion", mock.Anything, mock.Anything).
		Return("Este es un texto de prueba para el benchmark de traducción", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.Translate(ctx, text, targetLang)
	}
}

// Benchmark batch translation with 5 texts
func BenchmarkTranslationService_TranslateBatch_5Texts(b *testing.B) {
	mockClient := new(mocks.MockOpenAIClient)
	logger := &mockLogger{}
	service := NewTranslationService(mockClient, logger)

	texts := []string{
		"First text for batch translation",
		"Second text for batch translation",
		"Third text for batch translation",
		"Fourth text for batch translation",
		"Fifth text for batch translation",
	}
	targetLang := "Spanish"
	ctx := context.Background()

	// Mock API responses
	mockClient.On("ChatCompletion", mock.Anything, mock.Anything).
		Return("Traducción de texto", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.TranslateBatch(ctx, texts, targetLang)
	}
}

// Benchmark batch translation with 10 texts
func BenchmarkTranslationService_TranslateBatch_10Texts(b *testing.B) {
	mockClient := new(mocks.MockOpenAIClient)
	logger := &mockLogger{}
	service := NewTranslationService(mockClient, logger)

	texts := make([]string, 10)
	for i := range texts {
		texts[i] = "Text number " + string(rune(i)) + " for batch translation benchmark"
	}
	targetLang := "Spanish"
	ctx := context.Background()

	// Mock API responses
	mockClient.On("ChatCompletion", mock.Anything, mock.Anything).
		Return("Traducción de texto", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.TranslateBatch(ctx, texts, targetLang)
	}
}

// Benchmark batch translation with 20 texts
func BenchmarkTranslationService_TranslateBatch_20Texts(b *testing.B) {
	mockClient := new(mocks.MockOpenAIClient)
	logger := &mockLogger{}
	service := NewTranslationService(mockClient, logger)

	texts := make([]string, 20)
	for i := range texts {
		texts[i] = "Text number " + string(rune(i)) + " for batch translation benchmark"
	}
	targetLang := "Spanish"
	ctx := context.Background()

	// Mock API responses
	mockClient.On("ChatCompletion", mock.Anything, mock.Anything).
		Return("Traducción de texto", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.TranslateBatch(ctx, texts, targetLang)
	}
}
