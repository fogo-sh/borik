# Temporal Follow-Up Tasks

## Add heartbeats for long-running activities

Long-running image/video/AI activities use large `StartToCloseTimeout` values but do not currently heartbeat. Heartbeats make cancellation responsive and help Temporal detect stuck work more precisely.

- Add `HeartbeatTimeout` values to activity options for long-running work.
- Call `activity.RecordHeartbeat` in long-running or multi-step activities where progress can be reported.
- Prioritize video conversion, frame processing, ImageMagick loops, and AI image operations.
- Make activity code observe context cancellation where practical, especially around loops and subprocesses.

## Use typed workflow starts

Bot code starts workflows by string name with anonymous argument structs. This bypasses useful Go SDK validation and makes workflow/client argument drift easier.

- Replace string workflow names with workflow function references.
- Replace anonymous argument structs with exported workflow arg types.
- Update `triggerJob`, `triggerGenerateImage`, `triggerGif`, and `triggerAPNGToGIF`.
- Keep workflow registration names stable unless an intentional compatibility plan is in place.

## Run cleanup and failure notification from disconnected workflow contexts

Cleanup and failure-notification activities currently use the normal workflow context. If the workflow is cancelled, those follow-up activities can be cancelled or skipped as well, leaving workspaces behind and preventing users from seeing the failure message.

- Use `workflow.NewDisconnectedContext` for cleanup/compensation paths that should run after cancellation.
- Apply this to `notifyFailure` and `cleanupWorkspace`, or introduce explicit `notifyFailureAfterCancellation` / `cleanupWorkspaceAfterCancellation` helpers.
- Keep normal workflow context for ordinary result delivery where cancellation should still stop the workflow.
- Add workflow tests for cancelled workflows to verify cleanup still runs.

## Make workspace cleanup an explicit workflow responsibility

`SendDiscordResult` currently deletes the workspace inside the Discord delivery activity. That makes the activity non-idempotent: if the Discord send fails after the artifact is read, a retry cannot send the result because the workspace has already been removed. It also returns workflow results containing workspace/artifact references that no longer exist.

- Remove `jobWorkspace.Cleanup()` from `SendDiscordResult`.
- Have workflows call `cleanupWorkspace` after successful delivery as an explicit final step.
- Prefer a deferred cleanup helper in each workflow so success, failure, and cancellation paths are handled consistently.
- Use a disconnected workflow context for cleanup that must happen even after cancellation.

## Move Discord typing pulses out of workflow history

The typing indicator loop currently schedules a Discord activity and workflow timer every five seconds while work is running. Long jobs can add hundreds of low-value events to workflow history before any real processing history is counted.

- Replace the workflow-side typing loop with a single cancellable long-running activity.
- Have the activity send Discord typing notifications on an interval until its context is cancelled.
- Add a `HeartbeatTimeout` and call `activity.RecordHeartbeat` from the typing pulse activity.
- Cancel the pulse from the workflow when processing completes or fails.

## Limit frame-processing fan-out

`ProcessImageWorkflow` currently schedules one activity per input frame immediately. Worker concurrency limits execution, but the workflow can still create a large pending activity set and history for long animated images.

- Add a workflow-side concurrency window for frame operations, such as processing four frames at a time.
- Preserve output ordering while collecting batch results.
- Consider moving batches of frames into a `ProcessFrameBatch` activity if workflow histories remain large.
- Add workflow tests that cover multi-frame inputs and confirm failures still trigger notification and cleanup.
