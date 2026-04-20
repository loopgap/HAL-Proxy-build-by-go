package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"bridgeos/internal/core"
	"bridgeos/internal/domain"
	apperrors "bridgeos/internal/errors"
	"bridgeos/internal/store"
	"bridgeos/internal/version"
)

func main() {
	ctx := context.Background()
	repo, err := store.NewSQLiteRepository(dbPath())
	exitOnErr(err)
	defer repo.Close()

	svc := core.NewService(repo, "artifacts")
	exitOnErr(svc.Init(ctx))

	args := os.Args[1:]
	if len(args) == 0 {
		fatalf("usage: bridge <case|approval|report|device|session|version> ...")
	}

	switch args[0] {
	case "case":
		handleCase(ctx, svc, args[1:])
	case "approval":
		handleApproval(ctx, svc, args[1:])
	case "report":
		handleReport(ctx, svc, args[1:])
	case "device":
		handleDevice(ctx, svc, args[1:])
	case "session":
		handleSession(ctx, svc, args[1:])
	case "version":
		writeJSON(map[string]any{
			"name":       version.AppName,
			"version":    version.Version,
			"commit":     version.Commit,
			"build_date": version.BuildDate,
		})
	default:
		fatalf("unknown command %q", args[0])
	}
}

func handleCase(ctx context.Context, svc *core.Service, args []string) {
	if len(args) == 0 {
		fatalf("usage: bridge case <new|run|show|events>")
	}

	switch args[0] {
	case "new":
		fs := flag.NewFlagSet("case new", flag.ExitOnError)
		specPath := fs.String("spec", "", "path to case spec json")
		actor := fs.String("actor", "cli", "actor name")
		_ = fs.Parse(args[1:])
		if *specPath == "" {
			fatalf("--spec is required")
		}
		var spec domain.CaseSpec
		raw, err := os.ReadFile(*specPath)
		exitOnErr(err)
		exitOnErr(json.Unmarshal(raw, &spec))
		c, err := svc.CreateCase(ctx, spec, *actor)
		exitOnErr(err)
		writeJSON(c)
	case "run":
		fs := flag.NewFlagSet("case run", flag.ExitOnError)
		id := fs.String("id", "", "case id")
		actor := fs.String("actor", "cli", "actor name")
		_ = fs.Parse(args[1:])
		if *id == "" {
			fatalf("--id is required")
		}
		result, err := svc.RunCase(ctx, *id, *actor)
		exitOnErr(err)
		writeJSON(result)
	case "show":
		fs := flag.NewFlagSet("case show", flag.ExitOnError)
		id := fs.String("id", "", "case id")
		actor := fs.String("actor", "cli", "actor name")
		_ = fs.Parse(args[1:])
		if *id == "" {
			fatalf("--id is required")
		}
		c, err := svc.GetCase(ctx, *id, *actor)
		exitOnErr(err)
		writeJSON(c)
	case "events":
		fs := flag.NewFlagSet("case events", flag.ExitOnError)
		id := fs.String("id", "", "case id")
		_ = fs.Parse(args[1:])
		if *id == "" {
			fatalf("--id is required")
		}
		events, err := svc.ListEvents(ctx, *id)
		exitOnErr(err)
		writeJSON(events)
	default:
		fatalf("unknown case command %q", args[0])
	}
}

func handleApproval(ctx context.Context, svc *core.Service, args []string) {
	// NOTE: Approving/rejecting approvals requires the user to have
	// either "admin" or "approver" role in their JWT claims.
	// Users without these roles will receive a 403 Forbidden response.
	if len(args) == 0 {
		fatalf("usage: bridge approval <ls|approve|reject>")
	}

	switch args[0] {
	case "ls":
		fs := flag.NewFlagSet("approval ls", flag.ExitOnError)
		caseID := fs.String("case-id", "", "optional case id")
		_ = fs.Parse(args[1:])
		approvals, err := svc.ListApprovals(ctx, *caseID)
		exitOnErr(err)
		writeJSON(approvals)
	case "approve", "reject":
		fs := flag.NewFlagSet("approval "+args[0], flag.ExitOnError)
		id := fs.String("id", "", "approval id")
		actor := fs.String("actor", "cli", "actor name")
		reason := fs.String("reason", "", "optional reason")
		_ = fs.Parse(args[1:])
		if *id == "" {
			fatalf("--id is required")
		}
		approval, err := svc.ResolveApproval(ctx, *id, *actor, args[0], *reason)
		exitOnErr(err)
		writeJSON(approval)
	default:
		fatalf("unknown approval command %q", args[0])
	}
}

func handleReport(ctx context.Context, svc *core.Service, args []string) {
	if len(args) == 0 || args[0] != "build" {
		fatalf("usage: bridge report build --id <case-id>")
	}
	fs := flag.NewFlagSet("report build", flag.ExitOnError)
	id := fs.String("id", "", "case id")
	_ = fs.Parse(args[1:])
	if *id == "" {
		fatalf("--id is required")
	}
	report, err := svc.BuildReport(ctx, *id)
	exitOnErr(err)
	writeJSON(report)
}

func handleDevice(ctx context.Context, svc *core.Service, args []string) {
	if len(args) == 0 || args[0] != "ls" {
		fatalf("usage: bridge device ls")
	}
	devices, err := svc.ListDevices(ctx)
	exitOnErr(err)
	writeJSON(devices)
}

func handleSession(ctx context.Context, svc *core.Service, args []string) {
	if len(args) == 0 || args[0] != "ls" {
		fatalf("usage: bridge session ls")
	}
	sessions, err := svc.ListSessions(ctx)
	exitOnErr(err)
	writeJSON(sessions)
}

func dbPath() string {
	if env := os.Getenv("BRIDGEOS_DB"); env != "" {
		return env
	}
	if env := os.Getenv("HAL_PROXY_DB"); env != "" {
		return env
	}
	return "bridgeos.db"
}

func writeJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	exitOnErr(enc.Encode(v))
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func exitOnErr(err error) {
	if err == nil {
		return
	}
	enc := json.NewEncoder(os.Stderr)
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
	os.Exit(1)
}
