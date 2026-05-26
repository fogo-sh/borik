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

## T-010 Use activity-specific timeout policies

Main workflows currently apply one broad one-hour `StartToCloseTimeout` to all activities. That gives quick operations like workspace initialization, image download, split/join, cleanup, and Discord delivery far longer than they should need.

- Introduce helper functions for activity option classes: short admin/cleanup, bounded downloads, long CPU/video work, AI calls, and Discord delivery.
- Use narrower timeouts for workspace initialization, cleanup, image loading, split/join, and delivery.
- Keep longer timeouts only for operations that genuinely need them, such as video conversion and multi-step AI work.
- Revisit timeout values alongside retry and heartbeat policies so each activity has a coherent failure envelope.

## T-012 Use bounded contexts for workflow starts

Discord command handlers start Temporal workflows with `context.Background()`. If Temporal is unhealthy or unreachable, command handling can wait longer than is useful for an interactive Discord response.

- Wrap workflow starts with `context.WithTimeout`, using a short timeout appropriate for command handling.
- Surface a friendly user-facing error if the workflow could not be started in time.
- Apply this to image jobs, image generation, video-to-GIF, and APNG-to-GIF starts.
- Keep the timeout scoped to starting the workflow; do not tie workflow execution lifetime to the Discord request context.
