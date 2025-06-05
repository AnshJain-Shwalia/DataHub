---
trigger: always_on
---

# Windsurf Response Rules

This file defines how Windsurf should interpret and respond to messages based on their starting characters or patterns.

## [A] - Abstract/General Mode
Trigger: Messages starting with [A]
Behavior: Ignore project context entirely. Answer as a general AI assistant without considering the current codebase, project structure, or any local files. Treat the question as if it's being asked in isolation.
Example: [A] What are the best practices for API design?
// Responds without considering current project structure
