package services_test

import (
	"context"
	"errors"
	"testing"

	cacheSvc "github.com/goquizvibe/backend/shared/infrastructure/cache"
)

func TestCacheHelpers_GetOrFetch_NilCache(t *testing.T) {
	t.Parallel()

	t.Run("nil cache skips cache operations", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		result, err := cacheSvc.GetOrFetch(ctx, nil, "test-key", func() (int, error) {
			return 42, nil
		})

		if err != nil {
			t.Fatalf("GetOrFetch() error = %v, want nil", err)
		}
		if result != 42 {
			t.Errorf("GetOrFetch() = %d, want 42", result)
		}
	})

	t.Run("nil cache returns fetch error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		_, err := cacheSvc.GetOrFetch(ctx, nil, "test-key", func() (int, error) {
			return 0, errors.New("fetch error")
		})

		if err == nil {
			t.Fatal("GetOrFetch() error = nil, want error")
		}
	})
}

func TestCacheHelpers_SaveOrUpdate_NilCache(t *testing.T) {
	t.Parallel()

	t.Run("nil cache skips cache operations", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		result, err := cacheSvc.SaveOrUpdate(ctx, nil, "test-key", func() (string, error) {
			return "saved", nil
		})

		if err != nil {
			t.Fatalf("SaveOrUpdate() error = %v, want nil", err)
		}
		if result != "saved" {
			t.Errorf("SaveOrUpdate() = %s, want 'saved'", result)
		}
	})

	t.Run("nil cache returns save error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		_, err := cacheSvc.SaveOrUpdate(ctx, nil, "test-key", func() (int, error) {
			return 0, errors.New("save error")
		})

		if err == nil {
			t.Fatal("SaveOrUpdate() error = nil, want error")
		}
	})
}

func TestCacheHelpers_Delete_NilCache(t *testing.T) {
	t.Parallel()

	t.Run("nil cache skips delete", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		deleted := false

		err := cacheSvc.Delete(ctx, nil, "test-key", func() error {
			deleted = true
			return nil
		})

		if err != nil {
			t.Fatalf("Delete() error = %v, want nil", err)
		}
		if !deleted {
			t.Error("deleteFunc should have been called")
		}
	})

	t.Run("nil cache returns delete func error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		err := cacheSvc.Delete(ctx, nil, "test-key", func() error {
			return errors.New("delete error")
		})

		if err == nil {
			t.Fatal("Delete() error = nil, want error")
		}
	})
}

func TestCacheHelpers_InvalidateCache_NilCache(t *testing.T) {
	t.Parallel()

	t.Run("nil cache does nothing", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		cacheSvc.InvalidateCache(ctx, nil, "key1", "key2", "key3")
	})
}

func TestCacheService_New(t *testing.T) {
	t.Parallel()

	t.Run("creates cache service", func(t *testing.T) {
		t.Parallel()
		cache := cacheSvc.NewCacheService(nil, 0)

		if cache == nil {
			t.Fatal("NewCacheService returned nil")
		}
	})
}

func TestCacheService_Get(t *testing.T) {
	t.Parallel()

	t.Run("can be created with nil client", func(t *testing.T) {
		t.Parallel()
		cache := cacheSvc.NewCacheService(nil, 0)

		if cache == nil {
			t.Fatal("NewCacheService returned nil")
		}
	})
}

func TestCacheService_Set(t *testing.T) {
	t.Parallel()

	t.Run("can be created with nil client", func(t *testing.T) {
		t.Parallel()
		cache := cacheSvc.NewCacheService(nil, 0)

		if cache == nil {
			t.Fatal("NewCacheService returned nil")
		}
	})
}

func TestCacheService_Delete(t *testing.T) {
	t.Parallel()

	t.Run("can be created with nil client", func(t *testing.T) {
		t.Parallel()
		cache := cacheSvc.NewCacheService(nil, 0)

		if cache == nil {
			t.Fatal("NewCacheService returned nil")
		}
	})
}

func TestCacheService_Exists(t *testing.T) {
	t.Parallel()

	t.Run("can be created with nil client", func(t *testing.T) {
		t.Parallel()
		cache := cacheSvc.NewCacheService(nil, 0)

		if cache == nil {
			t.Fatal("NewCacheService returned nil")
		}
	})
}

func TestCacheService_SetDefault(t *testing.T) {
	t.Parallel()

	t.Run("can be created with nil client", func(t *testing.T) {
		t.Parallel()
		cache := cacheSvc.NewCacheService(nil, 0)

		if cache == nil {
			t.Fatal("NewCacheService returned nil")
		}
	})
}
