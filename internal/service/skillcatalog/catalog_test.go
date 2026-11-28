package skillcatalog

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
)

type recordingRepository struct {
	calls   int
	method  string
	command Command
	query   ListQuery
	err     error
}

func (r *recordingRepository) List(_ context.Context, q ListQuery) (Page, error) {
	r.calls++
	r.method = "list"
	r.query = q
	return Page{Page: q.Page, PageSize: q.PageSize, Items: []Item{}}, r.err
}
func (r *recordingRepository) Create(_ context.Context, c Command) (Item, error) {
	r.calls++
	r.method = "create"
	r.command = c
	return Item{ID: 7, Name: c.Name, Category: Category{ID: c.CategoryID}}, r.err
}
func (r *recordingRepository) Update(_ context.Context, c Command) (Item, error) {
	r.calls++
	r.method = "update"
	r.command = c
	return Item{ID: c.ID, Name: c.Name, Category: Category{ID: c.CategoryID}}, r.err
}
func (r *recordingRepository) Delete(_ context.Context, c Command) error {
	r.calls++
	r.method = "delete"
	r.command = c
	return r.err
}

func TestCatalogNameNormalization(t *testing.T) {
	spaces := "\t\n\v\f\r \u0085\u00a0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200a\u2028\u2029\u202f\u205f\u3000"
	for _, name := range []string{"ĐIỆN", strings.Repeat("ế", 100), "e\u0302\u0301", "ế", "\ufeffTên"} {
		r := &recordingRepository{}
		item, err := NewService(r).Create(t.Context(), Command{ActorID: 1, Kind: Categories, Name: spaces + name + spaces})
		if err != nil || item.Name != name || r.command.Name != name || r.calls != 1 {
			t.Fatal("tên phải trim White_Space, giữ hoa/thường, dấu và dạng Unicode gốc")
		}
	}
	for _, name := range []string{"", spaces, strings.Repeat("ế", 101), "Tên\x00", string([]byte{0xff})} {
		r := &recordingRepository{}
		_, err := NewService(r).Create(t.Context(), Command{ActorID: 1, Kind: Categories, Name: name})
		if !errors.Is(err, ErrValidation) || r.calls != 0 {
			t.Fatal("tên sai phải dừng trước repository")
		}
	}
}

func TestCatalogBeforeRemainsRaw(t *testing.T) {
	for _, kind := range []Kind{Categories, Skills} {
		for _, beforeName := range []string{"", " \u3000Tên cũ\u00a0 ", strings.Repeat("ế", 101)} {
			c := Command{ActorID: 1, Kind: kind, ID: 7, Name: " Tên mới ", ExpectedBefore: Before{Name: beforeName}}
			if kind == Skills {
				c.CategoryID = 9
				c.ExpectedBefore.CategoryID = 8
			}
			r := &recordingRepository{}
			s := NewService(r)
			_, err := s.Update(t.Context(), c)
			if err != nil || r.command.ExpectedBefore != c.ExpectedBefore || r.command.Name != "Tên mới" || r.command.CategoryID != c.CategoryID || r.calls != 1 || r.method != "update" {
				t.Fatal("update phải giữ raw before và normalize riêng tên đích")
			}
			if err = s.Delete(t.Context(), c); err != nil || r.command.ExpectedBefore != c.ExpectedBefore || r.calls != 2 || r.method != "delete" {
				t.Fatal("delete phải dùng đúng snapshot before mà không normalize/cắt")
			}
		}
	}
}

func TestCatalogNoOpStillReachesCurrentPermissionCheck(t *testing.T) {
	r := &recordingRepository{err: ErrForbidden}
	c := Command{ActorID: math.MaxInt32, Kind: Skills, ID: math.MaxInt32, Name: " Tên ", CategoryID: math.MaxInt32, ExpectedBefore: Before{Name: "Tên", CategoryID: math.MaxInt32}}
	_, err := NewService(r).Update(t.Context(), c)
	if !errors.Is(err, ErrForbidden) || r.calls != 1 || r.command.Name != "Tên" || r.command.ExpectedBefore != c.ExpectedBefore {
		t.Fatal("no-op vẫn phải gọi adapter kiểm quyền và dữ liệu hiện hành")
	}
}

func TestCatalogMutationValidationStopsBeforePort(t *testing.T) {
	valid := Command{ActorID: 1, Kind: Skills, ID: 7, Name: "Tên", CategoryID: 8, ExpectedBefore: Before{Name: "Tên cũ", CategoryID: 8}}
	for _, tc := range []struct {
		name string
		edit func(*Command)
		want error
	}{
		{"actor thiếu", func(c *Command) { c.ActorID = 0 }, ErrUnauthenticated},
		{"actor âm", func(c *Command) { c.ActorID = -1 }, ErrUnauthenticated},
		{"actor tràn", func(c *Command) { c.ActorID = math.MaxInt32 + 1 }, ErrUnauthenticated},
		{"kind sai", func(c *Command) { c.Kind = "departments" }, ErrValidation},
		{"nhóm thiếu", func(c *Command) { c.CategoryID = 0 }, ErrValidation},
		{"nhóm âm", func(c *Command) { c.CategoryID = -1 }, ErrValidation},
		{"nhóm tràn", func(c *Command) { c.CategoryID = math.MaxInt32 + 1 }, ErrValidation},
		{"category có nhóm", func(c *Command) { c.Kind = Categories }, ErrValidation},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := valid
			tc.edit(&c)
			r := &recordingRepository{}
			s := NewService(r)
			if _, err := s.Create(t.Context(), c); !errors.Is(err, tc.want) {
				t.Fatal("create nhận input sai")
			}
			if _, err := s.Update(t.Context(), c); !errors.Is(err, tc.want) {
				t.Fatal("update nhận input sai")
			}
			if r.calls != 0 {
				t.Fatal("validation lỗi vẫn gọi repository")
			}
		})
	}
	for _, edit := range []func(*Command){
		func(c *Command) { c.ID = 0 },
		func(c *Command) { c.ID = -1 },
		func(c *Command) { c.ID = math.MaxInt32 + 1 },
		func(c *Command) { c.ExpectedBefore.Name = "Tên\x00" },
		func(c *Command) { c.ExpectedBefore.Name = string([]byte{0xff}) },
		func(c *Command) { c.ExpectedBefore.CategoryID = 0 },
		func(c *Command) { c.ExpectedBefore.CategoryID = -1 },
		func(c *Command) { c.ExpectedBefore.CategoryID = math.MaxInt32 + 1 },
	} {
		c := valid
		edit(&c)
		r := &recordingRepository{}
		s := NewService(r)
		if _, err := s.Update(t.Context(), c); !errors.Is(err, ErrValidation) {
			t.Fatal("update nhận ID/before sai")
		}
		if err := s.Delete(t.Context(), c); !errors.Is(err, ErrValidation) || r.calls != 0 {
			t.Fatal("delete nhận ID/before sai hoặc gọi port")
		}
	}
}

func TestCatalogListDefaultsLiteralAndOffset(t *testing.T) {
	r := &recordingRepository{}
	s := NewService(r)
	_, err := s.List(t.Context(), ListQuery{ActorID: 1, Kind: Skills, Query: "\u3000%_\\Kỹ năng\u00a0"})
	if err != nil || r.query.Query != "%_\\Kỹ năng" || r.query.Page != 1 || r.query.PageSize != 20 || r.calls != 1 {
		t.Fatal("query chỉ trim, giữ ký tự tìm literal và default pagination")
	}
	q := ListQuery{ActorID: math.MaxInt32, Kind: Skills, Query: strings.Repeat("ế", 100), CategoryID: math.MaxInt32, Page: math.MaxInt32, PageSize: 100}
	_, err = s.List(t.Context(), q)
	if err != nil || r.query != q || r.query.Offset() != 214748364600 || r.calls != 2 {
		t.Fatal("biên ID/query hoặc offset bị thay/tràn")
	}
	_, err = s.List(t.Context(), ListQuery{ActorID: 1, Kind: Categories})
	if err != nil || r.query.CategoryID != 0 || r.calls != 3 {
		t.Fatal("category list không có filter nhóm phải hợp lệ")
	}
}

func TestCatalogListInvalidDoesNotCallPort(t *testing.T) {
	for _, q := range []ListQuery{
		{Kind: Skills},
		{ActorID: -1, Kind: Skills},
		{ActorID: math.MaxInt32 + 1, Kind: Skills},
		{ActorID: 1, Kind: "positions"},
		{ActorID: 1, Kind: Skills, Page: -1},
		{ActorID: 1, Kind: Skills, Page: math.MaxInt32 + 1},
		{ActorID: 1, Kind: Skills, PageSize: -1},
		{ActorID: 1, Kind: Skills, PageSize: 101},
		{ActorID: 1, Kind: Skills, CategoryID: -1},
		{ActorID: 1, Kind: Skills, CategoryID: math.MaxInt32 + 1},
		{ActorID: 1, Kind: Categories, CategoryID: 1},
		{ActorID: 1, Kind: Skills, Query: strings.Repeat("ế", 101)},
		{ActorID: 1, Kind: Skills, Query: "q\x00"},
		{ActorID: 1, Kind: Skills, Query: string([]byte{0xff})},
	} {
		r := &recordingRepository{}
		if _, err := NewService(r).List(t.Context(), q); err == nil || r.calls != 0 {
			t.Fatal("list input sai phải dừng trước port")
		}
	}
}

func TestCatalogPreservesRepositoryErrorsWithoutRetry(t *testing.T) {
	for _, want := range []error{ErrUnauthenticated, ErrPasswordChangeRequired, ErrForbidden, ErrNotFound, ErrStale, ErrDuplicateName, ErrCatalogInUse, ErrUnavailable, ErrPersistence, errors.Join(ErrCommitUnknown, ErrUnavailable)} {
		for _, operation := range []string{"list", "create", "update", "delete"} {
			r := &recordingRepository{err: want}
			s := NewService(r)
			c := Command{ActorID: 1, Kind: Skills, ID: 7, Name: "Tên", CategoryID: 8, ExpectedBefore: Before{Name: "Cũ", CategoryID: 8}}
			var err error
			switch operation {
			case "list":
				_, err = s.List(t.Context(), ListQuery{ActorID: 1, Kind: Skills})
			case "create":
				_, err = s.Create(t.Context(), c)
			case "update":
				_, err = s.Update(t.Context(), c)
			case "delete":
				err = s.Delete(t.Context(), c)
			}
			if err != want || r.calls != 1 || r.method != operation {
				t.Fatal("service phải giữ lỗi adapter và không tự retry")
			}
		}
	}
}
