package cmdutil

import (
	"context"
	"io"
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

func TestDelegateToList_MergesParentPersistentFlags(t *testing.T) {
	var got string
	list := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			got, _ = cmd.Flags().GetString("workspace")
			return nil
		},
	}
	parent := &cobra.Command{Use: "stack", Args: cobra.NoArgs, RunE: DelegateToList}
	parent.AddCommand(list)
	root := &cobra.Command{Use: "rde"}
	root.PersistentFlags().String("workspace", "", "")
	root.AddCommand(parent)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	root.SetArgs([]string{"stack", "--workspace", "ws-2"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got != "ws-2" {
		t.Errorf("workspace = %q, want %q", got, "ws-2")
	}
}

func TestDelegateToList_ValidatesRequiredFlags(t *testing.T) {
	ran := false
	list := &cobra.Command{
		Use: "list",
		RunE: func(*cobra.Command, []string) error {
			ran = true
			return nil
		},
	}
	list.Flags().String("stack", "", "")
	_ = list.MarkFlagRequired("stack")
	parent := &cobra.Command{Use: "machine-type"}
	parent.AddCommand(list)

	err := DelegateToList(parent, nil)
	if err == nil || !strings.Contains(err.Error(), `required flag(s) "stack" not set`) {
		t.Fatalf("error = %v, want a missing required flag error", err)
	}
	if ran {
		t.Error("list body ran despite the missing required flag")
	}
}
