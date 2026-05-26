# Temporal Follow-Up Tasks

## T-001 Add heartbeats for long-running activities

Long-running image/video/AI activities use large `StartToCloseTimeout` values but do not currently heartbeat. Heartbeats make cancellation responsive and help Temporal detect stuck work more precisely.

- Add `HeartbeatTimeout` values to activity options for long-running work.
- Call `activity.RecordHeartbeat` in long-running or multi-step activities where progress can be reported.
- Prioritize video conversion, frame processing, ImageMagick loops, and AI image operations.
- Make activity code observe context cancellation where practical, especially around loops and subprocesses.

## T-002 Use typed workflow starts

Bot code starts workflows by string name with anonymous argument structs. This bypasses useful Go SDK validation and makes workflow/client argument drift easier.

- Replace string workflow names with workflow function references.
- Replace anonymous argument structs with exported workflow arg types.
- Update `triggerJob`, `triggerGenerateImage`, `triggerGif`, and `triggerAPNGToGIF`.
- Keep workflow registration names stable unless an intentional compatibility plan is in place.

## T-003 Run cleanup and failure notification from disconnected workflow contexts

Cleanup and failure-notification activities currently use the normal workflow context. If the workflow is cancelled, those follow-up activities can be cancelled or skipped as well, leaving workspaces behind and preventing users from seeing the failure message.

- Use `workflow.NewDisconnectedContext` for cleanup/compensation paths that should run after cancellation.
- Apply this to `notifyFailure` and `cleanupWorkspace`, or introduce explicit `notifyFailureAfterCancellation` / `cleanupWorkspaceAfterCancellation` helpers.
- Keep normal workflow context for ordinary result delivery where cancellation should still stop the workflow.
- Add workflow tests for cancelled workflows to verify cleanup still runs.

## T-004 Make workspace cleanup an explicit workflow responsibility

`SendDiscordResult` currently deletes the workspace inside the Discord delivery activity. That makes the activity non-idempotent: if the Discord send fails after the artifact is read, a retry cannot send the result because the workspace has already been removed. It also returns workflow results containing workspace/artifact references that no longer exist.

- Remove `jobWorkspace.Cleanup()` from `SendDiscordResult`.
- Have workflows call `cleanupWorkspace` after successful delivery as an explicit final step.
- Prefer a deferred cleanup helper in each workflow so success, failure, and cancellation paths are handled consistently.
- Use a disconnected workflow context for cleanup that must happen even after cancellation.

## T-005 Move Discord typing pulses out of workflow history

The typing indicator loop currently schedules a Discord activity and workflow timer every five seconds while work is running. Long jobs can add hundreds of low-value events to workflow history before any real processing history is counted.

- Replace the workflow-side typing loop with a single cancellable long-running activity.
- Have the activity send Discord typing notifications on an interval until its context is cancelled.
- Add a `HeartbeatTimeout` and call `activity.RecordHeartbeat` from the typing pulse activity.
- Cancel the pulse from the workflow when processing completes or fails.

## T-006 Limit frame-processing fan-out

`ProcessImageWorkflow` currently schedules one activity per input frame immediately. Worker concurrency limits execution, but the workflow can still create a large pending activity set and history for long animated images.

- Add a workflow-side concurrency window for frame operations, such as processing four frames at a time.
- Preserve output ordering while collecting batch results.
- Consider moving batches of frames into a `ProcessFrameBatch` activity if workflow histories remain large.
- Add workflow tests that cover multi-frame inputs and confirm failures still trigger notification and cleanup.

## T-007 Add a workflow versioning and deploy-safety plan

Workflow registrations currently use plain function registration, and there are no patch markers, `workflow.GetVersion` calls, worker versioning, or replay tests. Several planned changes alter activity scheduling and command ordering, which can cause nondeterminism for open workflow executions during deploys.

- Decide how to handle in-flight executions before shipping command-shape changes: drain/cancel existing jobs, add Temporal patch/version gates, or adopt Temporal worker versioning.
- Use `workflow.GetVersion` or patch markers around changes that alter workflow command ordering.
- Keep workflow and activity names stable across deploys unless there is a deliberate migration plan.
- Add replay or workflow tests for representative histories before making larger workflow changes.

## T-008 Register activities with explicit stable names

Most activities are registered under default Go function names, while job arguments pass matching string literals through the workflow. Function renames or package reshuffling can silently break new schedules and pending retries.

- Define constants for all workflow/activity API names.
- Register regular activities with `worker.RegisterActivityWithOptions`, matching the explicit-name pattern already used by Discord delivery activities.
- Update job argument `ActivityName` values to use the shared constants.
- Avoid changing registered names unless there is an intentional compatibility plan.

## T-009 Make image downloads context-aware

`LoadImage` accepts an activity context but currently downloads with `http.Get`, so cancellation does not abort the request promptly. Video downloads already use `http.NewRequestWithContext`; image downloads should follow the same pattern.

- Replace `http.Get` in `LoadImage` with `http.NewRequestWithContext` and `http.DefaultClient.Do`.
- Return an error for non-2xx image responses before persisting data.
- Consider sharing a small download helper between image and video activity code.
- Add tests for cancellation and HTTP status handling where practical.

## T-010 Use activity-specific timeout policies

Main workflows currently apply one broad one-hour `StartToCloseTimeout` to all activities. That gives quick operations like workspace initialization, image download, split/join, cleanup, and Discord delivery far longer than they should need.

- Introduce helper functions for activity option classes: short admin/cleanup, bounded downloads, long CPU/video work, AI calls, and Discord delivery.
- Use narrower timeouts for workspace initialization, cleanup, image loading, split/join, and delivery.
- Keep longer timeouts only for operations that genuinely need them, such as video conversion and multi-step AI work.
- Revisit timeout values alongside retry and heartbeat policies so each activity has a coherent failure envelope.

## T-011 Close workspace files after reads and writes

Workspace persistence opens artifact files without closing them after writes, and retrieval opens files without closing them after reads. Temporal workers are long-lived, so leaked file descriptors can accumulate across many activity executions.

- Add `defer f.Close()` in `Workspace.Persist` and `Workspace.Retrieve`.
- Check close errors after writes so failed flushes are surfaced.
- Consider replacing manual open/write/read code with `os.WriteFile` and `os.ReadFile` if that keeps the code simpler.
- Add a small workspace unit test for persist/retrieve behavior.

## T-012 Use bounded contexts for workflow starts

Discord command handlers start Temporal workflows with `context.Background()`. If Temporal is unhealthy or unreachable, command handling can wait longer than is useful for an interactive Discord response.

- Wrap workflow starts with `context.WithTimeout`, using a short timeout appropriate for command handling.
- Surface a friendly user-facing error if the workflow could not be started in time.
- Apply this to image jobs, image generation, video-to-GIF, and APNG-to-GIF starts.
- Keep the timeout scoped to starting the workflow; do not tie workflow execution lifetime to the Discord request context.

## T-013 Make workflow ID conflict and reuse behavior explicit

Workflow starts use Discord message or interaction IDs as workflow IDs, which is a good idempotency base, but the default conflict/reuse behavior is implicit.

- Set `WorkflowIDReusePolicy` and/or `WorkflowIDConflictPolicy` intentionally in `client.StartWorkflowOptions`.
- Decide whether duplicate Discord deliveries should be treated as success, ignored, or reported to the user.
- Decide whether completed workflow IDs may be reused.
- Handle already-started errors explicitly so duplicate command delivery does not produce confusing failures.
