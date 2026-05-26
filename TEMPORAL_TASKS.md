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
