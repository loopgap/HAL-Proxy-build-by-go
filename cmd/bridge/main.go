package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"

	"bridgeos/internal/core"
	"bridgeos/internal/domain"
	apperrors "bridgeos/internal/errors"
	"bridgeos/internal/store"
	"bridgeos/internal/version"
)

var validActorRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

func isValidActor(actor string) bool {
	return validActorRegex.MatchString(actor)
}

func main() {
	os.Exit(run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr))
}

func run(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	ctx := context.Background()
	repo, err := store.NewSQLiteRepository(dbPath(getenv))
	if err != nil {
		return exitOnErr(stderr, err)
	}
	defer repo.Close()

	svc := core.NewService(repo, "artifacts")
	if err := svc.Init(ctx); err != nil {
		return exitOnErr(stderr, err)
	}
	if len(args) == 0 {
		return fatalf(stderr, "usage: bridge <case|approval|report|device|session|version> ...")
	}

	switch args[0] {
	case "case":
		return handleCase(ctx, svc, stdout, stderr, args[1:])
	case "approval":
		return handleApproval(ctx, svc, stdout, stderr, args[1:])
	case "report":
		return handleReport(ctx, svc, stdout, stderr, args[1:])
	case "device":
		return handleDevice(ctx, svc, stdout, stderr, args[1:])
	case "session":
		return handleSession(ctx, svc, stdout, stderr, args[1:])
	case "version":
		return writeJSON(stdout, stderr, map[string]any{
			"name":       version.AppName,
			"version":    version.Version,
			"commit":     version.Commit,
			"build_date": version.BuildDate,
		})
	default:
		return fatalf(stderr, "unknown command %q", args[0])
	}
}

func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

func handleCase(ctx context.Context, svc *core.Service, stdout, stderr io.Writer, args []string) int {
	if len(args) == 0 {
		return fatalf(stderr, "usage: bridge case <new|run|show|events>")
	}

	switch args[0] {
	case "new":
		fs := newFlagSet("case new", stderr)
		specPath := fs.String("spec", "", "path to case spec json")
		actor := fs.String("actor", "cli", "actor name")
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		if *specPath == "" {
			return fatalf(stderr, "--spec is required")
		}
		if !isValidActor(*actor) {
			return fatalf(stderr, "invalid actor name %q: must be 1-64 chars, alphanumeric with _-", *actor)
		}
		var spec domain.CaseSpec
		raw, err := os.ReadFile(*specPath)
		if err != nil {
			return exitOnErr(stderr, err)
		}
		if err := json.Unmarshal(raw, &spec); err != nil {
			return exitOnErr(stderr, err)
		}
		c, err := svc.CreateCase(ctx, spec, *actor)
		if err != nil {
			return exitOnErr(stderr, err)
		}
		return writeJSON(stdout, stderr, c)
	case "run":
		fs := newFlagSet("case run", stderr)
		id := fs.String("id", "", "case id")
		actor := fs.String("actor", "cli", "actor name")
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		if *id == "" {
			return fatalf(stderr, "--id is required")
		}
		if !isValidActor(*actor) {
			return fatalf(stderr, "invalid actor name %q: must be 1-64 chars, alphanumeric with _-", *actor)
		}
		result, err := svc.RunCase(ctx, *id, *actor)
		if err != nil {
			return exitOnErr(stderr, err)
		}
		return writeJSON(stdout, stderr, result)
	case "show":
		fs := newFlagSet("case show", stderr)
		id := fs.String("id", "", "case id")
		actor := fs.String("actor", "cli", "actor name")
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		if *id == "" {
			return fatalf(stderr, "--id is required")
		}
		if !isValidActor(*actor) {
			return fatalf(stderr, "invalid actor name %q: must be 1-64 chars, alphanumeric with _-", *actor)
		}
		c, err := svc.GetCase(ctx, *id, *actor)
		if err != nil {
			return exitOnErr(stderr, err)
		}
		return writeJSON(stdout, stderr, c)
	case "events":
		fs := newFlagSet("case events", stderr)
		id := fs.String("id", "", "case id")
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		if *id == "" {
			return fatalf(stderr, "--id is required")
		}
		events, err := svc.ListEvents(ctx, *id)
		if err != nil {
			return exitOnErr(stderr, err)
		}
		return writeJSON(stdout, stderr, events)
	default:
		return fatalf(stderr, "unknown case command %q", args[0])
	}
}

func handleApproval(ctx context.Context, svc *core.Service, stdout, stderr io.Writer, args []string) int {
	// NOTE: Approving/rejecting approvals requires the user to have
	// either "admin" or "approver" role in their JWT claims.
	// Users without these roles will receive a 403 Forbidden response.
	if len(args) == 0 {
		return fatalf(stderr, "usage: bridge approval <ls|approve|reject>")
	}

	switch args[0] {
	case "ls":
		fs := newFlagSet("approval ls", stderr)
		caseID := fs.String("case-id", "", "optional case id")
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		approvals, err := svc.ListApprovals(ctx, *caseID)
		if err != nil {
			return exitOnErr(stderr, err)
		}
		return writeJSON(stdout, stderr, approvals)
	case "approve", "reject":
		fs := newFlagSet("approval "+args[0], stderr)
		id := fs.String("id", "", "approval id")
		actor := fs.String("actor", "cli", "actor name")
		reason := fs.String("reason", "", "optional reason")
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		if *id == "" {
			return fatalf(stderr, "--id is required")
		}
		if !isValidActor(*actor) {
			return fatalf(stderr, "invalid actor name %q: must be 1-64 chars, alphanumeric with _-", *actor)
		}
		approval, err := svc.ResolveApproval(ctx, *id, *actor, args[0], *reason)
		if err != nil {
			return exitOnErr(stderr, err)
		}
		return writeJSON(stdout, stderr, approval)
	default:
		return fatalf(stderr, "unknown approval command %q", args[0])
	}
}

func handleReport(ctx context.Context, svc *core.Service, stdout, stderr io.Writer, args []string) int {
	if len(args) == 0 || args[0] != "build" {
		return fatalf(stderr, "usage: bridge report build --id <case-id>")
	}
	fs := newFlagSet("report build", stderr)
	id := fs.String("id", "", "case id")
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}
	if *id == "" {
		return fatalf(stderr, "--id is required")
	}
	report, err := svc.BuildReport(ctx, *id, "")
	if err != nil {
		return exitOnErr(stderr, err)
	}
	return writeJSON(stdout, stderr, report)
}

func handleDevice(ctx context.Context, svc *core.Service, stdout, stderr io.Writer, args []string) int {
	if len(args) == 0 || args[0] != "ls" {
		return fatalf(stderr, "usage: bridge device ls")
	}
	devices, err := svc.ListDevices(ctx)
	if err != nil {
		return exitOnErr(stderr, err)
	}
	return writeJSON(stdout, stderr, devices)
}

func handleSession(ctx context.Context, svc *core.Service, stdout, stderr io.Writer, args []string) int {
	if len(args) == 0 || args[0] != "ls" {
		return fatalf(stderr, "usage: bridge session ls")
	}
	sessions, err := svc.ListSessions(ctx)
	if err != nil {
		return exitOnErr(stderr, err)
	}
	return writeJSON(stdout, stderr, sessions)
}

func dbPath(getenv func(string) string) string {
	if env := getenv("BRIDGEOS_DB"); env != "" {
		return env
	}
	if env := getenv("HAL_PROXY_DB"); env != "" {
		return env
	}
	return "bridgeos.db"
}

func writeJSON(stdout, stderr io.Writer, v any) int {
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return exitOnErr(stderr, err)
	}
	return 0
}

func fatalf(stderr io.Writer, format string, args ...any) int {
	_, _ = fmt.Fprintf(stderr, format+"\n", args...)
	return 1
}

func exitOnErr(stderr io.Writer, err error) int {
	if err == nil {
		return 0
	}
	enc := json.NewEncoder(stderr)
	enc.SetIndent("", "  ")
	payload := map[string]any{"error": "internal_server_error", "message": "An unexpected error occurred"}
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		payload["error"] = appErr.Message
		payload["message"] = appErr.Message
		payload["code"] = appErr.Code
	} else if errors.Is(err, store.ErrNotFound) {
		payload["error"] = "resource_not_found"
		payload["message"] = "Resource not found"
	} else if errors.Is(err, store.ErrConcurrentModification) {
		payload["error"] = "concurrent_modification"
		payload["message"] = "Concurrent modification detected"
	}
	_ = enc.Encode(payload)
	return 1
}
