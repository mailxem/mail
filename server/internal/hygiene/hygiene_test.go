package hygiene

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// fakeResolver answers from maps; a missing name is NXDOMAIN, a name in fail times out.
type fakeResolver struct {
	mx    map[string][]*net.MX
	ip    map[string]bool
	fail  map[string]bool
	calls int
}

func (r *fakeResolver) LookupMX(_ context.Context, name string) ([]*net.MX, error) {
	r.calls++
	if r.fail[name] {
		return nil, &net.DNSError{Err: "i/o timeout", Name: name, IsTimeout: true}
	}
	if mx, ok := r.mx[name]; ok {
		return mx, nil
	}
	return nil, &net.DNSError{Err: "no such host", Name: name, IsNotFound: true}
}

func (r *fakeResolver) LookupIPAddr(_ context.Context, name string) ([]net.IPAddr, error) {
	if r.fail[name] {
		return nil, errors.New("server misbehaving")
	}
	if r.ip[name] {
		return []net.IPAddr{{IP: net.IPv4(192, 0, 2, 1)}}, nil
	}
	return nil, &net.DNSError{Err: "no such host", Name: name, IsNotFound: true}
}

func TestReason(t *testing.T) {
	r := &fakeResolver{
		mx:   map[string][]*net.MX{"good.org": {{Host: "mx.good.org."}}, "nullmx.org": {{Host: "."}}},
		ip:   map[string]bool{"aonly.org": true},
		fail: map[string]bool{"flaky.org": true},
	}
	imports := Checker{Resolver: r, Import: true}
	form := Checker{Resolver: r}
	ctx := context.Background()
	for email, want := range map[string]string{
		"jane@good.org":         "",
		"Jane@Good.org ":        "",
		"not-an-email":          Invalid,
		"jane@nodot":            Invalid,
		"Jane <jane@good.org>":  Invalid,
		"jane@gmial.com":        "did you mean gmail.com",
		"jane@gmail.co":         "did you mean gmail.com",
		"jane@hotmial.com":      "did you mean hotmail.com",
		"jane@yaho.com":         "did you mean yahoo.com",
		"jane@outlok.com":       "did you mean outlook.com",
		"jane@mailinator.com":   Disposable,
		"jane@yopmail.com":      Disposable,
		"info@good.org":         Role,
		"no-reply@good.org":     Role,
		"postmaster+x@good.org": Role,
		"jane@gone.org":         NoMail, // no MX, no A/AAAA
		"jane@nullmx.org":       NoMail,
		"jane@aonly.org":        "", // A record fallback
		"jane@flaky.org":        "", // DNS errors never skip
		"jane@example.com":      "", // reserved names are not looked up
		"jane@sub.example.test": "",
	} {
		require.Equal(t, want, imports.Reason(ctx, email), email)
	}
	require.Equal(t, "", form.Reason(ctx, "info@good.org"), "role addresses are allowed on form signups")

	calls := r.calls
	imports.Reason(ctx, "someone@good.org")
	imports.Reason(ctx, "other@gone.org")
	require.Equal(t, calls, r.calls, "lookups are cached per checker")
}

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	var dialect gorm.Dialector = sqlite.Open("file:" + uuid.NewString() + "?mode=memory&cache=shared")
	if dsn := os.Getenv("POSTHOOT_TEST_DATABASE_URL"); dsn != "" {
		admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		require.NoError(t, err)
		schema := "hygiene_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
		t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); raw, _ := admin.DB(); raw.Close() })
		dialect = postgres.Open(dsn + " search_path=" + schema)
	}
	db, err := gorm.Open(dialect, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&NeverEmailEntry{}))
	raw, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { raw.Close() })
	return db
}

func TestNeverEmailList(t *testing.T) {
	db := testDB(t)
	t.Setenv("SUPPRESSION_HASH_KEY", "test-key")
	now := time.Now()
	require.NoError(t, Block(db, "Gone@Example.net", now))
	require.NoError(t, Block(db, "gone@example.net", now)) // upsert, not a duplicate

	var rows []NeverEmailEntry
	require.NoError(t, db.Find(&rows).Error)
	require.Len(t, rows, 1)
	require.NotContains(t, rows[0].Hash, "example", "the plain address is never stored")
	require.Len(t, rows[0].Hash, 64)

	got, err := Blocked(db, "gone@example.net", "GONE@example.net", "kept@example.net")
	require.NoError(t, err)
	require.Equal(t, map[string]bool{"gone@example.net": true, "GONE@example.net": true}, got)

	// An entry stops applying after its expiry.
	require.NoError(t, Block(db, "old@example.net", now.Add(-NeverEmailFor-time.Hour)))
	got, err = Blocked(db, "old@example.net")
	require.NoError(t, err)
	require.Empty(t, got)

	// A different key cannot match existing hashes.
	t.Setenv("SUPPRESSION_HASH_KEY", "other-key")
	got, err = Blocked(db, "gone@example.net")
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestNeverEmailListOffWithoutKey(t *testing.T) {
	db := testDB(t)
	t.Setenv("SUPPRESSION_HASH_KEY", "")
	require.NoError(t, Block(db, "gone@example.net", time.Now()))
	var n int64
	require.NoError(t, db.Model(&NeverEmailEntry{}).Count(&n).Error)
	require.Zero(t, n)

	t.Setenv("SUPPRESSION_HASH_KEY", "test-key")
	require.NoError(t, Block(db, "gone@example.net", time.Now()))
	t.Setenv("SUPPRESSION_HASH_KEY", "")
	got, err := Blocked(db, "gone@example.net")
	require.NoError(t, err)
	require.Empty(t, got)
}
