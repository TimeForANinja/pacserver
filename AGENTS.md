# AI Agent Guidelines

- Make only the changes needed for the request. Follow existing project conventions and preserve unrelated user work.
- Do not spend time on Git history or repository status, and do not stage, commit, push, switch branches, or otherwise modify Git state unless explicitly asked.
- For code changes, prefer focused unit tests for small, independent modules and deterministic logic. Avoid complex integration or end-to-end tests unless explicitly requested or essential to verify the change.
- At the end of Go code changes, run `go vet ./...` as a linter, `go test ./...` where relevant, and the build command documented in `README.md` (`go build -o pacserver ./cmd/pacserver.go`). Report checks that could not be run and why. For documentation-only changes, skip code checks.
- Do not run the PAC server or launch long-running processes, including `pacserver --serve`, unless explicitly asked.
- Keep changes small, and summarize what changed and which checks were run.

## README

- Keep `README.md` up to date whenever behavior, configuration, architecture, or developer workflow changes.
- The README should always include:
  - High-level architecture (what the server does, request/admin flow, project layout)
  - Config variables required and available to the user (`config.yaml` fields, defaults, and related zone/route/PAC inputs)
  - How to develop and how to run (build, test, `--serve` / `--test` / `--reload`, and any other operator steps)
- If a change makes the README incomplete or wrong, update it in the same work.

## Stuck technical problems

- If you cannot find a workable solution after a reasonable attempt, stop. Do not loop, thrash, or invent workarounds.
- Hand the user a single, clean prompt they can paste into a frontier model with internet access.
- That prompt should include: the goal, what you already tried, the exact error or blocker, relevant file/command context, and the question you need answered. Keep it self-contained and free of agent chatter.
