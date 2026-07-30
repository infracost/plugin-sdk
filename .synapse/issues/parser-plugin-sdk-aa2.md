---
id: parser-plugin-sdk-aa2
title: Ensure documentation matches actual implementation
status: open
type: feature
priority: 2
created_at: 2026-07-30T14:36:38Z
updated_at: 2026-07-30T14:36:38Z
closed_at: ~
close_reason: ~
external_ref: ~
---
This repo is a documentation source to define our plugin SDK that is used by the CLI (../cli). We have both Parsers (../parser) and Providers (../providers). There may be some other repos you need to review to understand the plumbing (../proto, ../config, etc.). Please review these docs, compare to the actual implementation, and ensure they're still accurate. Also make sure it's using the latest template if there's an example in one of the other repos