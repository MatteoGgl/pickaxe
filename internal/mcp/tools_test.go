package mcp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	internalmcp "github.com/matteoggl/pickaxe/internal/mcp"
	"github.com/matteoggl/pickaxe/internal/vault"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type fakeVault struct {
	files      []vault.ResolvedFile
	listErr    error
	contents   map[string]string
	readErr    error
	writeErr   error
	written    map[string]string
	replaceErr error
	replaced   []fakeReplace
}

type fakeReplace struct {
	Name, OldStr, NewStr string
}

func (f *fakeVault) ListFiles() ([]vault.ResolvedFile, error) {
	return f.files, f.listErr
}

func (f *fakeVault) ReadFile(name string) (string, error) {
	if f.readErr != nil {
		return "", f.readErr
	}
	c, ok := f.contents[name]
	if !ok {
		return "", vault.ErrNotFound
	}
	return c, nil
}

func (f *fakeVault) WriteFile(name, content string) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	if f.written == nil {
		f.written = make(map[string]string)
	}
	f.written[name] = content
	return nil
}

func (f *fakeVault) ReplaceInFile(name, oldStr, newStr string) error {
	if f.replaceErr != nil {
		return f.replaceErr
	}
	f.replaced = append(f.replaced, fakeReplace{name, oldStr, newStr})
	return nil
}

func listHandler(fv *fakeVault) func(context.Context, *sdkmcp.CallToolRequest, internalmcp.ListVaultFilesParams) (*sdkmcp.CallToolResult, any, error) {
	return internalmcp.MakeListVaultFilesHandler(fv)
}

func updateHandler(fv *fakeVault, tracker *internalmcp.ReadTracker) func(context.Context, *sdkmcp.CallToolRequest, internalmcp.UpdateVaultFileParams) (*sdkmcp.CallToolResult, any, error) {
	return internalmcp.MakeUpdateVaultFileHandler(fv, tracker)
}

func readHandler(fv *fakeVault, tracker *internalmcp.ReadTracker) func(context.Context, *sdkmcp.CallToolRequest, internalmcp.ReadVaultFileParams) (*sdkmcp.CallToolResult, any, error) {
	return internalmcp.MakeReadVaultFileHandler(fv, tracker)
}

func TestListVaultFiles_Empty(t *testing.T) {
	fv := &fakeVault{}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected non-error result")
	}
	var items []any
	json.Unmarshal([]byte(result.Content[0].(*sdkmcp.TextContent).Text), &items)
	if len(items) != 0 {
		t.Errorf("expected empty array, got %d items", len(items))
	}
}

func TestListVaultFiles_WithFile(t *testing.T) {
	fv := &fakeVault{files: []vault.ResolvedFile{{Name: "adrs", LastMod: time.Now()}}}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("unexpected error result")
	}
	var items []map[string]any
	json.Unmarshal([]byte(result.Content[0].(*sdkmcp.TextContent).Text), &items)
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestListVaultFiles_UnavailableFile(t *testing.T) {
	fv := &fakeVault{files: []vault.ResolvedFile{{Name: "ghost", Unavailable: true}}}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected non-error result even for unavailable file")
	}
}

func TestListVaultFiles_NoRegistry(t *testing.T) {
	fv := &fakeVault{listErr: vault.ErrNotInitialized}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("missing registry should return a message, not an error")
	}
}

func TestReadVaultFile_Success(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"adrs": "# ADRs\nDecision 1"}}
	result, _, err := readHandler(fv, internalmcp.NewReadTracker())(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "adrs"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	want := "     1\t# ADRs\n     2\tDecision 1"
	if text != want {
		t.Errorf("unexpected content:\n got  %q\n want %q", text, want)
	}
}

func TestReadVaultFile_NotRegistered(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{}}
	result, _, err := readHandler(fv, internalmcp.NewReadTracker())(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unregistered name")
	}
}

func TestReadVaultFile_UnavailableFile(t *testing.T) {
	fv := &fakeVault{readErr: vault.ErrNotFound}
	result, _, err := readHandler(fv, internalmcp.NewReadTracker())(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "ghost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unavailable file")
	}
}

func TestReadVaultFile_NoRegistry(t *testing.T) {
	fv := &fakeVault{readErr: vault.ErrNotInitialized}
	result, _, err := readHandler(fv, internalmcp.NewReadTracker())(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "anything"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true with no registry")
	}
}

func TestListVaultFiles_NoPathInResponse(t *testing.T) {
	fv := &fakeVault{files: []vault.ResolvedFile{{Name: "note", LastMod: time.Now()}}}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var items []map[string]any
	json.Unmarshal([]byte(result.Content[0].(*sdkmcp.TextContent).Text), &items)
	for _, item := range items {
		if _, ok := item["path"]; ok {
			t.Errorf("path must not appear in list_vault_files response, got: %v", item)
		}
	}
}

func TestListVaultFiles_UnavailableNoPath(t *testing.T) {
	fv := &fakeVault{files: []vault.ResolvedFile{{Name: "ghost", Unavailable: true}}}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var items []map[string]any
	json.Unmarshal([]byte(result.Content[0].(*sdkmcp.TextContent).Text), &items)
	for _, item := range items {
		if _, ok := item["path"]; ok {
			t.Errorf("path must not appear for unavailable files, got: %v", item)
		}
	}
}

func TestReadVaultFile_UnavailableErrorNoPath(t *testing.T) {
	// vault.ReadFile returns "file %q is currently unavailable" (name, not path)
	fv := &fakeVault{readErr: fmt.Errorf("file %q is currently unavailable", "secret")}
	result, _, err := readHandler(fv, internalmcp.NewReadTracker())(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	if strings.Contains(result.Content[0].(*sdkmcp.TextContent).Text, "/nonexistent/secret") {
		t.Errorf("error message must not contain file path")
	}
}

func TestReadVaultFile_ReadErrorNoPath(t *testing.T) {
	// vault.ReadFile returns "could not read file %q" (name, not path)
	fv := &fakeVault{readErr: fmt.Errorf("could not read file %q", "locked")}
	result, _, err := readHandler(fv, internalmcp.NewReadTracker())(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "locked"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for unreadable file")
	}
	if strings.Contains(result.Content[0].(*sdkmcp.TextContent).Text, "/tmp/") {
		t.Errorf("error message must not contain file path")
	}
}

func TestReadVaultFile_StripsFrontmatter(t *testing.T) {
	// vault.ReadFile strips frontmatter before returning; handler formats with line numbers
	fv := &fakeVault{contents: map[string]string{"note": "\n# Hello"}}
	result, _, err := readHandler(fv, internalmcp.NewReadTracker())(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "note"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "# Hello") {
		t.Errorf("body missing: got %q", text)
	}
}

func TestListVaultFiles_IncludesWritableFlag(t *testing.T) {
	fv := &fakeVault{files: []vault.ResolvedFile{
		{Name: "rw", LastMod: time.Now(), Writable: true},
		{Name: "ro", LastMod: time.Now(), Writable: false},
	}}
	result, _, err := listHandler(fv)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ListVaultFilesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("unexpected error result")
	}
	var items []map[string]any
	if err := json.Unmarshal([]byte(result.Content[0].(*sdkmcp.TextContent).Text), &items); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	byName := map[string]map[string]any{}
	for _, item := range items {
		byName[item["name"].(string)] = item
	}
	if v, ok := byName["rw"]["writable"]; !ok || v != true {
		t.Errorf("rw: expected writable=true, got %v (present=%v)", v, ok)
	}
	if v, ok := byName["ro"]["writable"]; !ok || v != false {
		t.Errorf("ro: expected writable=false (present, not omitted), got %v (present=%v)", v, ok)
	}
}

func TestUpdateVaultFile_Success(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"note": "old content"}}
	tracker := internalmcp.NewReadTracker()
	readHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "note"})

	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{Name: "note", Content: "new content"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", result.Content[0].(*sdkmcp.TextContent).Text)
	}
	if fv.written["note"] != "new content" {
		t.Errorf("expected written content, got %q", fv.written["note"])
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "note") {
		t.Errorf("success message missing name: %q", text)
	}
}

func TestUpdateVaultFile_EmptyName(t *testing.T) {
	fv := &fakeVault{}
	result, _, err := updateHandler(fv, internalmcp.NewReadTracker())(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{Name: "", Content: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for empty name")
	}
}

func TestUpdateVaultFile_ReadOnly(t *testing.T) {
	fv := &fakeVault{writeErr: vault.ErrReadOnly}
	tracker := internalmcp.NewReadTracker()
	tracker.MarkRead("note")
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{Name: "note", Content: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for read-only file")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "pickaxe unlock") {
		t.Errorf("expected unlock hint, got: %q", text)
	}
}

func TestUpdateVaultFile_NotFound(t *testing.T) {
	fv := &fakeVault{writeErr: vault.ErrNotFound}
	tracker := internalmcp.NewReadTracker()
	tracker.MarkRead("ghost")
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{Name: "ghost", Content: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for not-found file")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "list_vault_files") {
		t.Errorf("expected list hint, got: %q", text)
	}
}

func TestUpdateVaultFile_RequiresRead(t *testing.T) {
	fv := &fakeVault{}
	tracker := internalmcp.NewReadTracker()
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{Name: "note", Content: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true when file has not been read")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "read_vault_file") {
		t.Errorf("expected read hint, got: %q", text)
	}
}

func TestUpdateVaultFile_SucceedsAfterRead(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"note": "original"}}
	tracker := internalmcp.NewReadTracker()

	readResult, _, err := readHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "note"})
	if err != nil || readResult.IsError {
		t.Fatalf("read failed: err=%v isError=%v", err, readResult.IsError)
	}

	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{Name: "note", Content: "updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success after read, got: %s", result.Content[0].(*sdkmcp.TextContent).Text)
	}
	if fv.written["note"] != "updated" {
		t.Errorf("expected written content %q, got %q", "updated", fv.written["note"])
	}
}

func TestUpdateVaultFile_ReadOneUpdateAnother(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"fileA": "content A"}}
	tracker := internalmcp.NewReadTracker()

	// Read fileA, then try to update fileB
	readHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "fileA"})

	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{Name: "fileB", Content: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true: reading fileA should not allow updating fileB")
	}
}

func TestUpdateVaultFile_WriteInvalidatesRead(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"note": "original"}}
	tracker := internalmcp.NewReadTracker()

	// Read then update (succeeds)
	readHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "note"})
	first, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{Name: "note", Content: "v1"})
	if err != nil || first.IsError {
		t.Fatalf("first update failed: err=%v isError=%v", err, first.IsError)
	}

	// Second update without re-reading should fail
	second, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{Name: "note", Content: "v2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !second.IsError {
		t.Fatal("expected IsError=true: write should invalidate read token")
	}
}


func TestFormatLines_FullFile(t *testing.T) {
	got := internalmcp.FormatLines("alpha\nbeta\ngamma", 0, 0)
	want := "     1\talpha\n     2\tbeta\n     3\tgamma"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatLines_WithOffset(t *testing.T) {
	got := internalmcp.FormatLines("a\nb\nc\nd\ne", 3, 0)
	want := "     3\tc\n     4\td\n     5\te"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatLines_WithLimit(t *testing.T) {
	got := internalmcp.FormatLines("a\nb\nc\nd\ne", 0, 2)
	want := "     1\ta\n     2\tb"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatLines_OffsetAndLimit(t *testing.T) {
	got := internalmcp.FormatLines("a\nb\nc\nd\ne", 2, 2)
	want := "     2\tb\n     3\tc"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatLines_OffsetBeyondEnd(t *testing.T) {
	got := internalmcp.FormatLines("a\nb", 10, 0)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestFormatLines_EmptyContent(t *testing.T) {
	got := internalmcp.FormatLines("", 0, 0)
	want := "     1\t"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatLines_SingleLine(t *testing.T) {
	got := internalmcp.FormatLines("hello", 0, 0)
	want := "     1\thello"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}


func TestReadVaultFile_WithOffset(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"note": "line1\nline2\nline3"}}
	result, _, err := readHandler(fv, internalmcp.NewReadTracker())(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "note", Offset: 2})
	if err != nil || result.IsError {
		t.Fatalf("err=%v isError=%v", err, result.IsError)
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	want := "     2\tline2\n     3\tline3"
	if text != want {
		t.Errorf("got %q, want %q", text, want)
	}
}

func TestReadVaultFile_WithLimit(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"note": "line1\nline2\nline3"}}
	result, _, err := readHandler(fv, internalmcp.NewReadTracker())(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "note", Limit: 1})
	if err != nil || result.IsError {
		t.Fatalf("err=%v isError=%v", err, result.IsError)
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	want := "     1\tline1"
	if text != want {
		t.Errorf("got %q, want %q", text, want)
	}
}

func TestReadVaultFile_OffsetAndLimit(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"note": "a\nb\nc\nd\ne"}}
	result, _, err := readHandler(fv, internalmcp.NewReadTracker())(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "note", Offset: 2, Limit: 2})
	if err != nil || result.IsError {
		t.Fatalf("err=%v isError=%v", err, result.IsError)
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	want := "     2\tb\n     3\tc"
	if text != want {
		t.Errorf("got %q, want %q", text, want)
	}
}

func TestReadVaultFile_OffsetBeyondEnd(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"note": "a\nb"}}
	tracker := internalmcp.NewReadTracker()
	result, _, err := readHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "note", Offset: 100})
	if err != nil || result.IsError {
		t.Fatalf("err=%v isError=%v", err, result.IsError)
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if text != "" {
		t.Errorf("expected empty, got %q", text)
	}
	if !tracker.HasRead("note") {
		t.Error("partial read should still mark file as read")
	}
}

func TestReadVaultFile_LineRangeReadEnablesUpdate(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"note": "a\nb\nc"}}
	tracker := internalmcp.NewReadTracker()

	// Partial read
	readHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "note", Offset: 2, Limit: 1})

	// Should still allow update
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{Name: "note", Content: "new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("partial read should satisfy ReadTracker: %s", result.Content[0].(*sdkmcp.TextContent).Text)
	}
}


func TestUpdateVaultFile_StringReplace_Success(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"note": "hello world"}}
	tracker := internalmcp.NewReadTracker()
	readHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "note"})

	old, new := "hello", "goodbye"
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{
		Name: "note", OldString: &old, NewString: &new,
	})
	if err != nil || result.IsError {
		t.Fatalf("err=%v isError=%v", err, result.IsError)
	}
	if len(fv.replaced) != 1 || fv.replaced[0].OldStr != "hello" || fv.replaced[0].NewStr != "goodbye" {
		t.Errorf("unexpected replace calls: %+v", fv.replaced)
	}
}

func TestUpdateVaultFile_StringReplace_RequiresRead(t *testing.T) {
	fv := &fakeVault{}
	tracker := internalmcp.NewReadTracker()
	old, new := "a", "b"
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{
		Name: "note", OldString: &old, NewString: &new,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true when file has not been read")
	}
}

func TestUpdateVaultFile_StringReplace_InvalidatesRead(t *testing.T) {
	fv := &fakeVault{contents: map[string]string{"note": "x"}}
	tracker := internalmcp.NewReadTracker()
	readHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.ReadVaultFileParams{Name: "note"})

	old, new := "x", "y"
	updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{
		Name: "note", OldString: &old, NewString: &new,
	})

	// Second update without re-reading should fail
	result, _, _ := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{
		Name: "note", OldString: &old, NewString: &new,
	})
	if !result.IsError {
		t.Fatal("expected IsError=true: replace should invalidate read")
	}
}

func TestUpdateVaultFile_MutualExclusion(t *testing.T) {
	fv := &fakeVault{}
	tracker := internalmcp.NewReadTracker()
	tracker.MarkRead("note")
	old := "x"
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{
		Name: "note", Content: "full", OldString: &old,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for mutual exclusion")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "cannot provide both") {
		t.Errorf("unexpected message: %q", text)
	}
}

func TestUpdateVaultFile_NoParams(t *testing.T) {
	fv := &fakeVault{}
	tracker := internalmcp.NewReadTracker()
	tracker.MarkRead("note")
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{
		Name: "note",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for no params")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "provide") {
		t.Errorf("unexpected message: %q", text)
	}
}

func TestUpdateVaultFile_OldStringOnly(t *testing.T) {
	fv := &fakeVault{}
	tracker := internalmcp.NewReadTracker()
	tracker.MarkRead("note")
	old := "x"
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{
		Name: "note", OldString: &old,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for old_string without new_string")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "both") {
		t.Errorf("unexpected message: %q", text)
	}
}

func TestUpdateVaultFile_EmptyOldString(t *testing.T) {
	fv := &fakeVault{}
	tracker := internalmcp.NewReadTracker()
	tracker.MarkRead("note")
	old, new := "", "y"
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{
		Name: "note", OldString: &old, NewString: &new,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for empty old_string")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "empty") {
		t.Errorf("unexpected message: %q", text)
	}
}

func TestUpdateVaultFile_StringReplace_NoMatch(t *testing.T) {
	fv := &fakeVault{replaceErr: vault.ErrNoMatch}
	tracker := internalmcp.NewReadTracker()
	tracker.MarkRead("note")
	old, new := "missing", "x"
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{
		Name: "note", OldString: &old, NewString: &new,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for no match")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "not found") {
		t.Errorf("unexpected message: %q", text)
	}
}

func TestUpdateVaultFile_StringReplace_AmbiguousMatch(t *testing.T) {
	fv := &fakeVault{replaceErr: vault.ErrAmbiguousMatch}
	tracker := internalmcp.NewReadTracker()
	tracker.MarkRead("note")
	old, new := "dup", "x"
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{
		Name: "note", OldString: &old, NewString: &new,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for ambiguous match")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "multiple") {
		t.Errorf("unexpected message: %q", text)
	}
}

func TestUpdateVaultFile_StringReplace_ReadOnly(t *testing.T) {
	fv := &fakeVault{replaceErr: vault.ErrReadOnly}
	tracker := internalmcp.NewReadTracker()
	tracker.MarkRead("note")
	old, new := "x", "y"
	result, _, err := updateHandler(fv, tracker)(context.Background(), &sdkmcp.CallToolRequest{}, internalmcp.UpdateVaultFileParams{
		Name: "note", OldString: &old, NewString: &new,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for read-only")
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, "pickaxe unlock") {
		t.Errorf("expected unlock hint: %q", text)
	}
}
