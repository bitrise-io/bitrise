package cmdutil

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRequireArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "view"}

	for _, tc := range []struct {
		name    string
		names   []string
		args    []string
		wantErr string
	}{
		{"no names required", nil, nil, ""},
		{"single missing", []string{"SESSION_ID"}, nil, "missing argument: SESSION_ID"},
		{"one of two missing", []string{"SESSION_ID", "KEY"}, []string{"sess-1"}, "missing argument: KEY"},
		{"two missing", []string{"SESSION_ID", "KEY"}, nil, "missing arguments: SESSION_ID KEY"},
		{"all provided", []string{"SESSION_ID", "KEY"}, []string{"sess-1", "key-1"}, ""},
		{"extra args are fine", []string{"SESSION_ID"}, []string{"sess-1", "extra"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := RequireArgs(tc.names...)(cmd, tc.args)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil || !strings.HasPrefix(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want prefix %q", err, tc.wantErr)
			}
		})
	}
}

func TestDelegateToList_ForwardsArgsAndContext(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "resolved-config")

	var gotArgs []string
	var gotCtx context.Context
	list := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			gotArgs = args
			gotCtx = cmd.Context()
			return nil
		},
	}
	parent := &cobra.Command{Use: "session"}
	parent.AddCommand(list)
	parent.SetContext(ctx)

	if err := DelegateToList(parent, []string{"a", "b"}); err != nil {
		t.Fatalf("DelegateToList: %v", err)
	}
	if len(gotArgs) != 2 || gotArgs[0] != "a" || gotArgs[1] != "b" {
		t.Errorf("args not forwarded, got %v", gotArgs)
	}
	if gotCtx == nil || gotCtx.Value(ctxKey{}) != "resolved-config" {
		t.Errorf("context not propagated to list subcommand")
	}
}

func TestDelegateToList_NoListSubcommandFallsBackToHelp(t *testing.T) {
	parent := &cobra.Command{Use: "session"}
	if err := DelegateToList(parent, nil); err != nil {
		t.Errorf("expected no error falling back to Help, got %v", err)
	}
}
