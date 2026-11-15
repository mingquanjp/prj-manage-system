package organization

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
)

func TestCatalogName(t *testing.T) {
	for _, tc := range []struct {
		name, input, want string
		invalid           bool
	}{
		{"trim Unicode", "\u00a0\tTên danh mục\u3000", "Tên danh mục", false},
		{"100 code point", strings.Repeat("ế", 100), strings.Repeat("ế", 100), false},
		{"101 code point", strings.Repeat("ế", 101), "", true},
		{"padding không tính độ dài", " " + strings.Repeat("ế", 100) + " ", strings.Repeat("ế", 100), false},
		{"trắng", "\u0085\u00a0", "", true}, {"UTF8 lỗi", string([]byte{255}), "", true},
		{"NUL không hợp lệ PostgreSQL", "a\x00b", "", true},
		{"không NFC", "e\u0301", "e\u0301", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeCatalogName(tc.input)
			if errors.Is(err, ErrValidation) != tc.invalid || got != tc.want {
				t.Fatal("sai normalize/validation")
			}
		})
	}
}

type catalogFake struct {
	query   ListCatalogQuery
	command CatalogCommand
	calls   int
	err     error
}

func (f *catalogFake) ListCatalog(_ context.Context, q ListCatalogQuery) (CatalogPage, error) {
	f.query = q
	f.calls++
	return CatalogPage{}, f.err
}
func (f *catalogFake) CreateCatalog(_ context.Context, c CatalogCommand) (CatalogItem, error) {
	f.command = c
	f.calls++
	return CatalogItem{}, f.err
}
func (f *catalogFake) RenameCatalog(_ context.Context, c CatalogCommand) (CatalogItem, error) {
	return f.CreateCatalog(context.Background(), c)
}
func (f *catalogFake) DeleteCatalog(_ context.Context, c CatalogCommand) error {
	_, err := f.CreateCatalog(context.Background(), c)
	return err
}
func TestCatalogServiceBoundary(t *testing.T) {
	ctx := context.Background()
	f := &catalogFake{}
	s := NewCatalogService(f)
	_, err := s.List(ctx, ListCatalogQuery{ActorID: 7, Kind: Departments, Query: " %_ "})
	if err != nil || f.query.Page != 1 || f.query.PageSize != 20 || f.query.Query != "%_" {
		t.Fatal("default/filter sai")
	}
	_, err = s.List(ctx, ListCatalogQuery{ActorID: 7, Kind: Departments, Page: math.MaxInt32, PageSize: 100})
	if err != nil || f.query.Offset() != 214748364600 {
		t.Fatal("offset phải int64")
	}
	for _, q := range []ListCatalogQuery{{ActorID: 7, Kind: "users"}, {ActorID: 7, Kind: Departments, Page: -1}, {ActorID: 7, Kind: Departments, PageSize: 101}, {ActorID: 7, Kind: Departments, Page: math.MaxInt32 + 1}, {ActorID: 7, Kind: Departments, Query: strings.Repeat("a", 101)}} {
		calls := f.calls
		_, err = s.List(ctx, q)
		if !errors.Is(err, ErrValidation) || f.calls != calls {
			t.Fatal("đầu vào sai không được gọi repository")
		}
	}
	c := CatalogCommand{ActorID: 7, Kind: Departments, ID: 1, Name: " New ", ExpectedBeforeName: " Old "}
	_, err = s.Rename(ctx, c)
	if err != nil || f.command.Name != "New" || f.command.ExpectedBeforeName != " Old " {
		t.Fatal("tên mới trim nhưng tên nền giữ nguyên")
	}
	for _, e := range []error{ErrUnauthenticated, ErrForbidden, ErrStale, ErrDuplicateName, ErrCatalogInUse, ErrPersistence, ErrUnavailable, ErrCommitUnknown} {
		f.err = e
		if _, err = s.Create(ctx, c); !errors.Is(err, e) {
			t.Fatal("phải giữ phân loại lỗi port")
		}
	}
	calls := f.calls
	c.ID = 0
	if err = s.Delete(ctx, c); !errors.Is(err, ErrValidation) || f.calls != calls {
		t.Fatal("ID sai")
	}
}
